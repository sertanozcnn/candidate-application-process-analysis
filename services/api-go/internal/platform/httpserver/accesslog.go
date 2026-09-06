package httpserver

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func AccessLog(log zerolog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", recorder.status).
			Dur("latency", time.Since(start)).
			Str("client_ip", r.RemoteAddr).
			Msg("http request")
	})
}
