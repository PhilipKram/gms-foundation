# GMS Foundation

A collection of production-ready Go packages for building robust HTTP services. Provides structured logging, health checks, metrics, middleware, and common application utilities (JSON helpers, password hashing, URL canonicalization, database pooling, file uploads, environment config).

Requires **Go 1.23+**.

## Features

**Infrastructure**
- **Structured Logging** - Flexible zerolog-based logging with Logstash support
- **HTTP Server Utilities** - Setup helpers for both Gin and Chi routers
- **Health Checks** - Kubernetes-compatible health check endpoints
- **Prometheus Metrics** - Easy metrics endpoint registration
- **Middleware** - Common HTTP middleware (CORS, logging, recovery, timeout)

**Application Utilities**
- **JSON Response Helpers** - `WriteJSON`, `WriteError`, `WritePaginated` for consistent API responses
- **Password Hashing** - Bcrypt wrapper with configurable cost
- **URL Canonicalization** - Normalize URLs, strip tracking params, compute dedup hashes
- **Database Connection Pool** - `sql.DB` setup with functional options and sensible defaults
- **MongoDB Client** - MongoDB connection wrapper with CSFLE auto-encryption support and functional options
- **Redis Client** - Redis connection wrapper with standalone and Sentinel failover support
- **File Upload Storage** - Configurable categories, MIME validation, size limits, path traversal protection
- **Environment Config** - Typed helpers for loading required/optional env vars

## Installation

```bash
go get github.com/PhilipKram/gms-foundation
```

## Quick Start

### Chi Router Setup

```go
import (
    "github.com/PhilipKram/gms-foundation/pkg/server/chi"
    "github.com/PhilipKram/gms-foundation/pkg/healthcheck"
    "github.com/PhilipKram/gms-foundation/pkg/prometheus"
)

func main() {
    srv, router := chi.Setup(chi.Config{
        Port:      "8080",
        AccessLog: true,
    })

    healthcheck.RegisterChi(router)
    prometheus.RegisterChi(router)

    router.Get("/api/hello", yourHandler)

    chi.Start(srv)
}
```

### Structured Logging

```go
import (
    "github.com/PhilipKram/gms-foundation/pkg/logger"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
    logger.SetupLogger(logger.Config{
        Level:    int8(zerolog.InfoLevel),
        Logstash: false,
    })

    log.Info().Msg("Server starting")
}
```

### JSON Response Helpers

```go
import "github.com/PhilipKram/gms-foundation/pkg/httputil"

func listHandler(w http.ResponseWriter, r *http.Request) {
    items, total := fetchItems()
    httputil.WritePaginated(w, http.StatusOK, items, total, 20, 0)
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
    httputil.WriteError(w, http.StatusNotFound, "resource not found")
}
```

### Password Hashing

```go
import "github.com/PhilipKram/gms-foundation/pkg/passwords"

hash, err := passwords.Hash("user-password")
err = passwords.Check(hash, "user-password") // nil on match
```

### URL Canonicalization

```go
import "github.com/PhilipKram/gms-foundation/pkg/canonical"

url, hash, err := canonical.Canonicalize("HTTPS://Example.COM/page?utm_source=twitter&q=test")
// url:  "https://example.com/page?q=test"
// hash: deterministic SHA-256 for deduplication
```

### Database Connection Pool

```go
import "github.com/PhilipKram/gms-foundation/pkg/dbutil"

db, err := dbutil.OpenMySQL(dsn,
    dbutil.WithMaxOpenConns(50),
    dbutil.WithConnMaxLifetime(10 * time.Minute),
)
```

### MongoDB Client

```go
import "github.com/PhilipKram/gms-foundation/pkg/mongodb"

client, err := mongodb.Connect(ctx, mongodb.Config{
    Host:     "mongodb://localhost:27017",
    Database: "myapp",
    Auth:     mongodb.AuthConfig{Username: "user", Password: "pass"},
}, mongodb.WithAppName("my-service"))
if err != nil {
    log.Fatal(err)
}
defer client.Close(ctx)

db := client.DB()
```

### Redis Client

```go
import "github.com/PhilipKram/gms-foundation/pkg/redis"

client, err := redis.Connect(ctx, redis.Config{
    Addr: "localhost:6379",
    Auth: redis.AuthConfig{Password: "secret"},
}, redis.WithPoolSize(20))
if err != nil {
    log.Fatal(err)
}
defer client.Close()

client.Unwrap().Set(ctx, "key", "value", 0)
```

### File Upload Storage

```go
import "github.com/PhilipKram/gms-foundation/pkg/uploads"

storage, err := uploads.NewStorage("/var/uploads")
relPath, err := storage.SaveFile(file, "image/jpeg") // "images/uuid.jpg"
err = storage.DeleteFile(relPath)
```

### Environment Config

```go
import "github.com/PhilipKram/gms-foundation/pkg/envconfig"

dbDSN, err := envconfig.Required("DB_DSN")
port := envconfig.Optional("PORT", "8080")
secure := envconfig.OptionalBool("SECURE_COOKIES", false)
origins := envconfig.OptionalStringSlice("CORS_ORIGINS", ",", []string{"http://localhost:3000"})
```

### Middleware

```go
import (
    "github.com/PhilipKram/gms-foundation/pkg/middleware"
    "github.com/rs/zerolog/log"
)

router.Use(middleware.RequestLoggerWithSkip(
    log.Logger,
    []string{"/healthz/liveness", "/healthz/readiness", "/metrics"},
))
router.Use(middleware.Recovery(log.Logger))
router.Use(middleware.CORS(middleware.CORSConfig{
    AllowedOrigins:   []string{"http://localhost:3000"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
    AllowCredentials: true,
    MaxAge:           300,
}))
router.Use(middleware.Timeout(60 * time.Second))
```

### Custom Health Checks

```go
import "github.com/PhilipKram/gms-foundation/pkg/healthcheck"

healthcheck.RegisterChiWithChecks(router,
    func() error { return database.Ping() },   // readiness
    func() error { return checkServices() },    // liveness
)
```

## Package Overview

### Infrastructure

| Package | Description |
|---------|-------------|
| `pkg/logger` | Structured logging via zerolog. Console and JSON (Logstash) output, configurable levels, custom writers. |
| `pkg/server/chi` | Chi router setup with sensible defaults, graceful shutdown, configurable timeouts. |
| `pkg/server` | Gin router setup with similar features to the Chi package. |
| `pkg/healthcheck` | Kubernetes-compatible `/healthz/readiness` and `/healthz/liveness` endpoints. Supports custom check functions. |
| `pkg/prometheus` | Automatic `/metrics` endpoint registration for Gin and Chi routers. |
| `pkg/middleware` | Request logging (with path exclusions), panic recovery, CORS, request timeout. |

### Application Utilities

| Package | Description |
|---------|-------------|
| `pkg/httputil` | `WriteJSON`, `WriteError`, `WritePaginated` — consistent JSON response helpers with `PaginatedResponse` envelope. |
| `pkg/passwords` | Bcrypt wrapper: `Hash` (cost 12), `HashWithCost`, `Check`. |
| `pkg/canonical` | URL normalization (lowercase, strip tracking params, sort query, remove fragments/default ports) + SHA-256 hash. Thread-safe `AddTrackingParams` to extend the strip list. |
| `pkg/dbutil` | `Open` / `OpenMySQL` with functional options (`WithMaxOpenConns`, `WithMaxIdleConns`, `WithConnMaxLifetime`, `WithConnMaxIdleTime`). Defaults: 25 open, 10 idle, 5m lifetime, 2m idle time. |
| `pkg/mongodb` | MongoDB client wrapper with optional CSFLE auto-encryption (bypass mode). Functional options (`WithPingTimeout`, `WithAppName`, `WithDirectConnection`). Plain fallback DB for encrypted collections. |
| `pkg/redis` | Redis client wrapper with standalone and Sentinel failover. Functional options (`WithPingTimeout`, `WithPoolSize`, `WithMinIdleConns`, `WithDialTimeout`, `WithReadTimeout`, `WithWriteTimeout`). Defaults: pool 10, idle 2, dial 5s, read 3s, write 3s. |
| `pkg/uploads` | File storage with configurable categories. Defaults: images (JPEG/PNG/GIF/WebP, 10 MB) and audio (MP3/WAV/M4A/OGG, 50 MB). Magic-byte content validation, UUID filenames, path traversal protection. |
| `pkg/envconfig` | `Required`, `Optional`, `OptionalBool` ("true"/"1"/"yes"), `OptionalStringSlice` (split + trim), `ResolveAbsPath`. |

## Router Compatibility

| Router | Server setup | Health checks | Metrics |
|--------|-------------|---------------|---------|
| **Chi** (`go-chi/chi/v5`) | `pkg/server/chi` | `healthcheck.RegisterChi()` | `prometheus.RegisterChi()` |
| **Gin** (`gin-gonic/gin`) | `pkg/server` | `healthcheck.Register()` | `prometheus.Register()` |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License. See [LICENSE](LICENSE) for details.
