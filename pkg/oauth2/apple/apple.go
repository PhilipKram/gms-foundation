package apple

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// TokenResponse holds the tokens returned by Apple's token endpoint.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
}

// ExchangeCode exchanges an authorization code for tokens at Apple's token endpoint.
// Supports PKCE via the codeVerifier parameter.
func ExchangeCode(ctx context.Context, code, clientID, clientSecret, redirectURI, codeVerifier string) (*TokenResponse, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
		"code_verifier": {codeVerifier},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://appleid.apple.com/auth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("token exchange failed")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed")
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close on HTTP response

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed")
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("token exchange failed")
	}
	return &token, nil
}

// GenerateClientSecret creates a signed ES256 JWT for Apple's token endpoint.
// Parameters: teamID, clientID, keyID, privateKeyPEM (PKCS8 PEM-encoded P-256 key).
// The JWT is valid for 5 minutes.
func GenerateClientSecret(teamID, clientID, keyID, privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parsing private key: %w", err)
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("key is not ECDSA")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": teamID,
		"sub": clientID,
		"aud": "https://appleid.apple.com",
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID

	return token.SignedString(ecKey)
}
