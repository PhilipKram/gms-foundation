package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	defaultPingTimeout  = 5 * time.Second
	defaultPoolSize     = 10
	defaultMinIdleConns = 2
	defaultDialTimeout  = 5 * time.Second
	defaultReadTimeout  = 3 * time.Second
	defaultWriteTimeout = 3 * time.Second
	defaultRedisPort    = "6379"
)

// AuthConfig holds Redis authentication credentials.
type AuthConfig struct {
	Username string
	Password string
}

// SentinelConfig holds Redis Sentinel settings.
// When non-nil, the client connects via Sentinel failover.
type SentinelConfig struct {
	MasterName string
	Nodes      []string
}

// Config holds the settings needed to connect to Redis.
type Config struct {
	Addr     string          // host:port for standalone mode
	DB       int             // database number
	Auth     AuthConfig      // authentication credentials
	Sentinel *SentinelConfig // nil = standalone mode
}

// clientConfig holds internal options applied via Option functions.
type clientConfig struct {
	PingTimeout  time.Duration
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func defaultClientConfig() clientConfig {
	return clientConfig{
		PingTimeout:  defaultPingTimeout,
		PoolSize:     defaultPoolSize,
		MinIdleConns: defaultMinIdleConns,
		DialTimeout:  defaultDialTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}
}

// Option customizes client configuration.
type Option func(*clientConfig)

// WithPingTimeout sets the timeout for the initial ping.
func WithPingTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.PingTimeout = d }
}

// WithPoolSize sets the maximum number of connections in the pool.
func WithPoolSize(n int) Option {
	return func(c *clientConfig) { c.PoolSize = n }
}

// WithMinIdleConns sets the minimum number of idle connections.
func WithMinIdleConns(n int) Option {
	return func(c *clientConfig) { c.MinIdleConns = n }
}

// WithDialTimeout sets the timeout for establishing new connections.
func WithDialTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.DialTimeout = d }
}

// WithReadTimeout sets the timeout for read operations.
func WithReadTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.ReadTimeout = d }
}

// WithWriteTimeout sets the timeout for write operations.
func WithWriteTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.WriteTimeout = d }
}

// Client wraps a Redis universal client.
type Client struct {
	client goredis.UniversalClient
}

// Ping verifies the connection to Redis.
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// Unwrap returns the underlying go-redis UniversalClient for direct access.
func (c *Client) Unwrap() goredis.UniversalClient {
	return c.client
}

// resolveAddr ensures the address contains a port, defaulting to 6379.
func resolveAddr(addr string) string {
	if addr == "" {
		return "localhost:" + defaultRedisPort
	}
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr
		}
		if addr[i] == ']' {
			// IPv6 without port
			break
		}
	}
	return addr + ":" + defaultRedisPort
}

// Connect creates a new Redis client, pings it, and returns a wrapper.
// When cfg.Sentinel is non-nil, a Sentinel failover client is used;
// otherwise a standalone client connects to cfg.Addr.
func Connect(ctx context.Context, cfg Config, opts ...Option) (*Client, error) {
	cc := defaultClientConfig()
	for _, o := range opts {
		o(&cc)
	}

	var uc goredis.UniversalClient

	if cfg.Sentinel != nil && cfg.Sentinel.MasterName != "" {
		uc = goredis.NewFailoverClient(&goredis.FailoverOptions{
			MasterName:    cfg.Sentinel.MasterName,
			SentinelAddrs: cfg.Sentinel.Nodes,
			Username:      cfg.Auth.Username,
			Password:      cfg.Auth.Password,
			DB:            cfg.DB,
			PoolSize:      cc.PoolSize,
			MinIdleConns:  cc.MinIdleConns,
			DialTimeout:   cc.DialTimeout,
			ReadTimeout:   cc.ReadTimeout,
			WriteTimeout:  cc.WriteTimeout,
		})
	} else {
		uc = goredis.NewClient(&goredis.Options{
			Addr:         resolveAddr(cfg.Addr),
			Username:     cfg.Auth.Username,
			Password:     cfg.Auth.Password,
			DB:           cfg.DB,
			PoolSize:     cc.PoolSize,
			MinIdleConns: cc.MinIdleConns,
			DialTimeout:  cc.DialTimeout,
			ReadTimeout:  cc.ReadTimeout,
			WriteTimeout: cc.WriteTimeout,
		})
	}

	pingCtx, cancel := context.WithTimeout(ctx, cc.PingTimeout)
	defer cancel()

	if err := uc.Ping(pingCtx).Err(); err != nil {
		_ = uc.Close()
		return nil, fmt.Errorf("pinging Redis: %w", err)
	}

	return &Client{client: uc}, nil
}
