package adminauth

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const sessionCookieName = "capa_admin_session"
const csrfCookieName = "capa_admin_csrf"

type Handler struct {
	service       Service
	secureCookies bool
	adminOrigins  map[string]struct{}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type adminResponse struct {
	Admin Admin `json:"admin"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewHandler(service Service, appEnv string) *Handler {
	return NewHandlerWithOrigins(service, appEnv, []string{"http://localhost:3001", "http://127.0.0.1:3001"})
}

func NewHandlerWithOrigins(service Service, appEnv string, origins []string) *Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		if origin = strings.TrimSpace(origin); origin != "" {
			allowed[origin] = struct{}{}
		}
	}
	return &Handler{service: service, secureCookies: appEnv == "production", adminOrigins: allowed}
}

func RegisterRoutes(mux *http.ServeMux, service Service, appEnv string, origins []string) {
	h := NewHandlerWithOrigins(service, appEnv, origins)
	mux.HandleFunc("/v1/admin/auth/login", h.login)
	mux.HandleFunc("/v1/admin/auth/logout", h.logout)
	mux.HandleFunc("/v1/admin/auth/me", h.me)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	if !h.validOrigin(r) {
		writeError(w, http.StatusForbidden, "forbidden_origin", "İstek kaynağına izin verilmiyor.")
		return
	}

	var input loginRequest
	if err := decodeJSON(w, r, &input); err != nil || strings.TrimSpace(input.Email) == "" || input.Password == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "İstek geçersiz.")
		return
	}

	result, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "E-posta veya parola hatalı.")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Beklenmeyen bir hata oluştu.")
		return
	}

	h.setSessionCookie(w, result.SessionToken, result.ExpiresAt)
	writeJSON(w, http.StatusOK, adminResponse{Admin: result.Admin})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	if !h.validOrigin(r) {
		writeError(w, http.StatusForbidden, "forbidden_origin", "İstek kaynağına izin verilmiyor.")
		return
	}
	if !h.validCSRF(r) {
		writeError(w, http.StatusForbidden, "csrf_failed", "İstek doğrulanamadı.")
		return
	}

	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "Beklenmeyen bir hata oluştu.")
			return
		}
	}

	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum gerekli.")
		return
	}

	admin, err := h.service.Authenticate(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum gerekli.")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Beklenmeyen bir hata oluştu.")
		return
	}

	writeJSON(w, http.StatusOK, adminResponse{Admin: admin})
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: token, Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, Expires: expiresAt, MaxAge: maxAge})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode, Expires: time.Unix(1, 0), MaxAge: -1})
}

func (h *Handler) validOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	_, ok := h.adminOrigins[origin]
	return origin != "" && ok
}

func (h *Handler) validCSRF(r *http.Request) bool {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	header := r.Header.Get("X-CSRF-Token")
	return header != "" && len(header) == len(cookie.Value) && subtle.ConstantTimeCompare([]byte(header), []byte(cookie.Value)) == 1
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	err := decoder.Decode(&extra)
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.New("request has multiple JSON values")
	}
	if err == nil {
		return errors.New("request has multiple JSON values")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}

func methodNotAllowed(w http.ResponseWriter, method string) {
	w.Header().Set("Allow", method)
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method desteklenmiyor.")
}
