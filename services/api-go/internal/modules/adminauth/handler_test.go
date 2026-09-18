package adminauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"capa/services/api-go/internal/platform/password"
)

func newHandlerTestService(t *testing.T) (*Service, *fakeStore) {
	t.Helper()
	hash, err := password.Hash("correct-password")
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}
	store := &fakeStore{credentials: adminCredentials{
		Admin:        Admin{ID: "admin-id", Email: "admin@example.com"},
		PasswordHash: hash,
	}}
	service := NewService(store)
	service.now = func() time.Time { return time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC) }
	return &service, store
}

func TestLoginReturnsAdminAndSessionCookie(t *testing.T) {
	service, _ := newHandlerTestService(t)
	handler := NewHandler(*service, "development")
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/auth/login", strings.NewReader(`{"email":" admin@example.com ","password":"correct-password"}`))
	req.Header.Set("Origin", "http://localhost:3001")
	response := httptest.NewRecorder()

	handler.login(response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var payload adminResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Admin.ID != "admin-id" || payload.Admin.Email != "admin@example.com" {
		t.Fatalf("admin = %+v, want public admin fields", payload.Admin)
	}
	if !strings.Contains(response.Body.String(), `"id":"admin-id"`) || !strings.Contains(response.Body.String(), `"email":"admin@example.com"`) {
		t.Fatalf("body = %s, want lowercase JSON fields", response.Body.String())
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Name != sessionCookieName || cookie.Value == "" || !cookie.HttpOnly || cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" {
		t.Fatalf("unexpected session cookie: %+v", cookie)
	}
	if strings.Contains(response.Body.String(), "correct-password") {
		t.Fatal("response contains password")
	}
}

func TestAdminAuthRejectsUnauthorizedMe(t *testing.T) {
	service, _ := newHandlerTestService(t)
	handler := NewHandler(*service, "development")
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/auth/me", nil)

	handler.me(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Fatalf("body = %s, want unauthorized code", response.Body.String())
	}
}

func TestMeReturnsAuthenticatedAdmin(t *testing.T) {
	service, store := newHandlerTestService(t)
	store.sessionAdmin = Admin{ID: "admin-id", Email: "admin@example.com"}
	handler := NewHandler(*service, "development")
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-token"})
	response := httptest.NewRecorder()

	handler.me(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	var payload adminResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Admin != store.sessionAdmin {
		t.Fatalf("admin = %+v, want %+v", payload.Admin, store.sessionAdmin)
	}
	if !strings.Contains(response.Body.String(), `"id":"admin-id"`) || !strings.Contains(response.Body.String(), `"email":"admin@example.com"`) {
		t.Fatalf("body = %s, want lowercase JSON fields", response.Body.String())
	}
}

func TestLogoutIsIdempotentAndClearsCookie(t *testing.T) {
	service, store := newHandlerTestService(t)
	handler := NewHandler(*service, "production")
	request := httptest.NewRequest(http.MethodPost, "/v1/admin/auth/logout", nil)
	request.Header.Set("Origin", "http://localhost:3001")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-token"})
	request.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "csrf-token"})
	request.Header.Set("X-CSRF-Token", "csrf-token")
	response := httptest.NewRecorder()

	handler.logout(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if store.revokeSessionCall != 1 || store.revokedTokenHash == "session-token" || store.revokedTokenHash == "" {
		t.Fatal("logout did not revoke only the hashed session token")
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Value != "" || cookie.MaxAge != -1 || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected clearing cookie: %+v", cookie)
	}
}

func TestLogoutWithoutSessionIsIdempotent(t *testing.T) {
	service, store := newHandlerTestService(t)
	handler := NewHandler(*service, "development")
	response := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodPost, "/v1/admin/auth/logout", nil)
	request.Header.Set("Origin", "http://localhost:3001")
	request.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "csrf-token"})
	request.Header.Set("X-CSRF-Token", "csrf-token")
	handler.logout(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if store.revokeSessionCall != 0 {
		t.Fatalf("RevokeSession called %d times, want 0", store.revokeSessionCall)
	}
}

func TestAdminAuthMethodAndValidationErrors(t *testing.T) {
	service, _ := newHandlerTestService(t)
	handler := NewHandler(*service, "development")

	tests := []struct {
		name string
		call func(http.ResponseWriter, *http.Request)
		req  *http.Request
		want int
	}{
		{"login method", handler.login, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusMethodNotAllowed},
		{"logout method", handler.logout, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusMethodNotAllowed},
		{"me method", handler.me, httptest.NewRequest(http.MethodPost, "/", nil), http.StatusMethodNotAllowed},
		{"login validation", handler.login, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"admin@example.com"}`)), http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.Header.Set("Origin", "http://localhost:3001")
			response := httptest.NewRecorder()
			tt.call(response, tt.req)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d", response.Code, tt.want)
			}
		})
	}
}

func TestAdminAuthRejectsUntrustedOrigin(t *testing.T) {
	service, _ := newHandlerTestService(t)
	handler := NewHandler(*service, "development")
	request := httptest.NewRequest(http.MethodPost, "/v1/admin/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"correct-password"}`))
	request.Header.Set("Origin", "https://evil.example")
	response := httptest.NewRecorder()

	handler.login(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestLoginDoesNotRevealWhetherAdminExists(t *testing.T) {
	tests := []struct {
		name      string
		findError error
		password  string
	}{
		{name: "unknown admin", findError: ErrAdminNotFound, password: "correct-password"},
		{name: "wrong password", password: "wrong-password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, store := newHandlerTestService(t)
			store.findErr = tt.findError
			handler := NewHandler(*service, "development")
			request := httptest.NewRequest(http.MethodPost, "/v1/admin/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"`+tt.password+`"}`))
			request.Header.Set("Origin", "http://localhost:3001")
			response := httptest.NewRecorder()

			handler.login(response, request)

			if response.Code != http.StatusUnauthorized || response.Body.String() != `{"code":"invalid_credentials","message":"E-posta veya parola hatalı."}
` {
				t.Fatalf("status/body = %d/%q, want common invalid credentials response", response.Code, response.Body.String())
			}
		})
	}
}

func TestMeRejectsExpiredAndRevokedSessions(t *testing.T) {
	for _, name := range []string{"expired session", "revoked session"} {
		t.Run(name, func(t *testing.T) {
			service, store := newHandlerTestService(t)
			store.sessionErr = ErrSessionNotFound
			handler := NewHandler(*service, "development")
			request := httptest.NewRequest(http.MethodGet, "/v1/admin/auth/me", nil)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "old-session-token"})
			response := httptest.NewRecorder()

			handler.me(response, request)

			if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
				t.Fatalf("status/body = %d/%s, want unauthorized", response.Code, response.Body.String())
			}
		})
	}
}

func TestLogoutRequiresMatchingCSRFToken(t *testing.T) {
	service, _ := newHandlerTestService(t)
	handler := NewHandler(*service, "development")
	request := httptest.NewRequest(http.MethodPost, "/v1/admin/auth/logout", nil)
	request.Header.Set("Origin", "http://localhost:3001")
	request.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "cookie-token"})
	request.Header.Set("X-CSRF-Token", "different-token")
	response := httptest.NewRecorder()

	handler.logout(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}
