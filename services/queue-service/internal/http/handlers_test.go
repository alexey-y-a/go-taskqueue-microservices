package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTaskReturnID(t *testing.T) {
    s := NewServer()
    mux := s.Mux()

    body := []byte(`{"type":"email","payload":"hello"}`)
    req := httptest.NewRequest(http.MethodPost, "/internal/tasks", bytes.NewReader(body))
    w := httptest.NewRecorder()

    mux.ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("unexpected status: got %d, want %d", w.Code, http.StatusCreated)
    }
    if !bytes.Contains(w.Body.Bytes(), []byte(`"id"`)) {
        t.Fatalf("response must contain id, got: %s", w.Body.String())
    }
}
