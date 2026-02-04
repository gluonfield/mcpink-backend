-- +goose Up
CREATE TABLE dns_records (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    app_id TEXT REFERENCES apps(id) ON DELETE SET NULL,
    cloudflare_record_id TEXT NOT NULL,
    subdomain TEXT NOT NULL,
    full_domain TEXT NOT NULL,
    target_ip TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_dns_records_cloudflare_id ON dns_records(cloudflare_record_id);
CREATE INDEX idx_dns_records_app_id ON dns_records(app_id);
CREATE UNIQUE INDEX idx_dns_records_subdomain ON dns_records(subdomain);

ALTER TABLE apps ADD COLUMN custom_domain TEXT;

-- +goose Down
ALTER TABLE apps DROP COLUMN IF EXISTS custom_domain;
DROP TABLE IF EXISTS dns_records;
