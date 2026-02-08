# GMS Foundation

A collection of production-ready Go packages for building robust HTTP services with structured logging, health checks, metrics, and common middleware patterns.

## Features

- **Structured Logging** - Flexible zerolog-based logging with Logstash support
- **HTTP Server Utilities** - Setup helpers for both Gin and Chi routers
- **Health Checks** - Kubernetes-compatible health check endpoints
- **Prometheus Metrics** - Easy metrics endpoint registration
- **Middleware** - Common HTTP middleware (CORS, logging, recovery, timeout)

## Installation

```bash
go get github.com/PhilipKram/gms-foundation
```

## Quick Start

### Structured Logging

```go
import (
    "github.com/PhilipKram/gms-foundation/pkg/logger"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
    // Setup global logger
    logger.SetupLogger(logger.ConfigSchema{
        Level:    int8(zerolog.InfoLevel),
        Logstash: false, // Set to true for JSON output
    })

    log.Info().Msg("Server starting")
    log.Error().Err(err).Msg("Something went wrong")
}
```

### Chi Router Setup

```go
import (
    "github.com/PhilipKram/gms-foundation/pkg/server/chi"
    "github.com/PhilipKram/gms-foundation/pkg/healthcheck"
    "github.com/PhilipKram/gms-foundation/pkg/prometheus"
)

func main() {
    // Create server and router
    srv, router := chi.Setup(chi.ConfigSchema{
        Port:      "8080",
        AccessLog: true,
    })

    // Register health checks
    healthcheck.RegisterChi(router)

    // Register Prometheus metrics
    prometheus.RegisterChi(router)

    // Add your routes
    router.Get("/api/hello", yourHandler)

    // Start server with graceful shutdown
    chi.Start(srv)
}
```

### Custom Health Checks

```go
import "github.com/PhilipKram/gms-foundation/pkg/healthcheck"

// Register health checks with custom logic
healthcheck.RegisterChiWithChecks(router,
    func() error {
        // Readiness check - verify database connection
        return database.Ping()
    },
    func() error {
        // Liveness check - verify critical services
        return checkCriticalServices()
    },
)
```

### Middleware

```go
import (
    "github.com/PhilipKram/gms-foundation/pkg/middleware"
    "github.com/rs/zerolog/log"
)

// Request logging with path exclusions
router.Use(middleware.RequestLoggerWithSkip(
    log.Logger,
    []string{"/healthz/liveness", "/healthz/readiness", "/metrics"},
))

// Panic recovery
router.Use(middleware.Recovery(log.Logger))

// CORS
router.Use(middleware.CORS(middleware.CORSConfig{
    AllowedOrigins:   []string{"http://localhost:3000"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
    AllowCredentials: true,
    MaxAge:           300,
}))

// Request timeout
router.Use(middleware.Timeout(60 * time.Second))
```

## Package Overview

### `pkg/logger`
Provides structured logging using zerolog with support for:
- Console and JSON (Logstash) output formats
- Configurable log levels
- Custom output writers
- Automatic timestamp and caller information

### `pkg/server/chi`
Chi router utilities:
- Server setup with sensible defaults
- Graceful shutdown handling
- Configurable timeouts

### `pkg/server` (Gin)
Gin router utilities with similar features to the Chi package.

### `pkg/healthcheck`
Kubernetes-compatible health check endpoints:
- `/healthz/readiness` - Readiness probe
- `/healthz/liveness` - Liveness probe
- Support for custom health check functions

### `pkg/prometheus`
Prometheus metrics integration:
- Automatic `/metrics` endpoint registration
- Works with both Gin and Chi routers

### `pkg/middleware`
Standard HTTP middleware:
- **RequestLogger** - Structured request logging
- **Recovery** - Panic recovery with stack traces
- **CORS** - Cross-Origin Resource Sharing
- **Timeout** - Request timeout enforcement

## Advanced Usage

### Creating Independent Logger Instances

```go
// Instead of modifying the global logger
customLogger := logger.New(logger.ConfigSchema{
    Level:            int8(zerolog.DebugLevel),
    Logstash:         true,
    Writer:           customWriter,
    DisableCaller:    false,
    DisableTimestamp: false,
})
```

### Custom CORS Configuration

```go
corsMiddleware := middleware.CORS(middleware.CORSConfig{
    AllowedOrigins:   []string{"https://example.com"},
    AllowedMethods:   []string{"GET", "POST"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    ExposedHeaders:   []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           600,
})
```

## Router Compatibility

This library supports both popular Go routers:

- **Chi** (`github.com/go-chi/chi/v5`) - Use `pkg/server/chi`, `healthcheck.RegisterChi()`, `prometheus.RegisterChi()`
- **Gin** (`github.com/gin-gonic/gin`) - Use `pkg/server`, `healthcheck.Register()`, `prometheus.Register()`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See [LICENSE](LICENSE) file for details.

