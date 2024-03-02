package healthcheck

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// healthCheckHandler responds with the health status of the application.
func healthCheckHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Register sets up health check endpoints on the provided router.
func Register(router *gin.Engine) {
	router.GET("/healthz/readiness", healthCheckHandler)
	router.GET("/healthz/liveness", healthCheckHandler)
}
