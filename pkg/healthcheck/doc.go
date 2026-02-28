// Package healthcheck registers /healthz/readiness and /healthz/liveness
// endpoints for both Gin and Chi routers. Optional HealthCheckFunc callbacks
// can be provided to perform custom readiness and liveness checks.
package healthcheck
