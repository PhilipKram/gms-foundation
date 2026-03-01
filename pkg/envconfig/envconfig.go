package envconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Required returns the value of the environment variable named by key.
// It returns an error if the variable is not set. An explicitly set empty
// value is returned without error; use os.LookupEnv semantics.
func Required(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
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
// "true", "1", or "yes" (case-insensitive), false if set to "false", "0",
// or "no", and defaultValue for unset or unrecognized values.
func OptionalBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return defaultValue
	}
}

// OptionalInt returns the environment variable value parsed as an integer,
// or defaultValue if the variable is not set, empty, or not a valid integer.
func OptionalInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}

// OptionalDuration returns the environment variable value parsed as a
// time.Duration, or defaultValue if the variable is not set, empty, or not
// a valid duration string.
func OptionalDuration(key string, defaultValue time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultValue
	}
	return d
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
