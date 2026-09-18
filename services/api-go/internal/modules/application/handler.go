package application

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"capa/services/api-go/internal/modules/adminauth"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type adminAuthenticator interface {
	Authenticate(context.Context, string) (adminauth.Admin, error)
}

type Handler struct {
	service Service
	auth    adminAuthenticator
}

type applicationResponse struct {
	Application Application `json:"application"`
}

type applicationListResponse struct {
	Applications []Application `json:"applications"`
	Limit        int           `json:"limit"`
	Offset       int           `json:"offset"`
}

type createResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PositionCode string `json:"position_code"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RegisterRoutes(mux *http.ServeMux, service Service, auth adminAuthenticator) {
	h := Handler{service: service, auth: auth}
	mux.HandleFunc("/v1/candidate/applications", h.create)
	mux.HandleFunc("/v1/admin/applications", h.list)
	mux.HandleFunc("/v1/admin/applications/", h.detail)
}

func (h Handler) create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var input CreateInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "İstek geçersiz.")
		return
	}

	item, err := h.service.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidApplication):
			writeError(w, http.StatusBadRequest, "invalid_request", "İstek geçersiz.")
		case errors.Is(err, ErrApplicationExists):
			writeError(w, http.StatusConflict, "application_exists", "Bu pozisyon için başvuru zaten mevcut.")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Beklenmeyen bir hata oluştu.")
		}
		return
	}

	writeJSON(w, http.StatusCreated, createResponse{ID: item.ID, Email: item.Email, PositionCode: item.PositionCode})
}

func (h Handler) list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	if !h.authenticateAdmin(w, r) {
		return
	}

	limit, offset, ok := pagination(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_request", "İstek geçersiz.")
		return
	}
	items, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		if errors.Is(err, ErrInvalidApplication) {
			writeError(w, http.StatusBadRequest, "invalid_request", "İstek geçersiz.")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Beklenmeyen bir hata oluştu.")
		return
	}
	writeJSON(w, http.StatusOK, applicationListResponse{Applications: items, Limit: limit, Offset: offset})
}

func (h Handler) detail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	if !h.authenticateAdmin(w, r) {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/admin/applications/")
	if strings.Contains(id, "/") || !uuidPattern.MatchString(id) {
		writeError(w, http.StatusNotFound, "not_found", "Başvuru bulunamadı.")
		return
	}
	item, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrApplicationNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Başvuru bulunamadı.")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Beklenmeyen bir hata oluştu.")
		return
	}
	writeJSON(w, http.StatusOK, applicationResponse{Application: item})
}

func (h Handler) authenticateAdmin(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("capa_admin_session")
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum gerekli.")
		return false
	}
	if _, err := h.auth.Authenticate(r.Context(), cookie.Value); err != nil {
		if errors.Is(err, adminauth.ErrUnauthorized) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Oturum gerekli.")
			return false
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Beklenmeyen bir hata oluştu.")
		return false
	}
	return true
}

func pagination(r *http.Request) (int, int, bool) {
	limit := 20
	offset := 0
	query := r.URL.Query()
	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, 0, false
		}
		limit = parsed
	}
	if value := query.Get("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, 0, false
		}
		offset = parsed
	}
	return limit, offset, limit >= 1 && limit <= 100 && offset >= 0
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
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

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}

func methodNotAllowed(w http.ResponseWriter, method string) {
	w.Header().Set("Allow", method)
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method desteklenmiyor.")
}
