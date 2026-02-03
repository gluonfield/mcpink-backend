-- +goose Up
ALTER TABLE apps ADD COLUMN git_provider TEXT NOT NULL DEFAULT 'github';

-- +goose Down
ALTER TABLE apps DROP COLUMN IF EXISTS git_provider;
