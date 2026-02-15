-- +goose Up
ALTER TABLE clusters ADD COLUMN has_dns BOOLEAN NOT NULL DEFAULT false;
UPDATE clusters SET has_dns = true WHERE region = 'eu-central-1';

-- +goose Down
ALTER TABLE clusters DROP COLUMN IF EXISTS has_dns;
