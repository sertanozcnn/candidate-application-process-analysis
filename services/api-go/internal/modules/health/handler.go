package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const pingTimeout = 2 * time.Second

type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	db Pinger
}

type response struct {
	Status string `json:"status"`
}

func NewHandler(db Pinger) *Handler {
	return &Handler{db: db}
}

func RegisterRoutes(mux *http.ServeMux, db Pinger) {
	handler := NewHandler(db)
	mux.HandleFunc("/health/live", handler.live)
	mux.HandleFunc("/health/ready", handler.ready)
}

func (h *Handler) live(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, response{Status: "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), pingTimeout)
	defer cancel()

	if h.db == nil || h.db.Ping(ctx) != nil {
		writeJSON(w, http.StatusServiceUnavailable, response{Status: "not_ready"})
		return
	}

	writeJSON(w, http.StatusOK, response{Status: "ready"})
}

func writeJSON(w http.ResponseWriter, status int, payload response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
