-- +goose Up
ALTER TABLE users DROP COLUMN IF EXISTS coolify_github_app_uuid;
ALTER TABLE apps DROP COLUMN IF EXISTS coolify_app_uuid;

-- +goose Down
ALTER TABLE users ADD COLUMN coolify_github_app_uuid TEXT;
ALTER TABLE apps ADD COLUMN coolify_app_uuid TEXT;
