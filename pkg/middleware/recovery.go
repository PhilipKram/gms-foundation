package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
)

// Recovery creates a middleware that recovers from panics and logs them.
// It prevents the server from crashing and returns a 500 Internal Server Error.
func Recovery(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					logger.Error().
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Str("panic", fmt.Sprintf("%v", err)).
						Str("stack", string(debug.Stack())).
						Msg("Panic recovered")

					// Return 500 Internal Server Error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
