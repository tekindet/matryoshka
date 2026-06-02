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

func TestPostgresContainerStart(t *testing.T) {
	// some configs for a container that we will get
	// from the user????

	// is it a network test or just a function test

	// we need a dockerfile here
	StartPostgresContainer()
}
