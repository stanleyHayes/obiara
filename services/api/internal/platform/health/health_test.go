package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLive(t *testing.T) {
	recorder := httptest.NewRecorder()
	Live().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/live", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("live status = %d, want 200", recorder.Code)
	}
}

func TestReadyWhenDependencyHealthy(t *testing.T) {
	recorder := httptest.NewRecorder()
	Ready(func(context.Context) error { return nil }).ServeHTTP(
		recorder, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("ready status = %d, want 200", recorder.Code)
	}
}

func TestReadyWhenDependencyDegraded(t *testing.T) {
	recorder := httptest.NewRecorder()
	Ready(func(context.Context) error { return errors.New("mongo unreachable") }).ServeHTTP(
		recorder, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready status = %d, want 503", recorder.Code)
	}
	if body := recorder.Body.String(); body != "dependency unavailable" {
		t.Fatalf("ready body = %q, must not leak dependency internals", body)
	}
}
