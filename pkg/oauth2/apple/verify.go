package apple

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	jwksCacheDuration     = 1 * time.Hour
	clockSkewToleranceSec = 60
)

var jwksCache struct {
	mu        sync.Mutex
	keys      map[string]*ecdsa.PublicKey
	expiresAt time.Time
}

func init() {
	jwksCache.keys = make(map[string]*ecdsa.PublicKey)
}

// VerifyIDToken verifies an Apple ID token's signature against Apple's JWKS endpoint
// and validates standard claims (exp, iat, aud, iss).
// Returns the token's claims map on success.
func VerifyIDToken(ctx context.Context, tokenString, audience string) (map[string]interface{}, error) {
	// Parse token to get kid
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing kid in token header")
	}

	// Get public key from JWKS
	publicKey, err := getApplePublicKey(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and verify token with signature validation
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	now := time.Now().Unix()

	// Check expiration: allow tokens up to clockSkewToleranceSec past their exp.
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp)+clockSkewToleranceSec < now {
			return nil, fmt.Errorf("token has expired")
		}
	} else {
		return nil, fmt.Errorf("missing exp claim")
	}

	// Check issued at: reject tokens issued more than clockSkewToleranceSec in the future.
	if iat, ok := claims["iat"].(float64); ok {
		if int64(iat) > now+clockSkewToleranceSec {
			return nil, fmt.Errorf("token issued in the future")
		}
	}

	// Validate audience
	if aud, ok := claims["aud"].(string); ok {
		if aud != audience {
			return nil, fmt.Errorf("invalid audience")
		}
	} else {
		return nil, fmt.Errorf("missing aud claim")
	}

	// Validate issuer
	if iss, ok := claims["iss"].(string); ok {
		if iss != "https://appleid.apple.com" {
			return nil, fmt.Errorf("invalid issuer")
		}
	} else {
		return nil, fmt.Errorf("missing iss claim")
	}

	return claims, nil
}

// getApplePublicKey fetches and caches Apple's public keys from JWKS endpoint.
func getApplePublicKey(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
	jwksCache.mu.Lock()
	defer jwksCache.mu.Unlock()

	cacheValid := time.Now().Before(jwksCache.expiresAt)
	if cacheValid {
		if key, ok := jwksCache.keys[kid]; ok {
			return key, nil
		}
		return nil, fmt.Errorf("key %s not found in JWKS", kid)
	}

	// Cache expired — fetch new keys
	req, err := http.NewRequestWithContext(ctx, "GET", "https://appleid.apple.com/auth/keys", nil)
	if err != nil {
		return nil, fmt.Errorf("creating JWKS request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close on HTTP response

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Use string `json:"use"`
			Alg string `json:"alg"`
			X   string `json:"x,omitempty"`
			Y   string `json:"y,omitempty"`
			Crv string `json:"crv,omitempty"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	newKeys := make(map[string]*ecdsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty != "EC" || k.Use != "sig" {
			continue
		}

		x, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil {
			continue
		}
		y, err := base64.RawURLEncoding.DecodeString(k.Y)
		if err != nil {
			continue
		}

		pubKey := &ecdsa.PublicKey{
			X: new(big.Int).SetBytes(x),
			Y: new(big.Int).SetBytes(y),
		}

		switch k.Crv {
		case "P-256":
			pubKey.Curve = elliptic.P256()
		case "P-384":
			pubKey.Curve = elliptic.P384()
		case "P-521":
			pubKey.Curve = elliptic.P521()
		default:
			continue
		}

		newKeys[k.Kid] = pubKey
	}

	jwksCache.keys = newKeys
	jwksCache.expiresAt = time.Now().Add(jwksCacheDuration)

	if key, ok := newKeys[kid]; ok {
		return key, nil
	}

	return nil, fmt.Errorf("key %s not found in JWKS", kid)
}
