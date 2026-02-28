package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

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

func TestHandleRequestBody_UnsupportedContentType(t *testing.T) {
	_, router := Setup(Config{Port: "0"})
	router.POST("/test", func(c *gin.Context) {
		err := HandleRequestBody(c, "text/plain", &struct{}{})
		if err == nil {
			t.Fatal("expected error for unsupported content type")
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("body"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", w.Code)
	}
}

func TestHandleRequestBody_NonProtoMessage(t *testing.T) {
	_, router := Setup(Config{Port: "0"})
	router.POST("/test", func(c *gin.Context) {
		// Pass a non-proto.Message struct
		var out struct{ Name string }
		err := HandleRequestBody(c, "application/json", &out)
		if err == nil {
			t.Fatal("expected error for non-proto.Message")
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(`{"name":"test"}`)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleRequestBody_NilPointer(t *testing.T) {
	_, router := Setup(Config{Port: "0"})
	router.POST("/test", func(c *gin.Context) {
		err := HandleRequestBody(c, "application/json", nil)
		if err == nil {
			t.Fatal("expected error for nil pointer")
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}
