-- +goose Up

-- Track if user has granted repo scope
ALTER TABLE users ADD COLUMN has_repo_scope BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS has_repo_scope;
