-- +goose Up
ALTER TABLE apps ADD COLUMN build_config JSONB NOT NULL DEFAULT '{}'::JSONB;
ALTER TABLE apps DROP COLUMN IF EXISTS publish_directory;

-- +goose Down
ALTER TABLE apps ADD COLUMN publish_directory TEXT;
ALTER TABLE apps DROP COLUMN IF EXISTS build_config;
