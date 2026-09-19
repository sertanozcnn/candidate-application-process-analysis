package application

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"capa/services/api-go/internal/modules/adminauth"
)

type fakeStore struct {
	created    CreateInput
	createErr  error
	items      []Application
	listErr    error
	listLimit  int
	listOffset int
	found      Application
	findErr    error
}

func (f *fakeStore) Create(_ context.Context, input CreateInput) (Application, error) {
	f.created = input
	if f.createErr != nil {
		return Application{}, f.createErr
	}
	return Application{ID: "application-id", PositionCode: input.PositionCode, FullName: input.FullName, Email: input.Email, Experience: input.Experience, SubmittedAt: time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)}, nil
}

func (f *fakeStore) List(_ context.Context, limit, offset int) ([]Application, error) {
	f.listLimit = limit
	f.listOffset = offset
	return f.items, f.listErr
}

func (f *fakeStore) FindByID(context.Context, string) (Application, error) {
	return f.found, f.findErr
}

type fakeAuth struct {
	err error
}

type fakeLinker struct {
	sessionID     string
	token         string
	applicationID string
	positionCode  string
}

func (f *fakeLinker) AttachApplication(_ context.Context, sessionID, token, applicationID, positionCode string) error {
	f.sessionID = sessionID
	f.token = token
	f.applicationID = applicationID
	f.positionCode = positionCode
	return nil
}

func (f fakeAuth) Authenticate(context.Context, string) (adminauth.Admin, error) {
	if f.err != nil {
		return adminauth.Admin{}, f.err
	}
	return adminauth.Admin{ID: "admin-id", Email: "admin@example.com"}, nil
}

func TestCreateNormalizesAndReturnsPublicResult(t *testing.T) {
	store := &fakeStore{}
	handler := Handler{service: NewService(store), auth: fakeAuth{}}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/applications", strings.NewReader(`{"full_name":"  Ada   Lovelace ","email":" ADA@EXAMPLE.COM ","position_code":" BACKEND ","experience":"This is a sufficiently long experience description."}`))
	response := httptest.NewRecorder()

	handler.create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	if store.created.FullName != "Ada Lovelace" || store.created.Email != "ada@example.com" || store.created.PositionCode != "backend" {
		t.Fatalf("created input = %+v, want normalized values", store.created)
	}
	var payload createResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ID != "application-id" || payload.Email != "ada@example.com" || payload.PositionCode != "backend" {
		t.Fatalf("payload = %+v, want public application result", payload)
	}
	if strings.Contains(response.Body.String(), "experience") || strings.Contains(response.Body.String(), "Ada Lovelace") {
		t.Fatal("create response exposed unnecessary application data")
	}
}

func TestCreateRejectsInvalidPosition(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{}), auth: fakeAuth{}}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/applications", strings.NewReader(`{"full_name":"Ada Lovelace","email":"ada@example.com","position_code":"designer","experience":"This is a sufficiently long experience description."}`))
	response := httptest.NewRecorder()

	handler.create(response, request)

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("status/body = %d/%s, want invalid request", response.Code, response.Body.String())
	}
}

func TestCreateMapsDuplicateApplication(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{createErr: ErrApplicationExists}), auth: fakeAuth{}}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/applications", strings.NewReader(`{"full_name":"Ada Lovelace","email":"ada@example.com","position_code":"backend","experience":"This is a sufficiently long experience description."}`))
	response := httptest.NewRecorder()

	handler.create(response, request)

	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"application_exists"`) {
		t.Fatalf("status/body = %d/%s, want application exists", response.Code, response.Body.String())
	}
}

func TestCreateMapsInternalError(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{createErr: errors.New("database unavailable")}), auth: fakeAuth{}}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/applications", strings.NewReader(`{"full_name":"Ada Lovelace","email":"ada@example.com","position_code":"backend","experience":"This is a sufficiently long experience description."}`))
	response := httptest.NewRecorder()

	handler.create(response, request)

	if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("status/body = %d/%s, want internal error", response.Code, response.Body.String())
	}
}

func TestCreateAttemptsToAttachInteractionSession(t *testing.T) {
	linker := &fakeLinker{}
	handler := Handler{service: NewService(&fakeStore{}), auth: fakeAuth{}, linker: linker}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/applications", strings.NewReader(`{"full_name":"Ada Lovelace","email":"ada@example.com","position_code":"backend","experience":"This is a sufficiently long experience description."}`))
	request.AddCookie(&http.Cookie{Name: "capa_interaction_session", Value: "session-token"})
	request.Header.Set("X-Interaction-Session-ID", "123e4567-e89b-12d3-a456-426614174000")
	response := httptest.NewRecorder()

	handler.create(response, request)

	if response.Code != http.StatusCreated || linker.sessionID != "123e4567-e89b-12d3-a456-426614174000" || linker.token != "session-token" || linker.applicationID != "application-id" || linker.positionCode != "backend" {
		t.Fatalf("status/link = %d/%+v, want created and interaction link", response.Code, linker)
	}
}

func TestAdminApplicationsRequireAuthentication(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{}), auth: fakeAuth{}}
	response := httptest.NewRecorder()

	handler.list(response, httptest.NewRequest(http.MethodGet, "/v1/admin/applications", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestAdminListReturnsAuthenticatedApplications(t *testing.T) {
	store := &fakeStore{items: []Application{{ID: "application-id", PositionCode: "backend", Email: "ada@example.com"}}}
	handler := Handler{service: NewService(store), auth: fakeAuth{}}
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/applications?limit=10&offset=0", nil)
	request.AddCookie(&http.Cookie{Name: "capa_admin_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.list(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"applications"`) || !strings.Contains(response.Body.String(), "application-id") {
		t.Fatalf("body = %s, want application list", response.Body.String())
	}
}

func TestAdminListPassesPagination(t *testing.T) {
	store := &fakeStore{}
	handler := Handler{service: NewService(store), auth: fakeAuth{}}
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/applications?limit=10&offset=20", nil)
	request.AddCookie(&http.Cookie{Name: "capa_admin_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.list(response, request)

	if response.Code != http.StatusOK || store.listLimit != 10 || store.listOffset != 20 {
		t.Fatalf("status/pagination = %d/%d/%d, want 200/10/20", response.Code, store.listLimit, store.listOffset)
	}
}

func TestAdminDetailReturnsAuthenticatedApplication(t *testing.T) {
	store := &fakeStore{found: Application{ID: "123e4567-e89b-12d3-a456-426614174000", FullName: "Ada Lovelace", Email: "ada@example.com", PositionCode: "backend", Experience: "A sufficiently long experience description.", SubmittedAt: time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)}}
	handler := Handler{service: NewService(store), auth: fakeAuth{}}
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/applications/123e4567-e89b-12d3-a456-426614174000", nil)
	request.AddCookie(&http.Cookie{Name: "capa_admin_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.detail(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"application"`) || !strings.Contains(response.Body.String(), "Ada Lovelace") {
		t.Fatalf("status/body = %d/%s, want authenticated application detail", response.Code, response.Body.String())
	}
}

func TestAdminDetailRejectsUntrustedIDFormat(t *testing.T) {
	store := &fakeStore{}
	handler := Handler{service: NewService(store), auth: fakeAuth{}}
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/applications/not-a-uuid", nil)
	request.AddCookie(&http.Cookie{Name: "capa_admin_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.detail(response, request)

	if response.Code != http.StatusNotFound || store.found.ID != "" {
		t.Fatalf("status = %d, want controlled 404 without lookup", response.Code)
	}
}

func TestAdminDetailReturnsNotFound(t *testing.T) {
	service := NewService(&fakeStore{findErr: ErrApplicationNotFound})
	handler := Handler{service: service, auth: fakeAuth{}}
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/applications/123e4567-e89b-12d3-a456-426614174000", nil)
	request.AddCookie(&http.Cookie{Name: "capa_admin_session", Value: "session-token"})
	response := httptest.NewRecorder()

	handler.detail(response, request)

	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"not_found"`) {
		t.Fatalf("status/body = %d/%s, want not found", response.Code, response.Body.String())
	}
}

func TestAdminRejectsInvalidSession(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{}), auth: fakeAuth{err: adminauth.ErrUnauthorized}}
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/applications", nil)
	request.AddCookie(&http.Cookie{Name: "capa_admin_session", Value: "revoked-token"})
	response := httptest.NewRecorder()

	handler.list(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestServiceRejectsInvalidApplication(t *testing.T) {
	_, err := NewService(&fakeStore{}).Create(context.Background(), CreateInput{FullName: "A", Email: "bad", PositionCode: "backend", Experience: "short"})
	if !errors.Is(err, ErrInvalidApplication) {
		t.Fatalf("error = %v, want invalid application", err)
	}
}
