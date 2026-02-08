package envconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Required returns the value of the environment variable named by key.
// It returns an error if the variable is not set or is empty.
func Required(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("%s environment variable is required", key)
	}
	return v, nil
}

// Optional returns the value of the environment variable named by key,
// or defaultValue if the variable is not set or is empty.
func Optional(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// OptionalBool returns true if the environment variable is set to
// "true", "1", or "yes" (case-insensitive). Returns defaultValue otherwise.
func OptionalBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

// OptionalStringSlice splits the environment variable value by separator,
// trims whitespace from each element, and removes empty strings.
// Returns defaultValue if the variable is not set or is empty.
func OptionalStringSlice(key, separator string, defaultValue []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}

	parts := strings.Split(v, separator)
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}

// ResolveAbsPath resolves the given path to an absolute path. If the path is
// already absolute, it is returned as-is. Otherwise it is resolved relative to
// the current working directory.
func ResolveAbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path for %q: %w", path, err)
	}
	return abs, nil
}
