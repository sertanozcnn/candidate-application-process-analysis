package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSAllowsConfiguredCandidateOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := CORS([]string{"http://localhost:3000"}, next)

	request := httptest.NewRequest(http.MethodOptions, "/v1/candidate/applications", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("allow origin = %q, want configured origin", got)
	}
}

func TestCORSDoesNotAllowUnknownOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := CORS([]string{"http://localhost:3000"}, next)

	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/applications", nil)
	request.Header.Set("Origin", "https://evil.example")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unexpected CORS header for unknown origin")
	}
}
