package redis

import (
	"context"
	"testing"
	"time"
)

func TestDefaultClientConfig(t *testing.T) {
	cfg := defaultClientConfig()
	if cfg.PingTimeout != 5*time.Second {
		t.Errorf("expected PingTimeout 5s, got %v", cfg.PingTimeout)
	}
	if cfg.PoolSize != 10 {
		t.Errorf("expected PoolSize 10, got %d", cfg.PoolSize)
	}
	if cfg.MinIdleConns != 2 {
		t.Errorf("expected MinIdleConns 2, got %d", cfg.MinIdleConns)
	}
	if cfg.DialTimeout != 5*time.Second {
		t.Errorf("expected DialTimeout 5s, got %v", cfg.DialTimeout)
	}
	if cfg.ReadTimeout != 3*time.Second {
		t.Errorf("expected ReadTimeout 3s, got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 3*time.Second {
		t.Errorf("expected WriteTimeout 3s, got %v", cfg.WriteTimeout)
	}
}

func TestWithPingTimeout(t *testing.T) {
	cfg := defaultClientConfig()
	WithPingTimeout(10 * time.Second)(&cfg)
	if cfg.PingTimeout != 10*time.Second {
		t.Errorf("expected 10s, got %v", cfg.PingTimeout)
	}
}

func TestWithPoolSize(t *testing.T) {
	cfg := defaultClientConfig()
	WithPoolSize(50)(&cfg)
	if cfg.PoolSize != 50 {
		t.Errorf("expected 50, got %d", cfg.PoolSize)
	}
}

func TestWithMinIdleConns(t *testing.T) {
	cfg := defaultClientConfig()
	WithMinIdleConns(5)(&cfg)
	if cfg.MinIdleConns != 5 {
		t.Errorf("expected 5, got %d", cfg.MinIdleConns)
	}
}

func TestWithDialTimeout(t *testing.T) {
	cfg := defaultClientConfig()
	WithDialTimeout(15 * time.Second)(&cfg)
	if cfg.DialTimeout != 15*time.Second {
		t.Errorf("expected 15s, got %v", cfg.DialTimeout)
	}
}

func TestWithReadTimeout(t *testing.T) {
	cfg := defaultClientConfig()
	WithReadTimeout(10 * time.Second)(&cfg)
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("expected 10s, got %v", cfg.ReadTimeout)
	}
}

func TestWithWriteTimeout(t *testing.T) {
	cfg := defaultClientConfig()
	WithWriteTimeout(8 * time.Second)(&cfg)
	if cfg.WriteTimeout != 8*time.Second {
		t.Errorf("expected 8s, got %v", cfg.WriteTimeout)
	}
}

func TestMultipleOptions(t *testing.T) {
	cfg := defaultClientConfig()
	opts := []Option{
		WithPingTimeout(20 * time.Second),
		WithPoolSize(100),
		WithMinIdleConns(10),
		WithDialTimeout(30 * time.Second),
		WithReadTimeout(15 * time.Second),
		WithWriteTimeout(12 * time.Second),
	}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.PingTimeout != 20*time.Second {
		t.Errorf("expected PingTimeout 20s, got %v", cfg.PingTimeout)
	}
	if cfg.PoolSize != 100 {
		t.Errorf("expected PoolSize 100, got %d", cfg.PoolSize)
	}
	if cfg.MinIdleConns != 10 {
		t.Errorf("expected MinIdleConns 10, got %d", cfg.MinIdleConns)
	}
	if cfg.DialTimeout != 30*time.Second {
		t.Errorf("expected DialTimeout 30s, got %v", cfg.DialTimeout)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("expected ReadTimeout 15s, got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 12*time.Second {
		t.Errorf("expected WriteTimeout 12s, got %v", cfg.WriteTimeout)
	}
}

func TestResolveAddr(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{"host:port passthrough", "myhost:6380", "myhost:6380"},
		{"host without port", "myhost", "myhost:6379"},
		{"empty defaults", "", "localhost:6379"},
		{"ipv4 with port", "127.0.0.1:6380", "127.0.0.1:6380"},
		{"ipv4 without port", "127.0.0.1", "127.0.0.1:6379"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveAddr(tt.addr)
			if got != tt.want {
				t.Errorf("resolveAddr(%q) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}

func TestConnect_InvalidHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := Connect(ctx, Config{
		Addr: "localhost:1",
	}, WithPingTimeout(1*time.Second), WithDialTimeout(1*time.Second))
	if err == nil {
		t.Fatal("expected error for unreachable Redis host")
	}
}
