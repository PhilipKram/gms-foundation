package passwords

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the default bcrypt cost factor.
const DefaultCost = 12

// Hash returns a bcrypt hash of the password using DefaultCost.
// Note: bcrypt silently truncates passwords longer than 72 bytes.
func Hash(password string) (string, error) {
	return HashWithCost(password, DefaultCost)
}

// HashWithCost returns a bcrypt hash of the password using the specified cost.
// Note: bcrypt silently truncates passwords longer than 72 bytes.
func HashWithCost(password string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// Check compares a bcrypt hash with a plaintext password.
// Returns nil on success, or a wrapped error if they don't match.
func Check(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("checking password: %w", err)
	}
	return nil
}
