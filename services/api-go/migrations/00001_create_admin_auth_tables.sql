-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE admin_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT admin_users_email_not_blank CHECK (btrim(email) <> ''),
  CONSTRAINT admin_users_email_normalized CHECK (email = lower(btrim(email))),
  CONSTRAINT admin_users_password_hash_not_blank CHECK (btrim(password_hash) <> '')
);

CREATE TABLE admin_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  admin_user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at TIMESTAMPTZ,
  CONSTRAINT admin_sessions_token_hash_not_blank CHECK (btrim(token_hash) <> ''),
  CONSTRAINT admin_sessions_expires_after_created CHECK (expires_at > created_at)
);

CREATE INDEX admin_sessions_admin_user_id_idx ON admin_sessions(admin_user_id);
CREATE INDEX admin_sessions_active_lookup_idx ON admin_sessions(token_hash, expires_at)
  WHERE revoked_at IS NULL;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION capa_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER admin_users_set_updated_at
BEFORE UPDATE ON admin_users
FOR EACH ROW
EXECUTE FUNCTION capa_set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS admin_users_set_updated_at ON admin_users;
DROP FUNCTION IF EXISTS capa_set_updated_at();
DROP TABLE IF EXISTS admin_sessions;
DROP TABLE IF EXISTS admin_users;
DROP EXTENSION IF EXISTS pgcrypto;
