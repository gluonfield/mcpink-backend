-- +goose Up
CREATE TABLE internal_repos (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    provider TEXT NOT NULL DEFAULT 'gitea',
    repo_id BIGINT NOT NULL,
    full_name TEXT NOT NULL UNIQUE,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_internal_repos_user_id ON internal_repos(user_id);
CREATE INDEX idx_internal_repos_full_name ON internal_repos(full_name);

ALTER TABLE users ADD COLUMN gitea_username TEXT;

-- +goose Down
DROP TABLE IF EXISTS internal_repos;
ALTER TABLE users DROP COLUMN IF EXISTS gitea_username;
