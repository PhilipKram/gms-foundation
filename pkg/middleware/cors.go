package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig defines CORS configuration options.
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins (e.g., ["http://localhost:3000"]).
	// Use ["*"] to allow all origins (not recommended for production).
	AllowedOrigins []string
	// AllowedMethods is a list of allowed HTTP methods (e.g., ["GET", "POST"]).
	AllowedMethods []string
	// AllowedHeaders is a list of allowed headers.
	AllowedHeaders []string
	// ExposedHeaders is a list of headers that are safe to expose to the client.
	ExposedHeaders []string
	// AllowCredentials indicates whether credentials (cookies, auth headers) are allowed.
	AllowCredentials bool
	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached.
	MaxAge int
}

// DefaultCORSConfig returns a CORS configuration with sensible defaults.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           300,
	}
}

// CORS creates a middleware that handles Cross-Origin Resource Sharing (CORS).
func CORS(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowOrigin := isOriginAllowed(origin, config.AllowedOrigins)

			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				if len(config.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				}

				if len(config.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				}

				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				}

				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if the origin is in the allowed list.
func isOriginAllowed(origin string, allowedOrigins []string) string {
	if len(allowedOrigins) == 0 {
		return ""
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return "*"
		}
		if allowed == origin {
			return origin
		}
	}

	return ""
}
