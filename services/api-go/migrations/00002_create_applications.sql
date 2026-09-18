-- +goose Up
CREATE TABLE applications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  position_code TEXT NOT NULL,
  full_name TEXT NOT NULL,
  email_snapshot TEXT NOT NULL,
  experience TEXT NOT NULL,
  submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT applications_position_code_allowed CHECK (position_code IN ('frontend', 'backend', 'fullstack')),
  CONSTRAINT applications_full_name_length CHECK (char_length(full_name) BETWEEN 2 AND 120),
  CONSTRAINT applications_email_not_blank CHECK (btrim(email_snapshot) <> ''),
  CONSTRAINT applications_email_normalized CHECK (email_snapshot = lower(btrim(email_snapshot))),
  CONSTRAINT applications_experience_length CHECK (char_length(experience) BETWEEN 20 AND 5000),
  CONSTRAINT applications_email_position_key UNIQUE (email_snapshot, position_code)
);

CREATE INDEX applications_submitted_at_id_idx ON applications(submitted_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS applications;
