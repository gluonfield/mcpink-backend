-- +goose Up
UPDATE services SET memory = '256Mi' WHERE memory = '128Mi';

-- +goose Down
-- No rollback: 128Mi is no longer a valid tier
