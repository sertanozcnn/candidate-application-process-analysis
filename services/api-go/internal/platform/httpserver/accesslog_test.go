package httpserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestAccessLogWritesRequestSummary(t *testing.T) {
	var output bytes.Buffer
	log := zerolog.New(&output)

	handler := AccessLog(log, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	request := httptest.NewRequest(http.MethodPost, "/health/live?token=secret", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	logLine := output.String()
	for _, want := range []string{
		`"method":"POST"`,
		`"path":"/health/live"`,
		`"status":201`,
		`"message":"http request"`,
	} {
		if !strings.Contains(logLine, want) {
			t.Fatalf("log line %q does not contain %q", logLine, want)
		}
	}
	if strings.Contains(logLine, "token=secret") {
		t.Fatalf("log line leaked query string: %q", logLine)
	}
}

func TestAccessLogCanUseConsoleWriter(t *testing.T) {
	var output bytes.Buffer
	writer := zerolog.ConsoleWriter{
		Out:        &output,
		NoColor:    false,
		TimeFormat: time.Kitchen,
	}
	log := zerolog.New(writer).With().Timestamp().Logger()

	handler := AccessLog(log, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	logLine := output.String()
	for _, want := range []string{"GET", "/health/ready", "200", "http request"} {
		if !strings.Contains(logLine, want) {
			t.Fatalf("console log line %q does not contain %q", logLine, want)
		}
	}
}
