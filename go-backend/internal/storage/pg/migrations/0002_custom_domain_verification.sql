-- +goose Up
ALTER TABLE custom_domains ADD COLUMN verification_token TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE custom_domains DROP COLUMN IF EXISTS verification_token;
