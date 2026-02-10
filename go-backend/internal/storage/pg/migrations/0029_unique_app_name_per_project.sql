-- +goose Up
CREATE UNIQUE INDEX idx_apps_name_project_active ON apps(name, project_id) WHERE is_deleted = false;

-- +goose Down
DROP INDEX IF EXISTS idx_apps_name_project_active;
