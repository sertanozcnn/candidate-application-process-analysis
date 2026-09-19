package interaction

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const sessionCookieName = "capa_interaction_session"

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type Handler struct {
	service       Service
	secureCookies bool
}

type createSessionRequest struct {
	PositionCode string `json:"position_code"`
}
type createSessionResponse struct {
	ID           string `json:"id"`
	PositionCode string `json:"position_code"`
}
type eventBatchRequest struct {
	Events []Event `json:"events"`
}
type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RegisterRoutes(mux *http.ServeMux, service Service, appEnv string) {
	h := Handler{service: service, secureCookies: appEnv == "production"}
	mux.HandleFunc("/v1/candidate/interaction-sessions", h.createSession)
	mux.HandleFunc("/v1/candidate/interaction-sessions/", h.events)
}

func (h Handler) createSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var input createSessionRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	token, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	session, err := h.service.CreateSession(r.Context(), input.PositionCode, hashToken(token))
	if errors.Is(err, ErrInvalidSession) {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: token, Path: "/v1/candidate", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode})
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/v1/candidate/interaction-sessions", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	writeJSON(w, http.StatusCreated, createSessionResponse{ID: session.ID, PositionCode: session.PositionCode})
}

func (h Handler) events(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	sessionID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/candidate/interaction-sessions/"), "/events")
	if !strings.HasSuffix(r.URL.Path, "/events") || !uuidPattern.MatchString(sessionID) {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var input eventBatchRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	result, err := h.service.AppendEvents(r.Context(), sessionID, hashToken(cookie.Value), input.Events)
	if err != nil {
		switch {
		case errors.Is(err, ErrSessionNotFound):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, ErrSessionExpired):
			writeError(w, http.StatusGone, "session_expired")
		case errors.Is(err, ErrSequenceOrder):
			writeError(w, http.StatusConflict, "sequence_conflict")
		case errors.Is(err, ErrInvalidBatch):
			writeError(w, http.StatusBadRequest, "invalid_request")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("multiple JSON values")
	}
	return nil
}

func randomToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, errorResponse{Code: code, Message: "İstek geçersiz."})
}
func methodNotAllowed(w http.ResponseWriter, method string) {
	w.Header().Set("Allow", method)
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
}
