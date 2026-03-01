package healthcheck

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ginHealthCheckHandler responds with the health status of the application.
func ginHealthCheckHandler(check HealthCheckFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if check != nil {
			if err := check(); err != nil {
				c.String(http.StatusServiceUnavailable, err.Error())
				return
			}
		}
		c.Status(http.StatusOK)
	}
}

// Register sets up health check endpoints on the provided Gin router.
// Registers both /healthz/readiness and /healthz/liveness endpoints.
func Register(router *gin.Engine) {
	router.GET("/healthz/readiness", ginHealthCheckHandler(nil))
	router.GET("/healthz/liveness", ginHealthCheckHandler(nil))
}

// RegisterWithChecks sets up health check endpoints with custom health checks on a Gin router.
// The readinessCheck is called for /healthz/readiness, and livenessCheck for /healthz/liveness.
// If a check function is nil, the endpoint will always return 200 OK.
func RegisterWithChecks(router *gin.Engine, readinessCheck, livenessCheck HealthCheckFunc) {
	router.GET("/healthz/readiness", ginHealthCheckHandler(readinessCheck))
	router.GET("/healthz/liveness", ginHealthCheckHandler(livenessCheck))
}
