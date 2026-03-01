// Package redis provides a wrapper for connecting to Redis in standalone or
// Sentinel failover mode. The Client type supports connection pooling and
// configurable timeouts via functional options. Use Unwrap to access the
// underlying go-redis UniversalClient for direct command execution.
package redis
