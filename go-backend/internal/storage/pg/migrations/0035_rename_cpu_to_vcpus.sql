-- +goose Up
ALTER TABLE services RENAME COLUMN cpu TO vcpus;

-- +goose Down
ALTER TABLE services RENAME COLUMN vcpus TO cpu;
