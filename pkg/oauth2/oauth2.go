package oauth2

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// PKCE holds a code verifier and its S256 challenge.
type PKCE struct {
	Verifier  string
	Challenge string
}

// GenerateState creates a cryptographically random state string (base64url-encoded).
func GenerateState(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GeneratePKCE creates a PKCE verifier and S256 challenge pair.
func GeneratePKCE(byteLength int) (*PKCE, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)

	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	return &PKCE{
		Verifier:  verifier,
		Challenge: challenge,
	}, nil
}
