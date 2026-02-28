package chi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetup_ReturnsValidServerAndRouter(t *testing.T) {
	srv, router := Setup(Config{Port: "0"})
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
	if router == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestSetup_DefaultTimeouts(t *testing.T) {
	srv, _ := Setup(Config{Port: "0"})
	if srv.ReadTimeout != 15*time.Second {
		t.Errorf("expected ReadTimeout 15s, got %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 60*time.Second {
		t.Errorf("expected WriteTimeout 60s, got %v", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 120*time.Second {
		t.Errorf("expected IdleTimeout 120s, got %v", srv.IdleTimeout)
	}
}

func TestSetup_CustomTimeouts(t *testing.T) {
	srv, _ := Setup(Config{
		Port:         "0",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})
	if srv.ReadTimeout != 5*time.Second {
		t.Errorf("expected ReadTimeout 5s, got %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 10*time.Second {
		t.Errorf("expected WriteTimeout 10s, got %v", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 30*time.Second {
		t.Errorf("expected IdleTimeout 30s, got %v", srv.IdleTimeout)
	}
}

func TestSetup_AccessLogMiddleware(t *testing.T) {
	_, router := Setup(Config{Port: "0", AccessLog: true})
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
