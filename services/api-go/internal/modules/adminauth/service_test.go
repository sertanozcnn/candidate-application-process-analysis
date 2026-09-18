package adminauth

import (
	"context"
	"errors"
	"testing"
	"time"

	"capa/services/api-go/internal/platform/password"
)

type fakeStore struct {
	credentials  adminCredentials
	findErr      error
	sessionAdmin Admin
	sessionErr   error

	createdAdminID    string
	createdTokenHash  string
	createdExpiresAt  time.Time
	createSessionErr  error
	createSessionCall int
	revokedTokenHash  string
	revokeSessionCall int
}

func (f *fakeStore) FindCredentialsByEmail(context.Context, string) (adminCredentials, error) {
	if f.findErr != nil {
		return adminCredentials{}, f.findErr
	}

	return f.credentials, nil
}

func (f *fakeStore) CreateSession(_ context.Context, adminID string, tokenHash string, expiresAt time.Time) error {
	f.createSessionCall++
	f.createdAdminID = adminID
	f.createdTokenHash = tokenHash
	f.createdExpiresAt = expiresAt
	return f.createSessionErr
}

func (f *fakeStore) FindActiveSession(context.Context, string, time.Time) (Admin, error) {
	if f.sessionErr != nil {
		return Admin{}, f.sessionErr
	}
	return f.sessionAdmin, nil
}

func (f *fakeStore) RevokeSession(_ context.Context, tokenHash string, _ time.Time) error {
	f.revokeSessionCall++
	f.revokedTokenHash = tokenHash
	return nil
}

func TestLoginCreatesSession(t *testing.T) {
	passwordHash, err := password.Hash("local-admin-password")
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}

	store := &fakeStore{
		credentials: adminCredentials{
			Admin: Admin{
				ID:    "admin-id",
				Email: "admin@example.com",
			},
			PasswordHash: passwordHash,
		},
	}

	svc := NewService(store)
	now := time.Date(2026, 9, 11, 16, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	result, err := svc.Login(context.Background(), " ADMIN@EXAMPLE.COM ", "local-admin-password")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if result.Admin.Email != "admin@example.com" {
		t.Fatalf("Admin.Email = %q, want admin@example.com", result.Admin.Email)
	}
	if result.SessionToken == "" {
		t.Fatal("Login returned empty session token")
	}
	if !result.ExpiresAt.Equal(now.Add(defaultSessionTTL)) {
		t.Fatalf("ExpiresAt = %v, want %v", result.ExpiresAt, now.Add(defaultSessionTTL))
	}
	if store.createSessionCall != 1 {
		t.Fatalf("CreateSession called %d times, want 1", store.createSessionCall)
	}
	if store.createdAdminID != "admin-id" {
		t.Fatalf("created admin id = %q, want admin-id", store.createdAdminID)
	}
	if store.createdTokenHash == "" || store.createdTokenHash == result.SessionToken {
		t.Fatal("CreateSession did not receive a hashed token")
	}
}

func TestLoginRejectsUnknownAdmin(t *testing.T) {
	svc := NewService(&fakeStore{findErr: ErrAdminNotFound})

	_, err := svc.Login(context.Background(), "missing@example.com", "local-admin-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	passwordHash, err := password.Hash("local-admin-password")
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}

	store := &fakeStore{
		credentials: adminCredentials{
			Admin:        Admin{ID: "admin-id", Email: "admin@example.com"},
			PasswordHash: passwordHash,
		},
	}
	svc := NewService(store)

	_, err = svc.Login(context.Background(), "admin@example.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want ErrInvalidCredentials", err)
	}
	if store.createSessionCall != 0 {
		t.Fatalf("CreateSession called %d times, want 0", store.createSessionCall)
	}
}
