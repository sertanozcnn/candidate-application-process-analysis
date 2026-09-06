package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubPinger struct {
	err error
}

func (s stubPinger) Ping(context.Context) error {
	return s.err
}

func TestLive(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, stubPinger{})

	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	responseRecorder := httptest.NewRecorder()

	mux.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusOK)
	}

	var body response
	decodeBody(t, responseRecorder, &body)
	if body.Status != "ok" {
		t.Fatalf("status body = %q, want ok", body.Status)
	}
}

func TestLiveRejectsNonGet(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, stubPinger{})

	request := httptest.NewRequest(http.MethodPost, "/health/live", nil)
	responseRecorder := httptest.NewRecorder()

	mux.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusMethodNotAllowed)
	}
	if allow := responseRecorder.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("Allow = %q, want GET", allow)
	}
}

func TestReady(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, stubPinger{})

	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	responseRecorder := httptest.NewRecorder()

	mux.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusOK)
	}

	var body response
	decodeBody(t, responseRecorder, &body)
	if body.Status != "ready" {
		t.Fatalf("status body = %q, want ready", body.Status)
	}
}

func TestReadyReturnsUnavailableWhenDatabasePingFails(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, stubPinger{err: errors.New("database unavailable")})

	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	responseRecorder := httptest.NewRecorder()

	mux.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusServiceUnavailable)
	}

	var body response
	decodeBody(t, responseRecorder, &body)
	if body.Status != "not_ready" {
		t.Fatalf("status body = %q, want not_ready", body.Status)
	}
}

func TestReadyRejectsNonGet(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, stubPinger{})

	request := httptest.NewRequest(http.MethodPost, "/health/ready", nil)
	responseRecorder := httptest.NewRecorder()

	mux.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusMethodNotAllowed)
	}
	if allow := responseRecorder.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("Allow = %q, want GET", allow)
	}
}

func decodeBody(t *testing.T, responseRecorder *httptest.ResponseRecorder, target any) {
	t.Helper()

	if contentType := responseRecorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	if err := json.NewDecoder(responseRecorder.Body).Decode(target); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
}
