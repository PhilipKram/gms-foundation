package prometheus

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterChi exposes Prometheus metrics endpoint on the provided Chi router.
// The metrics will be available at GET /metrics.
func RegisterChi(router chi.Router) {
	router.Get("/metrics", promhttp.Handler().ServeHTTP)
}
