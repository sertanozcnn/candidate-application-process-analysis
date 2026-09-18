package adminauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"capa/services/api-go/internal/modules/adminuser"
	"capa/services/api-go/internal/platform/password"
	"capa/services/api-go/internal/platform/sessiontoken"
)

const defaultSessionTTL = 8 * time.Hour

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUnauthorized = errors.New("unauthorized")

type store interface {
	FindCredentialsByEmail(ctx context.Context, email string) (adminCredentials, error)
	CreateSession(ctx context.Context, adminID string, tokenHash string, expiresAt time.Time) error
	FindActiveSession(ctx context.Context, tokenHash string, now time.Time) (Admin, error)
	RevokeSession(ctx context.Context, tokenHash string, now time.Time) error
}

type Service struct {
	store      store
	sessionTTL time.Duration
	now        func() time.Time
}

func NewService(store store) Service {
	return Service{
		store:      store,
		sessionTTL: defaultSessionTTL,
		now:        time.Now,
	}
}

func (s Service) Login(ctx context.Context, email string, plainPassword string) (LoginResult, error) {
	normalizedEmail := adminuser.NormalizeEmail(email)
	if normalizedEmail == "" || plainPassword == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	credentials, err := s.store.FindCredentialsByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, ErrAdminNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}

		return LoginResult{}, err
	}

	matched, err := password.Verify(plainPassword, credentials.PasswordHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("verify admin password: %w", err)
	}
	if !matched {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, err := sessiontoken.Generate()
	if err != nil {
		return LoginResult{}, err
	}

	expiresAt := s.now().Add(s.sessionTTL)
	if err := s.store.CreateSession(ctx, credentials.ID, sessiontoken.Hash(token), expiresAt); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		Admin:        credentials.Admin,
		SessionToken: token,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s Service) Authenticate(ctx context.Context, token string) (Admin, error) {
	if token == "" {
		return Admin{}, ErrUnauthorized
	}

	admin, err := s.store.FindActiveSession(ctx, sessiontoken.Hash(token), s.now())
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return Admin{}, ErrUnauthorized
		}
		return Admin{}, err
	}

	return admin, nil
}

func (s Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	return s.store.RevokeSession(ctx, sessiontoken.Hash(token), s.now())
}
