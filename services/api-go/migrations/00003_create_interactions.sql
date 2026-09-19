-- +goose Up
CREATE TABLE interaction_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  position_code TEXT NOT NULL,
  session_token_hash TEXT NOT NULL UNIQUE,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  application_id UUID REFERENCES applications(id) ON DELETE SET NULL,
  status TEXT NOT NULL DEFAULT 'open',
  data_quality_flags JSONB NOT NULL DEFAULT '[]'::jsonb,
  CONSTRAINT interaction_sessions_position_allowed CHECK (position_code IN ('frontend', 'backend', 'fullstack')),
  CONSTRAINT interaction_sessions_status_allowed CHECK (status IN ('open', 'closed', 'expired'))
);

CREATE TABLE interaction_events (
  id BIGSERIAL PRIMARY KEY,
  session_id UUID NOT NULL REFERENCES interaction_sessions(id) ON DELETE CASCADE,
  event_id UUID NOT NULL,
  sequence INTEGER NOT NULL,
  elapsed_ms BIGINT NOT NULL,
  type TEXT NOT NULL,
  field_code TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT interaction_events_sequence_positive CHECK (sequence > 0),
  CONSTRAINT interaction_events_elapsed_nonnegative CHECK (elapsed_ms >= 0),
  CONSTRAINT interaction_events_event_key UNIQUE (session_id, event_id),
  CONSTRAINT interaction_events_sequence_key UNIQUE (session_id, sequence)
);

CREATE INDEX interaction_sessions_application_idx ON interaction_sessions(application_id);
CREATE INDEX interaction_events_session_sequence_idx ON interaction_events(session_id, sequence);

-- +goose Down
DROP TABLE IF EXISTS interaction_events;
DROP TABLE IF EXISTS interaction_sessions;
