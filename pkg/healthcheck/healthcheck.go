package healthcheck

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
)

// HealthCheckFunc is a function that performs a health check.
// It should return an error if the check fails, or nil if healthy.
type HealthCheckFunc func() error

// healthCheckHandler responds with the health status of the application.
func healthCheckHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Register sets up health check endpoints on the provided Gin router.
// Registers both /healthz/readiness and /healthz/liveness endpoints.
func Register(router *gin.Engine) {
	router.GET("/healthz/readiness", healthCheckHandler)
	router.GET("/healthz/liveness", healthCheckHandler)
}

// RegisterChi sets up health check endpoints on the provided Chi router.
// Registers both /healthz/readiness and /healthz/liveness endpoints.
func RegisterChi(router chi.Router) {
	router.Get("/healthz/readiness", chiHealthCheckHandler(nil))
	router.Get("/healthz/liveness", chiHealthCheckHandler(nil))
}

// RegisterChiWithChecks sets up health check endpoints with custom health checks.
// The readinessCheck is called for /healthz/readiness, and livenessCheck for /healthz/liveness.
// If a check function is nil, the endpoint will always return 200 OK.
func RegisterChiWithChecks(router chi.Router, readinessCheck, livenessCheck HealthCheckFunc) {
	router.Get("/healthz/readiness", chiHealthCheckHandler(readinessCheck))
	router.Get("/healthz/liveness", chiHealthCheckHandler(livenessCheck))
}

// chiHealthCheckHandler creates a http.HandlerFunc for chi router with optional health check.
func chiHealthCheckHandler(check HealthCheckFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			if err := check(); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
