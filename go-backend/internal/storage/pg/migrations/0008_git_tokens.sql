-- +goose Up

CREATE TABLE git_tokens (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    token_hash TEXT NOT NULL,
    token_prefix TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    repo_id TEXT REFERENCES internal_repos(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL DEFAULT '{push,pull}',
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_git_tokens_hash ON git_tokens(token_hash);
CREATE INDEX idx_git_tokens_user_id ON git_tokens(user_id);
CREATE INDEX idx_git_tokens_repo_id ON git_tokens(repo_id);

ALTER TABLE internal_repos ADD COLUMN bare_path TEXT;

UPDATE services SET git_provider = 'internal' WHERE git_provider = 'gitea';
UPDATE internal_repos SET provider = 'internal' WHERE provider = 'gitea';

-- +goose Down

DROP TABLE IF EXISTS git_tokens;
ALTER TABLE internal_repos DROP COLUMN IF EXISTS bare_path;
UPDATE services SET git_provider = 'gitea' WHERE git_provider = 'internal';
UPDATE internal_repos SET provider = 'gitea' WHERE provider = 'internal';
