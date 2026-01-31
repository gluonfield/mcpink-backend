-- +goose Up
-- Store actual OAuth scopes from GitHub (space-separated string like "read:user user:email repo")
ALTER TABLE users ADD COLUMN github_scopes TEXT NOT NULL DEFAULT '';
-- Drop unused boolean column from previous migration
ALTER TABLE users DROP COLUMN IF EXISTS has_repo_scope;

-- +goose Down
ALTER TABLE users ADD COLUMN has_repo_scope BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users DROP COLUMN IF EXISTS github_scopes;
