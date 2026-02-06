package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// RequestLogger creates a middleware that logs HTTP requests using zerolog.
// It logs the method, path, status code, duration, and bytes written.
func RequestLogger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log the request
			duration := time.Since(start)
			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status", wrapped.statusCode).
				Dur("duration_ms", duration).
				Int64("bytes", wrapped.written).
				Msg("HTTP request")
		})
	}
}

// RequestLoggerWithSkip creates a request logger that skips certain paths.
// This is useful for skipping health check or metrics endpoints that generate too much noise.
func RequestLoggerWithSkip(logger zerolog.Logger, skipPaths []string) func(http.Handler) http.Handler {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for certain paths
			if skipMap[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status", wrapped.statusCode).
				Dur("duration_ms", duration).
				Int64("bytes", wrapped.written).
				Msg("HTTP request")
		})
	}
}
