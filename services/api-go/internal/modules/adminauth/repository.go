package adminauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"capa/services/api-go/internal/modules/adminuser"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAdminNotFound = errors.New("admin not found")
var ErrSessionNotFound = errors.New("session not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return Repository{db: db}
}

func (r Repository) FindCredentialsByEmail(ctx context.Context, email string) (adminCredentials, error) {
	normalizedEmail := adminuser.NormalizeEmail(email)

	var credentials adminCredentials
	err := r.db.QueryRow(ctx, `
		SELECT id::text, email, password_hash
		FROM admin_users
		WHERE email = $1
	`, normalizedEmail).Scan(&credentials.ID, &credentials.Email, &credentials.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return adminCredentials{}, ErrAdminNotFound
		}

		return adminCredentials{}, fmt.Errorf("find admin credentials: %w", err)
	}

	return credentials, nil
}

func (r Repository) CreateSession(ctx context.Context, adminID string, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO admin_sessions (admin_user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, adminID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("create admin session: %w", err)
	}

	return nil
}

func (r Repository) FindActiveSession(ctx context.Context, tokenHash string, now time.Time) (Admin, error) {
	var admin Admin
	err := r.db.QueryRow(ctx, `
		UPDATE admin_sessions AS s
		SET last_seen_at = $2
		FROM admin_users AS u
		WHERE s.token_hash = $1
		  AND s.admin_user_id = u.id
		  AND s.revoked_at IS NULL
		  AND s.expires_at > $2
		RETURNING u.id::text, u.email
	`, tokenHash, now).Scan(&admin.ID, &admin.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Admin{}, ErrSessionNotFound
		}
		return Admin{}, fmt.Errorf("find active admin session: %w", err)
	}

	return admin, nil
}

func (r Repository) RevokeSession(ctx context.Context, tokenHash string, now time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE admin_sessions
		SET revoked_at = $2
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash, now)
	if err != nil {
		return fmt.Errorf("revoke admin session: %w", err)
	}

	return nil
}
