package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", body["key"])
	}
}

func TestWriteJSON_CustomStatus(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusCreated, map[string]int{"id": 42})

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusBadRequest, "something went wrong")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if body["error"] != "something went wrong" {
		t.Errorf("expected error message, got %q", body["error"])
	}
}

func TestWriteError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusNotFound, "not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestWritePaginated(t *testing.T) {
	w := httptest.NewRecorder()
	items := []string{"a", "b", "c"}
	WritePaginated(w, http.StatusOK, items, 100, 10, 20)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var body PaginatedResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if body.Total != 100 {
		t.Errorf("expected total 100, got %d", body.Total)
	}
	if body.Limit != 10 {
		t.Errorf("expected limit 10, got %d", body.Limit)
	}
	if body.Offset != 20 {
		t.Errorf("expected offset 20, got %d", body.Offset)
	}
}

func TestWriteJSON_NilValue(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusOK, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if body != "null\n" {
		t.Errorf("expected null JSON, got %q", body)
	}
}

func TestWriteJSON_SlicePayload(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusOK, []int{1, 2, 3})

	var body []int
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if len(body) != 3 || body[0] != 1 {
		t.Errorf("expected [1 2 3], got %v", body)
	}
}

func TestWriteError_ServerError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusInternalServerError, "internal error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if body["error"] != "internal error" {
		t.Errorf("expected 'internal error', got %q", body["error"])
	}
}

func TestWritePaginated_ItemsContent(t *testing.T) {
	w := httptest.NewRecorder()
	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	items := []item{{1, "a"}, {2, "b"}}
	WritePaginated(w, http.StatusOK, items, 2, 10, 0)

	var body struct {
		Items  []item `json:"items"`
		Total  int    `json:"total"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if len(body.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(body.Items))
	}
	if body.Items[0].Name != "a" {
		t.Errorf("expected first item name 'a', got %q", body.Items[0].Name)
	}
}

func TestWritePaginated_EmptyItems(t *testing.T) {
	w := httptest.NewRecorder()
	WritePaginated(w, http.StatusOK, []string{}, 0, 10, 0)

	var body PaginatedResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if body.Total != 0 {
		t.Errorf("expected total 0, got %d", body.Total)
	}
}
