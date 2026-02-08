package passwords

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHash_ValidHash(t *testing.T) {
	hash, err := Hash("mysecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	// Verify it's a valid bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("mysecret")); err != nil {
		t.Errorf("hash should be valid bcrypt: %v", err)
	}
}

func TestHash_UniqueSalts(t *testing.T) {
	h1, _ := Hash("samepassword")
	h2, _ := Hash("samepassword")
	if h1 == h2 {
		t.Error("two hashes of the same password should differ (unique salts)")
	}
}

func TestCheck_CorrectPassword(t *testing.T) {
	hash, _ := Hash("correct")
	if err := Check(hash, "correct"); err != nil {
		t.Errorf("expected no error for correct password, got: %v", err)
	}
}

func TestCheck_WrongPassword(t *testing.T) {
	hash, _ := Hash("correct")
	if err := Check(hash, "wrong"); err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestHashWithCost(t *testing.T) {
	hash, err := HashWithCost("test", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("reading cost: %v", err)
	}
	if cost != bcrypt.MinCost {
		t.Errorf("expected cost %d, got %d", bcrypt.MinCost, cost)
	}
}

func TestHashWithCost_InvalidCost(t *testing.T) {
	_, err := HashWithCost("test", 100) // cost too high
	if err == nil {
		t.Fatal("expected error for invalid cost")
	}
}

func TestHashWithCost_DefaultCost(t *testing.T) {
	hash, err := HashWithCost("test", DefaultCost)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("reading cost: %v", err)
	}
	if cost != DefaultCost {
		t.Errorf("expected cost %d, got %d", DefaultCost, cost)
	}
}

func TestCheck_InvalidHash(t *testing.T) {
	err := Check("not-a-valid-hash", "password")
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
}

func TestHash_EmptyPassword(t *testing.T) {
	hash, err := Hash("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := Check(hash, ""); err != nil {
		t.Errorf("expected empty password to match: %v", err)
	}
	if err := Check(hash, "notempty"); err == nil {
		t.Error("expected error when checking non-empty against empty hash")
	}
}
