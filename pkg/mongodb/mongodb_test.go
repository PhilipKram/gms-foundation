package mongodb

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultClientConfig(t *testing.T) {
	cfg := defaultClientConfig()
	if cfg.PingTimeout != 5*time.Second {
		t.Errorf("expected PingTimeout 5s, got %v", cfg.PingTimeout)
	}
	if cfg.AppName != "" {
		t.Errorf("expected empty AppName, got %q", cfg.AppName)
	}
	if cfg.DirectConnection != nil {
		t.Errorf("expected nil DirectConnection, got %v", *cfg.DirectConnection)
	}
}

func TestWithPingTimeout(t *testing.T) {
	cfg := defaultClientConfig()
	WithPingTimeout(10 * time.Second)(&cfg)
	if cfg.PingTimeout != 10*time.Second {
		t.Errorf("expected 10s, got %v", cfg.PingTimeout)
	}
}

func TestWithAppName(t *testing.T) {
	cfg := defaultClientConfig()
	WithAppName("test-app")(&cfg)
	if cfg.AppName != "test-app" {
		t.Errorf("expected %q, got %q", "test-app", cfg.AppName)
	}
}

func TestWithDirectConnection(t *testing.T) {
	cfg := defaultClientConfig()
	WithDirectConnection(true)(&cfg)
	if cfg.DirectConnection == nil || !*cfg.DirectConnection {
		t.Error("expected DirectConnection to be true")
	}
}

func TestMultipleOptions(t *testing.T) {
	cfg := defaultClientConfig()
	opts := []Option{
		WithPingTimeout(20 * time.Second),
		WithAppName("multi-test"),
		WithDirectConnection(false),
	}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.PingTimeout != 20*time.Second {
		t.Errorf("expected PingTimeout 20s, got %v", cfg.PingTimeout)
	}
	if cfg.AppName != "multi-test" {
		t.Errorf("expected AppName %q, got %q", "multi-test", cfg.AppName)
	}
	if cfg.DirectConnection == nil || *cfg.DirectConnection {
		t.Error("expected DirectConnection to be false")
	}
}

func TestIsMongocryptError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"matching error", errors.New("mongocrypt: decryption failed"), true},
		{"non-matching error", errors.New("connection refused"), false},
		{"embedded match", errors.New("failed due to mongocrypt issue"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMongocryptError(tt.err); got != tt.want {
				t.Errorf("IsMongocryptError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestConnect_InvalidHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Connect(ctx, Config{
		Host:     "mongodb://localhost:1",
		Database: "testdb",
	}, WithPingTimeout(2*time.Second))
	if err == nil {
		t.Fatal("expected error for unreachable MongoDB host")
	}
}

func TestClient_PlainDB_NilFallback(t *testing.T) {
	// A Client without CSFLE should return the primary DB from PlainDB().
	c := &Client{
		db: nil, // will be nil in this synthetic test
	}
	if c.PlainDB() != nil {
		t.Error("expected nil from PlainDB when both plainDB and db are nil")
	}
}

func TestClient_Close_NilFields(t *testing.T) {
	// Close on a zero-value Client should not panic.
	c := &Client{}
	if err := c.Close(context.Background()); err != nil {
		t.Errorf("expected no error closing zero-value Client, got %v", err)
	}
}
