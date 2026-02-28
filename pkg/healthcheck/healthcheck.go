package healthcheck

// HealthCheckFunc is a function that performs a health check.
// It should return an error if the check fails, or nil if healthy.
type HealthCheckFunc func() error
