package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func newTestLogger(buf *bytes.Buffer) zerolog.Logger {
	return zerolog.New(buf).With().Logger()
}

func TestRequestLogger_200_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshaling log: %v", err)
	}
	if entry["level"] != "info" {
		t.Errorf("expected info level for 200, got %v", entry["level"])
	}
	if entry["path"] != "/ok" {
		t.Errorf("expected path /ok, got %v", entry["path"])
	}
}

func TestRequestLogger_404_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshaling log: %v", err)
	}
	if entry["level"] != "warn" {
		t.Errorf("expected warn level for 404, got %v", entry["level"])
	}
}

func TestRequestLogger_500_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodPost, "/fail", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshaling log: %v", err)
	}
	if entry["level"] != "error" {
		t.Errorf("expected error level for 500, got %v", entry["level"])
	}
	if entry["method"] != "POST" {
		t.Errorf("expected method POST, got %v", entry["method"])
	}
}

func TestRequestLoggerWithSkip(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := RequestLoggerWithSkip(logger, []string{"/healthz"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Skipped path should produce no log
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if buf.Len() > 0 {
		t.Error("expected no log output for skipped path")
	}

	// Non-skipped path should produce log
	buf.Reset()
	req = httptest.NewRequest(http.MethodGet, "/api", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if buf.Len() == 0 {
		t.Error("expected log output for non-skipped path")
	}
}

func TestRequestLogger_CapturesMethodPathStatusBytes(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodPut, "/resource", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshaling log: %v", err)
	}
	if entry["method"] != "PUT" {
		t.Errorf("expected PUT, got %v", entry["method"])
	}
	if entry["path"] != "/resource" {
		t.Errorf("expected /resource, got %v", entry["path"])
	}
	if int(entry["status"].(float64)) != 201 {
		t.Errorf("expected status 201, got %v", entry["status"])
	}
	if int(entry["bytes"].(float64)) != 5 {
		t.Errorf("expected 5 bytes, got %v", entry["bytes"])
	}
}
