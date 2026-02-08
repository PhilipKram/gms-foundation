package oauth2

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGenerateState(t *testing.T) {
	state, err := GenerateState(32)
	if err != nil {
		t.Fatalf("GenerateState() error: %v", err)
	}

	if state == "" {
		t.Fatal("GenerateState() returned empty string")
	}

	// Should be base64url encoded 32 bytes
	decoded, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		t.Fatalf("state is not valid base64url: %v", err)
	}
	if len(decoded) != 32 {
		t.Errorf("decoded state length = %d, want %d", len(decoded), 32)
	}

	// Two calls should produce different states
	state2, _ := GenerateState(32)
	if state == state2 {
		t.Error("two GenerateState() calls returned same value")
	}
}

func TestGeneratePKCE(t *testing.T) {
	pkce, err := GeneratePKCE(32)
	if err != nil {
		t.Fatalf("GeneratePKCE() error: %v", err)
	}

	if pkce.Verifier == "" {
		t.Fatal("PKCE verifier is empty")
	}
	if pkce.Challenge == "" {
		t.Fatal("PKCE challenge is empty")
	}

	// Verify S256: challenge = base64url(sha256(verifier))
	h := sha256.Sum256([]byte(pkce.Verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(h[:])
	if pkce.Challenge != expectedChallenge {
		t.Errorf("PKCE challenge mismatch: got %q, want %q", pkce.Challenge, expectedChallenge)
	}

	// Two calls should produce different values
	pkce2, _ := GeneratePKCE(32)
	if pkce.Verifier == pkce2.Verifier {
		t.Error("two GeneratePKCE() calls returned same verifier")
	}
}
