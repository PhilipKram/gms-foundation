package healthcheck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterChi_Returns200(t *testing.T) {
	r := chi.NewRouter()
	RegisterChi(r)

	for _, path := range []string{"/healthz/readiness", "/healthz/liveness"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", path, w.Code)
		}
	}
}

func TestRegisterChiWithChecks_NilChecks(t *testing.T) {
	r := chi.NewRouter()
	RegisterChiWithChecks(r, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz/readiness", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for nil check, got %d", w.Code)
	}
}

func TestRegisterChiWithChecks_PassingCheck(t *testing.T) {
	r := chi.NewRouter()
	RegisterChiWithChecks(r, func() error { return nil }, func() error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/healthz/readiness", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for passing check, got %d", w.Code)
	}
}

func TestRegisterChiWithChecks_FailingCheck(t *testing.T) {
	r := chi.NewRouter()
	RegisterChiWithChecks(r, func() error { return errors.New("db down") }, nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz/readiness", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for failing check, got %d", w.Code)
	}
}
