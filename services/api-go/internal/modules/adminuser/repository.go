package adminuser

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAlreadyExists = errors.New("admin user already exists")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return Repository{db: db}
}

func (r Repository) Create(ctx context.Context, email string, passwordHash string) error {
	normalizedEmail := NormalizeEmail(email)

	_, err := r.db.Exec(ctx, `
		INSERT INTO admin_users (email, password_hash)
		VALUES ($1, $2)
	`, normalizedEmail, passwordHash)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}

		return fmt.Errorf("insert admin user: %w", err)
	}

	return nil
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
