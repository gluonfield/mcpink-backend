-- +goose Up
ALTER TABLE apps ADD COLUMN build_progress JSONB;

-- +goose Down
ALTER TABLE apps DROP COLUMN build_progress;
