package interaction

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct {
	session       Session
	createErr     error
	appendResult  BatchResult
	appendErr     error
	lastSessionID string
	lastTokenHash string
	lastEvents    []Event
}

func (f *fakeStore) CreateSession(_ context.Context, positionCode, tokenHash string) (Session, error) {
	f.lastTokenHash = tokenHash
	if f.createErr != nil {
		return Session{}, f.createErr
	}
	if f.session.ID == "" {
		f.session = Session{ID: "123e4567-e89b-12d3-a456-426614174000", PositionCode: positionCode}
	}
	return f.session, nil
}

func (f *fakeStore) AppendEvents(_ context.Context, sessionID, tokenHash string, events []Event) (BatchResult, error) {
	f.lastSessionID = sessionID
	f.lastTokenHash = tokenHash
	f.lastEvents = events
	return f.appendResult, f.appendErr
}

func (f *fakeStore) AttachApplication(context.Context, string, string, string, string) error {
	return nil
}

func TestCreateSessionSetsHttpOnlyCookie(t *testing.T) {
	store := &fakeStore{}
	handler := Handler{service: NewService(store)}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/interaction-sessions", strings.NewReader(`{"position_code":" BACKEND "}`))
	response := httptest.NewRecorder()

	handler.createSession(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", response.Code)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("cookies = %d, want session cookie plus legacy path cleanup", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != sessionCookieName || cookie.Value == "" || !cookie.HttpOnly || cookie.Path != "/v1/candidate" {
		t.Fatalf("cookie = %+v, want scoped HttpOnly interaction cookie", cookie)
	}
	if cookies[1].Path != "/v1/candidate/interaction-sessions" || cookies[1].MaxAge != -1 {
		t.Fatalf("legacy cleanup cookie = %+v, want old path cleared", cookies[1])
	}
	if store.lastTokenHash == "" || store.lastTokenHash == cookie.Value {
		t.Fatal("session token must be stored as a hash")
	}
}

func TestEventsRequireSessionCookie(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{})}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/interaction-sessions/123e4567-e89b-12d3-a456-426614174000/events", strings.NewReader(`{"events":[]}`))
	response := httptest.NewRecorder()

	handler.events(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestEventsRejectUnknownMetadata(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{})}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/interaction-sessions/123e4567-e89b-12d3-a456-426614174000/events", strings.NewReader(`{"events":[{"event_id":"123e4567-e89b-12d3-a456-426614174001","sequence":1,"elapsed_ms":10,"type":"field_focus","field_code":"email","metadata":{"value":"secret"}}]}`))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "synthetic-session-token"})
	response := httptest.NewRecorder()

	handler.events(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestServiceRejectsOversizedBatch(t *testing.T) {
	events := make([]Event, 51)
	for index := range events {
		events[index] = Event{EventID: "123e4567-e89b-12d3-a456-426614174001", Sequence: index + 1, Type: "session_start", Metadata: map[string]any{}}
	}

	_, err := NewService(&fakeStore{}).AppendEvents(context.Background(), "session", "token", events)
	if err != ErrInvalidBatch {
		t.Fatalf("error = %v, want invalid batch", err)
	}
}

func TestEventsMapsSequenceConflict(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{appendErr: ErrSequenceOrder})}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/interaction-sessions/123e4567-e89b-12d3-a456-426614174000/events", strings.NewReader(`{"events":[{"event_id":"123e4567-e89b-12d3-a456-426614174001","sequence":1,"elapsed_ms":10,"type":"session_start","metadata":{}}]}`))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "synthetic-session-token"})
	response := httptest.NewRecorder()

	handler.events(response, request)

	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"code":"sequence_conflict"`) {
		t.Fatalf("status/body = %d/%s, want sequence conflict", response.Code, response.Body.String())
	}
}

func TestEventsMapsExpiredSession(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{appendErr: ErrSessionExpired})}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/interaction-sessions/123e4567-e89b-12d3-a456-426614174000/events", strings.NewReader(`{"events":[{"event_id":"123e4567-e89b-12d3-a456-426614174001","sequence":1,"elapsed_ms":10,"type":"session_start","metadata":{}}]}`))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "synthetic-session-token"})
	response := httptest.NewRecorder()

	handler.events(response, request)

	if response.Code != http.StatusGone || !strings.Contains(response.Body.String(), `"code":"session_expired"`) {
		t.Fatalf("status/body = %d/%s, want expired session", response.Code, response.Body.String())
	}
}

func TestEventsResponseDoesNotEchoPayload(t *testing.T) {
	handler := Handler{service: NewService(&fakeStore{appendResult: BatchResult{Accepted: 1}})}
	request := httptest.NewRequest(http.MethodPost, "/v1/candidate/interaction-sessions/123e4567-e89b-12d3-a456-426614174000/events", strings.NewReader(`{"events":[{"event_id":"123e4567-e89b-12d3-a456-426614174001","sequence":1,"elapsed_ms":10,"type":"field_focus","field_code":"experience","metadata":{}}]}`))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "synthetic-session-token"})
	response := httptest.NewRecorder()

	handler.events(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", response.Code)
	}
	body := response.Body.String()
	if strings.Contains(body, "experience") || strings.Contains(body, "event_id") || strings.Contains(body, "field_focus") {
		t.Fatalf("body = %s, must not echo event payload", body)
	}
}
