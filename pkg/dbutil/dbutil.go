package dbutil

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// PoolConfig holds connection pool settings for a *sql.DB.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultPoolConfig returns the default pool configuration.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}
}

// Option customizes pool configuration.
type Option func(*PoolConfig)

// WithMaxOpenConns sets the maximum number of open connections.
func WithMaxOpenConns(n int) Option {
	return func(c *PoolConfig) { c.MaxOpenConns = n }
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) Option {
	return func(c *PoolConfig) { c.MaxIdleConns = n }
}

// WithConnMaxLifetime sets the maximum lifetime of a connection.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *PoolConfig) { c.ConnMaxLifetime = d }
}

// WithConnMaxIdleTime sets the maximum idle time of a connection.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *PoolConfig) { c.ConnMaxIdleTime = d }
}

// Open opens a database connection using the given driver and DSN, configures
// the connection pool, and pings the database to verify connectivity.
func Open(driverName, dsn string, opts ...Option) (*sql.DB, error) {
	cfg := DefaultPoolConfig()
	for _, o := range opts {
		o(&cfg)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}

// OpenMySQL opens a MySQL/MariaDB connection using the "mysql" driver.
func OpenMySQL(dsn string, opts ...Option) (*sql.DB, error) {
	return Open("mysql", dsn, opts...)
}
