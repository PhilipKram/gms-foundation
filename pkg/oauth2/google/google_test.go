package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExchangeCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
				t.Errorf("expected content-type application/x-www-form-urlencoded, got %s", ct)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			if r.FormValue("code") != "test-code" {
				t.Errorf("code = %q, want %q", r.FormValue("code"), "test-code")
			}
			if r.FormValue("code_verifier") != "test-verifier" {
				t.Errorf("code_verifier = %q, want %q", r.FormValue("code_verifier"), "test-verifier")
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(TokenResponse{
				AccessToken: "access-token-123",
				IDToken:     "id-token-456",
				TokenType:   "Bearer",
			})
		}))
		defer server.Close()

		// Override the token endpoint for testing
		origClient := httpClient
		httpClient = server.Client()
		defer func() { httpClient = origClient }()

		// We need to make the function hit our test server instead of Google.
		// Since the URL is hardcoded, we'll test the HTTP interaction via a transport override.
		transport := &rewriteTransport{
			base:    http.DefaultTransport,
			rewrite: server.URL,
		}
		httpClient = &http.Client{Transport: transport}

		resp, err := ExchangeCode(context.Background(), "test-code", "client-id", "client-secret", "http://redirect", "test-verifier")
		if err != nil {
			t.Fatalf("ExchangeCode() error: %v", err)
		}
		if resp.AccessToken != "access-token-123" {
			t.Errorf("AccessToken = %q, want %q", resp.AccessToken, "access-token-123")
		}
		if resp.IDToken != "id-token-456" {
			t.Errorf("IDToken = %q, want %q", resp.IDToken, "id-token-456")
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

func TestGetUserInfo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
				t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token")
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(UserInfo{
				Sub:           "google-sub-123",
				Email:         "user@example.com",
				EmailVerified: true,
				Name:          "Test User",
				Picture:       "https://example.com/photo.jpg",
			})
		}))
		defer server.Close()

		transport := &rewriteTransport{base: http.DefaultTransport, rewrite: server.URL}
		origClient := httpClient
		httpClient = &http.Client{Transport: transport}
		defer func() { httpClient = origClient }()

		info, err := GetUserInfo(context.Background(), "test-token")
		if err != nil {
			t.Fatalf("GetUserInfo() error: %v", err)
		}
		if info.Email != "user@example.com" {
			t.Errorf("Email = %q, want %q", info.Email, "user@example.com")
		}
		if !info.EmailVerified {
			t.Error("EmailVerified should be true")
		}
		if info.Sub != "google-sub-123" {
			t.Errorf("Sub = %q, want %q", info.Sub, "google-sub-123")
		}
	})

	t.Run("error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		transport := &rewriteTransport{base: http.DefaultTransport, rewrite: server.URL}
		origClient := httpClient
		httpClient = &http.Client{Transport: transport}
		defer func() { httpClient = origClient }()

		_, err := GetUserInfo(context.Background(), "bad-token")
		if err == nil {
			t.Fatal("expected error for bad status, got nil")
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
