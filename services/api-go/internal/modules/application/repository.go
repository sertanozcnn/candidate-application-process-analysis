package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return Repository{db: db}
}

func (r Repository) Create(ctx context.Context, input CreateInput) (Application, error) {
	var item Application
	err := r.db.QueryRow(ctx, `
		INSERT INTO applications (position_code, full_name, email_snapshot, experience)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, position_code, full_name, email_snapshot, experience, submitted_at
	`, input.PositionCode, input.FullName, input.Email, input.Experience).Scan(
		&item.ID, &item.PositionCode, &item.FullName, &item.Email, &item.Experience, &item.SubmittedAt,
	)
	if err != nil {
		if isApplicationDuplicate(err) {
			return Application{}, ErrApplicationExists
		}
		return Application{}, fmt.Errorf("create application: %w", err)
	}
	return item, nil
}

func (r Repository) List(ctx context.Context, limit, offset int) ([]Application, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, position_code, full_name, email_snapshot, experience, submitted_at
		FROM applications
		ORDER BY submitted_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	items := make([]Application, 0)
	for rows.Next() {
		var item Application
		if err := rows.Scan(&item.ID, &item.PositionCode, &item.FullName, &item.Email, &item.Experience, &item.SubmittedAt); err != nil {
			return nil, fmt.Errorf("scan application: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applications: %w", err)
	}
	return items, nil
}

func (r Repository) FindByID(ctx context.Context, id string) (Application, error) {
	var item Application
	err := r.db.QueryRow(ctx, `
		SELECT id::text, position_code, full_name, email_snapshot, experience, submitted_at
		FROM applications
		WHERE id = $1
	`, id).Scan(&item.ID, &item.PositionCode, &item.FullName, &item.Email, &item.Experience, &item.SubmittedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, ErrApplicationNotFound
	}
	if err != nil {
		return Application{}, fmt.Errorf("find application: %w", err)
	}
	return item, nil
}

func isApplicationDuplicate(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "applications_email_position_key"
}
