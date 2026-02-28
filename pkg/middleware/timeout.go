package middleware

import (
	"net/http"
	"time"
)

// Timeout creates a middleware that enforces a request timeout.
// If a request takes longer than the specified duration, it will be cancelled
// and a 504 Gateway Timeout will be returned.
//
// This uses http.TimeoutHandler which safely serialises writes to the
// ResponseWriter, preventing the race condition that occurs when a handler
// goroutine and a timeout path both write concurrently.
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "Request Timeout")
	}
}
