package prometheus

import (
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Register exposes Prometheus metrics endpoint on the provided Gin router.
// The metrics will be available at GET /metrics.
func Register(router *gin.Engine) {
	// Expose the registered Prometheus metrics via HTTP
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// RegisterChi exposes Prometheus metrics endpoint on the provided Chi router.
// The metrics will be available at GET /metrics.
func RegisterChi(router chi.Router) {
	router.Get("/metrics", promhttp.Handler().ServeHTTP)
}
