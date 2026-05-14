package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMain(m *testing.M) {
}

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	HealthCheckHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Fatalf(`expected "ok", got %q`, w.Body.String())
	}
}
