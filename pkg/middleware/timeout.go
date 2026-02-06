package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout creates a middleware that enforces a request timeout.
// If a request takes longer than the specified duration, it will be cancelled
// and a 504 Gateway Timeout will be returned.
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			// Wrap the request with timeout context
			r = r.WithContext(ctx)

			// Execute the handler in a goroutine
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			// Wait for completion or timeout
			select {
			case <-done:
				// Request completed successfully
				return
			case <-ctx.Done():
				// Timeout occurred
				http.Error(w, "Request Timeout", http.StatusGatewayTimeout)
				return
			}
		})
	}
}
