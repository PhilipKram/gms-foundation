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

// Flush delegates to the underlying ResponseWriter if it implements http.Flusher.
// This is required for SSE/streaming responses to work through the logging middleware.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap returns the underlying ResponseWriter, allowing http.ResponseController
// to access interfaces (like http.Flusher) on the original writer.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// logEvent returns a zerolog event at the appropriate level for the given HTTP status code.
// >= 500 -> Error, >= 400 -> Warn, else -> Info.
func logEvent(logger zerolog.Logger, status int) *zerolog.Event {
	switch {
	case status >= 500:
		return logger.Error()
	case status >= 400:
		return logger.Warn()
	default:
		return logger.Info()
	}
}

// RequestLogger creates a middleware that logs HTTP requests using zerolog.
// It logs the method, path, status code, duration, and bytes written.
// Log level is determined by status code: 5xx=Error, 4xx=Warn, else=Info.
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
			logEvent(logger, wrapped.statusCode).
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
			logEvent(logger, wrapped.statusCode).
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
