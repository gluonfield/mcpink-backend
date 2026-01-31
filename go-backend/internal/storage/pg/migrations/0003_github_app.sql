-- +goose Up

-- Add GitHub App installation tracking
ALTER TABLE users ADD COLUMN github_app_installation_id BIGINT;

CREATE INDEX idx_users_github_app_installation_id ON users(github_app_installation_id);

-- +goose Down
DROP INDEX IF EXISTS idx_users_github_app_installation_id;
ALTER TABLE users DROP COLUMN IF EXISTS github_app_installation_id;
