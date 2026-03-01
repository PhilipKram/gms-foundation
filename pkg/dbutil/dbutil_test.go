package dbutil

import (
	"testing"
	"time"
)

func TestDefaultPoolConfig(t *testing.T) {
	cfg := DefaultPoolConfig()
	if cfg.MaxOpenConns != 25 {
		t.Errorf("expected MaxOpenConns 25, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 10 {
		t.Errorf("expected MaxIdleConns 10, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("expected ConnMaxLifetime 5m, got %v", cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime != 2*time.Minute {
		t.Errorf("expected ConnMaxIdleTime 2m, got %v", cfg.ConnMaxIdleTime)
	}
	if !cfg.Ping {
		t.Error("expected Ping to default to true")
	}
}

func TestWithMaxOpenConns(t *testing.T) {
	cfg := DefaultPoolConfig()
	WithMaxOpenConns(50)(&cfg)
	if cfg.MaxOpenConns != 50 {
		t.Errorf("expected 50, got %d", cfg.MaxOpenConns)
	}
}

func TestWithMaxIdleConns(t *testing.T) {
	cfg := DefaultPoolConfig()
	WithMaxIdleConns(20)(&cfg)
	if cfg.MaxIdleConns != 20 {
		t.Errorf("expected 20, got %d", cfg.MaxIdleConns)
	}
}

func TestWithConnMaxLifetime(t *testing.T) {
	cfg := DefaultPoolConfig()
	WithConnMaxLifetime(10 * time.Minute)(&cfg)
	if cfg.ConnMaxLifetime != 10*time.Minute {
		t.Errorf("expected 10m, got %v", cfg.ConnMaxLifetime)
	}
}

func TestWithConnMaxIdleTime(t *testing.T) {
	cfg := DefaultPoolConfig()
	WithConnMaxIdleTime(30 * time.Second)(&cfg)
	if cfg.ConnMaxIdleTime != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.ConnMaxIdleTime)
	}
}

func TestWithPing(t *testing.T) {
	cfg := DefaultPoolConfig()
	if !cfg.Ping {
		t.Fatal("expected default Ping=true")
	}
	WithPing(false)(&cfg)
	if cfg.Ping {
		t.Error("expected Ping=false after WithPing(false)")
	}
}

func TestMultipleOptions(t *testing.T) {
	cfg := DefaultPoolConfig()
	opts := []Option{
		WithMaxOpenConns(100),
		WithMaxIdleConns(50),
		WithConnMaxLifetime(15 * time.Minute),
		WithConnMaxIdleTime(5 * time.Minute),
	}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.MaxOpenConns != 100 {
		t.Errorf("expected MaxOpenConns 100, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 50 {
		t.Errorf("expected MaxIdleConns 50, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 15*time.Minute {
		t.Errorf("expected ConnMaxLifetime 15m, got %v", cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime != 5*time.Minute {
		t.Errorf("expected ConnMaxIdleTime 5m, got %v", cfg.ConnMaxIdleTime)
	}
}

func TestOpen_WithOptions(t *testing.T) {
	// sql.Open for mysql succeeds even with bad DSN; only Ping fails.
	// This tests that options are applied before the ping attempt.
	_, err := Open("mysql", "bad:dsn@tcp(localhost:1)/db",
		WithMaxOpenConns(5),
		WithMaxIdleConns(2),
		WithConnMaxLifetime(1*time.Minute),
		WithConnMaxIdleTime(30*time.Second),
	)
	// Should fail at ping, proving the code path through options + pool config ran
	if err == nil {
		t.Fatal("expected error from ping")
	}
}

func TestOpen_NoPing(t *testing.T) {
	// With Ping disabled, Open should succeed even with an unreachable DSN.
	db, err := Open("mysql", "bad:dsn@tcp(localhost:1)/db", WithPing(false))
	if err != nil {
		t.Fatalf("expected no error with ping disabled, got: %v", err)
	}
	_ = db.Close()
}

func TestOpen_InvalidDriver(t *testing.T) {
	_, err := Open("nosuchdriver", "dsn")
	if err == nil {
		t.Fatal("expected error for unknown driver")
	}
}

func TestOpenMySQL_InvalidDSN(t *testing.T) {
	_, err := OpenMySQL("invalid:dsn@tcp(localhost:99999)/nonexistent")
	if err == nil {
		t.Fatal("expected error for unreachable MySQL")
	}
}
