package healthcheck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRegisterGin_Returns200(t *testing.T) {
	r := gin.New()
	Register(r)

	for _, path := range []string{"/healthz/readiness", "/healthz/liveness"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", path, w.Code)
		}
	}
}

func TestRegisterWithChecks_NilChecks(t *testing.T) {
	r := gin.New()
	RegisterWithChecks(r, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz/readiness", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for nil check, got %d", w.Code)
	}
}

func TestRegisterWithChecks_PassingCheck(t *testing.T) {
	r := gin.New()
	RegisterWithChecks(r, func() error { return nil }, func() error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/healthz/liveness", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for passing check, got %d", w.Code)
	}
}

func TestRegisterWithChecks_FailingCheck(t *testing.T) {
	r := gin.New()
	RegisterWithChecks(r, nil, func() error { return errors.New("unhealthy") })

	req := httptest.NewRequest(http.MethodGet, "/healthz/liveness", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for failing check, got %d", w.Code)
	}
}
