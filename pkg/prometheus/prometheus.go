package prometheus

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Register(router *gin.Engine) {

	// Expose the registered Prometheus metrics via HTTP
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

}
