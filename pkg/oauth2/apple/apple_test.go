package apple

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateClientSecret(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("failed to marshal key: %v", err)
	}

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	secret, err := GenerateClientSecret("TEAM123", "com.example.app", "KEY456", string(pemBlock))
	if err != nil {
		t.Fatalf("GenerateClientSecret() error: %v", err)
	}

	if secret == "" {
		t.Fatal("generated secret is empty")
	}

	// Parse and verify the generated JWT
	token, err := jwt.Parse(secret, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			t.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse generated JWT: %v", err)
	}
	if !token.Valid {
		t.Fatal("generated JWT is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to get claims from JWT")
	}

	if iss, _ := claims["iss"].(string); iss != "TEAM123" {
		t.Errorf("iss = %q, want %q", iss, "TEAM123")
	}
	if sub, _ := claims["sub"].(string); sub != "com.example.app" {
		t.Errorf("sub = %q, want %q", sub, "com.example.app")
	}
	if aud, _ := claims["aud"].(string); aud != "https://appleid.apple.com" {
		t.Errorf("aud = %q, want %q", aud, "https://appleid.apple.com")
	}

	if kid, _ := token.Header["kid"].(string); kid != "KEY456" {
		t.Errorf("kid = %q, want %q", kid, "KEY456")
	}
}

func TestGenerateClientSecret_InvalidPEM(t *testing.T) {
	_, err := GenerateClientSecret("TEAM123", "com.example.app", "KEY456", "not-a-pem")
	if err == nil {
		t.Fatal("expected error for invalid PEM, got nil")
	}
}

func TestExchangeCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			if r.FormValue("code") != "test-code" {
				t.Errorf("code = %q, want %q", r.FormValue("code"), "test-code")
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(TokenResponse{
				AccessToken: "apple-access-token",
				IDToken:     "apple-id-token",
				TokenType:   "Bearer",
			})
		}))
		defer server.Close()

		transport := &rewriteTransport{base: http.DefaultTransport, rewrite: server.URL}
		origClient := httpClient
		httpClient = &http.Client{Transport: transport}
		defer func() { httpClient = origClient }()

		resp, err := ExchangeCode(context.Background(), "test-code", "client-id", "client-secret", "http://redirect", "verifier")
		if err != nil {
			t.Fatalf("ExchangeCode() error: %v", err)
		}
		if resp.AccessToken != "apple-access-token" {
			t.Errorf("AccessToken = %q, want %q", resp.AccessToken, "apple-access-token")
		}
		if resp.IDToken != "apple-id-token" {
			t.Errorf("IDToken = %q, want %q", resp.IDToken, "apple-id-token")
		}
	})

	t.Run("error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		transport := &rewriteTransport{base: http.DefaultTransport, rewrite: server.URL}
		origClient := httpClient
		httpClient = &http.Client{Transport: transport}
		defer func() { httpClient = origClient }()

		_, err := ExchangeCode(context.Background(), "bad-code", "client-id", "client-secret", "http://redirect", "verifier")
		if err == nil {
			t.Fatal("expected error for bad status, got nil")
		}
	})
}

func TestVerifyIDToken(t *testing.T) {
	// Generate a test key pair for signing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	// Set up a mock JWKS server
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x := privateKey.X.Bytes()
		y := privateKey.Y.Bytes()

		// Pad to 32 bytes for P-256
		for len(x) < 32 {
			x = append([]byte{0}, x...)
		}
		for len(y) < 32 {
			y = append([]byte{0}, y...)
		}

		jwks := map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"kty": "EC",
					"kid": "test-kid",
					"use": "sig",
					"alg": "ES256",
					"crv": "P-256",
					"x":   base64.RawURLEncoding.EncodeToString(x),
					"y":   base64.RawURLEncoding.EncodeToString(y),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	// Redirect httpClient to our test server for JWKS fetches
	transport := &rewriteTransport{base: http.DefaultTransport, rewrite: jwksServer.URL}
	origClient := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = origClient }()

	// Helper to create a signed test token
	createToken := func(claims jwt.MapClaims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		token.Header["kid"] = "test-kid"
		signed, err := token.SignedString(privateKey)
		if err != nil {
			t.Fatalf("failed to sign test token: %v", err)
		}
		return signed
	}

	t.Run("valid token", func(t *testing.T) {
		// Reset cache to force JWKS fetch
		jwksCache.mu.Lock()
		jwksCache.expiresAt = time.Time{}
		jwksCache.mu.Unlock()

		now := time.Now()
		tokenStr := createToken(jwt.MapClaims{
			"iss": "https://appleid.apple.com",
			"aud": "com.example.app",
			"sub": "apple-user-123",
			"exp": float64(now.Add(5 * time.Minute).Unix()),
			"iat": float64(now.Unix()),
		})

		claims, err := VerifyIDToken(context.Background(), tokenStr, "com.example.app")
		if err != nil {
			t.Fatalf("VerifyIDToken() error: %v", err)
		}

		if sub, _ := claims["sub"].(string); sub != "apple-user-123" {
			t.Errorf("sub = %q, want %q", sub, "apple-user-123")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		jwksCache.mu.Lock()
		jwksCache.expiresAt = time.Time{}
		jwksCache.mu.Unlock()

		past := time.Now().Add(-2 * time.Hour)
		tokenStr := createToken(jwt.MapClaims{
			"iss": "https://appleid.apple.com",
			"aud": "com.example.app",
			"sub": "apple-user-123",
			"exp": float64(past.Unix()),
			"iat": float64(past.Add(-5 * time.Minute).Unix()),
		})

		_, err := VerifyIDToken(context.Background(), tokenStr, "com.example.app")
		if err == nil {
			t.Fatal("expected error for expired token, got nil")
		}
	})

	t.Run("invalid audience", func(t *testing.T) {
		jwksCache.mu.Lock()
		jwksCache.expiresAt = time.Time{}
		jwksCache.mu.Unlock()

		now := time.Now()
		tokenStr := createToken(jwt.MapClaims{
			"iss": "https://appleid.apple.com",
			"aud": "com.wrong.app",
			"sub": "apple-user-123",
			"exp": float64(now.Add(5 * time.Minute).Unix()),
			"iat": float64(now.Unix()),
		})

		_, err := VerifyIDToken(context.Background(), tokenStr, "com.example.app")
		if err == nil {
			t.Fatal("expected error for invalid audience, got nil")
		}
	})

	t.Run("JWKS cache", func(t *testing.T) {
		// First call should populate cache
		jwksCache.mu.Lock()
		jwksCache.expiresAt = time.Time{}
		jwksCache.mu.Unlock()

		now := time.Now()
		tokenStr := createToken(jwt.MapClaims{
			"iss": "https://appleid.apple.com",
			"aud": "com.example.app",
			"sub": "apple-user-123",
			"exp": float64(now.Add(5 * time.Minute).Unix()),
			"iat": float64(now.Unix()),
		})

		_, err := VerifyIDToken(context.Background(), tokenStr, "com.example.app")
		if err != nil {
			t.Fatalf("first call error: %v", err)
		}

		// Cache should now be populated
		jwksCache.mu.Lock()
		cacheValid := time.Now().Before(jwksCache.expiresAt)
		hasKey := jwksCache.keys["test-kid"] != nil
		jwksCache.mu.Unlock()

		if !cacheValid {
			t.Error("cache should be valid after first call")
		}
		if !hasKey {
			t.Error("cache should contain test-kid")
		}

		// Second call should use cache (no JWKS fetch needed)
		_, err = VerifyIDToken(context.Background(), tokenStr, "com.example.app")
		if err != nil {
			t.Fatalf("second call (cached) error: %v", err)
		}
	})
}

// rewriteTransport redirects all requests to the test server URL.
type rewriteTransport struct {
	base    http.RoundTripper
	rewrite string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.rewrite[len("http://"):]
	return t.base.RoundTrip(req)
}
