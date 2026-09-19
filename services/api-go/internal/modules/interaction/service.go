package interaction

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidSession  = errors.New("invalid interaction session")
	ErrInvalidBatch    = errors.New("invalid interaction batch")
	ErrSessionNotFound = errors.New("interaction session not found")
	ErrSessionExpired  = errors.New("interaction session expired")
	ErrSequenceOrder   = errors.New("interaction sequence out of order")
)

var allowedPositions = map[string]struct{}{"frontend": {}, "backend": {}, "fullstack": {}}
var allowedFields = map[string]struct{}{"full_name": {}, "email": {}, "position_code": {}, "experience": {}}
var allowedTypes = map[string]struct{}{
	"session_start": {}, "field_focus": {}, "field_blur": {}, "field_paste": {},
	"validation_error": {}, "visibility_change": {}, "activity_state": {}, "session_end": {},
}
var eventIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type store interface {
	CreateSession(context.Context, string, string) (Session, error)
	AppendEvents(context.Context, string, string, []Event) (BatchResult, error)
	AttachApplication(context.Context, string, string, string, string) error
}

type Service struct{ store store }

func NewService(store store) Service { return Service{store: store} }

func (s Service) CreateSession(ctx context.Context, positionCode, tokenHash string) (Session, error) {
	positionCode = strings.ToLower(strings.TrimSpace(positionCode))
	if _, ok := allowedPositions[positionCode]; !ok || tokenHash == "" {
		return Session{}, ErrInvalidSession
	}
	return s.store.CreateSession(ctx, positionCode, tokenHash)
}

func (s Service) AppendEvents(ctx context.Context, sessionID, tokenHash string, events []Event) (BatchResult, error) {
	if sessionID == "" || tokenHash == "" || len(events) == 0 || len(events) > 50 {
		return BatchResult{}, ErrInvalidBatch
	}
	if err := validateEvents(events); err != nil {
		return BatchResult{}, err
	}
	return s.store.AppendEvents(ctx, sessionID, tokenHash, events)
}

func (s Service) AttachApplication(ctx context.Context, sessionID, token, applicationID, positionCode string) error {
	if sessionID == "" || token == "" || applicationID == "" || positionCode == "" {
		return ErrInvalidSession
	}
	return s.store.AttachApplication(ctx, sessionID, hashToken(token), applicationID, strings.ToLower(strings.TrimSpace(positionCode)))
}

func validateEvents(events []Event) error {
	previousSequence := 0
	for _, event := range events {
		if !eventIDPattern.MatchString(event.EventID) || event.Sequence <= 0 || event.Sequence <= previousSequence || event.ElapsedMS < 0 || event.ElapsedMS > 86_400_000 {
			return ErrInvalidBatch
		}
		if _, ok := allowedTypes[event.Type]; !ok {
			return ErrInvalidBatch
		}
		if err := validateField(event); err != nil {
			return err
		}
		if event.Metadata == nil {
			event.Metadata = map[string]any{}
		}
		if err := validateMetadata(event); err != nil {
			return err
		}
		previousSequence = event.Sequence
	}
	return nil
}

func validateField(event Event) error {
	_, hasField := allowedFields[event.FieldCode]
	fieldRequired := event.Type == "field_focus" || event.Type == "field_blur" || event.Type == "field_paste" || event.Type == "validation_error"
	if fieldRequired != hasField {
		return ErrInvalidBatch
	}
	return nil
}

func validateMetadata(event Event) error {
	allowed := map[string]struct{}{}
	switch event.Type {
	case "field_paste":
		allowed["count"] = struct{}{}
	case "visibility_change", "activity_state":
		allowed["state"] = struct{}{}
	}
	for key := range event.Metadata {
		if _, ok := allowed[key]; !ok {
			return ErrInvalidBatch
		}
	}
	if event.Type == "field_paste" {
		count, ok := event.Metadata["count"].(float64)
		if !ok || count != 1 {
			return ErrInvalidBatch
		}
	}
	if event.Type == "visibility_change" || event.Type == "activity_state" {
		state, ok := event.Metadata["state"].(string)
		if !ok || (event.Type == "visibility_change" && state != "visible" && state != "hidden") || (event.Type == "activity_state" && state != "active" && state != "idle") {
			return ErrInvalidBatch
		}
	}
	return nil
}

func marshalMetadata(metadata map[string]any) ([]byte, error) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	return json.Marshal(metadata)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
