-- +goose Up
UPDATE apps SET build_pack = 'railpack' WHERE build_pack = 'nixpacks';
ALTER TABLE apps ALTER COLUMN build_pack SET DEFAULT 'railpack';

-- +goose Down
UPDATE apps SET build_pack = 'nixpacks' WHERE build_pack = 'railpack';
ALTER TABLE apps ALTER COLUMN build_pack SET DEFAULT 'nixpacks';
