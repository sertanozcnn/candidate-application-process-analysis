package interaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) Repository { return Repository{db: db} }

func (r Repository) CreateSession(ctx context.Context, positionCode, tokenHash string) (Session, error) {
	var session Session
	err := r.db.QueryRow(ctx, `INSERT INTO interaction_sessions (position_code, session_token_hash) VALUES ($1, $2) RETURNING id::text, position_code`, positionCode, tokenHash).Scan(&session.ID, &session.PositionCode)
	if err != nil {
		return Session{}, fmt.Errorf("create interaction session: %w", err)
	}
	return session, nil
}

func (r Repository) AppendEvents(ctx context.Context, sessionID, tokenHash string, events []Event) (BatchResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return BatchResult{}, fmt.Errorf("begin interaction batch: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	var lastReceivedAt time.Time
	err = tx.QueryRow(ctx, `SELECT status, last_received_at FROM interaction_sessions WHERE id = $1 AND session_token_hash = $2 FOR UPDATE`, sessionID, tokenHash).Scan(&status, &lastReceivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return BatchResult{}, ErrSessionNotFound
	}
	if err != nil {
		return BatchResult{}, fmt.Errorf("lock interaction session: %w", err)
	}
	if status != "open" {
		return BatchResult{}, ErrSessionExpired
	}
	if time.Since(lastReceivedAt) > 30*time.Minute {
		if _, err := tx.Exec(ctx, `UPDATE interaction_sessions SET status = 'expired' WHERE id = $1`, sessionID); err != nil {
			return BatchResult{}, fmt.Errorf("expire interaction session: %w", err)
		}
		return BatchResult{}, ErrSessionExpired
	}
	var lastSequence int
	err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(sequence), 0) FROM interaction_events WHERE session_id = $1`, sessionID).Scan(&lastSequence)
	if errors.Is(err, pgx.ErrNoRows) {
		return BatchResult{}, ErrSessionNotFound
	}
	if err != nil {
		return BatchResult{}, fmt.Errorf("lock interaction session: %w", err)
	}

	result := BatchResult{}
	for _, event := range events {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM interaction_events WHERE session_id = $1 AND event_id = $2)`, sessionID, event.EventID).Scan(&exists); err != nil {
			return BatchResult{}, fmt.Errorf("check interaction event: %w", err)
		}
		if exists {
			result.Duplicate++
			continue
		}
		if event.Sequence <= lastSequence {
			return BatchResult{}, ErrSequenceOrder
		}
		metadata, err := marshalMetadata(event.Metadata)
		if err != nil {
			return BatchResult{}, fmt.Errorf("marshal interaction metadata: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO interaction_events (session_id, event_id, sequence, elapsed_ms, type, field_code, metadata) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7)`, sessionID, event.EventID, event.Sequence, event.ElapsedMS, event.Type, event.FieldCode, metadata); err != nil {
			return BatchResult{}, fmt.Errorf("insert interaction event: %w", err)
		}
		lastSequence = event.Sequence
		result.Accepted++
	}
	if result.Accepted > 0 {
		if _, err := tx.Exec(ctx, `UPDATE interaction_sessions SET last_received_at = now() WHERE id = $1`, sessionID); err != nil {
			return BatchResult{}, fmt.Errorf("update interaction session: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return BatchResult{}, fmt.Errorf("commit interaction batch: %w", err)
	}
	return result, nil
}

func (r Repository) AttachApplication(ctx context.Context, sessionID, tokenHash, applicationID, positionCode string) error {
	var expired bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM interaction_sessions WHERE id = $1 AND session_token_hash = $2 AND status = 'open' AND last_received_at < now() - interval '30 minutes')`, sessionID, tokenHash).Scan(&expired); err != nil {
		return fmt.Errorf("check interaction session age: %w", err)
	}
	if expired {
		if _, err := r.db.Exec(ctx, `UPDATE interaction_sessions SET status = 'expired' WHERE id = $1 AND session_token_hash = $2 AND status = 'open'`, sessionID, tokenHash); err != nil {
			return fmt.Errorf("expire interaction session: %w", err)
		}
		return ErrSessionExpired
	}
	var linkedID string
	err := r.db.QueryRow(ctx, `UPDATE interaction_sessions SET application_id = $1, last_received_at = now() WHERE id = $2 AND session_token_hash = $3 AND position_code = $4 AND status = 'open' AND application_id IS NULL RETURNING id::text`, applicationID, sessionID, tokenHash, positionCode).Scan(&linkedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrSessionNotFound
	}
	if err != nil {
		return fmt.Errorf("attach application to interaction session: %w", err)
	}
	return nil
}
