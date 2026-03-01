package prometheus

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Register exposes Prometheus metrics endpoint on the provided Gin router.
// The metrics will be available at GET /metrics.
func Register(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
