-- +goose Up

-- Drop custom domain tables (no users, clean slate)
DROP TABLE IF EXISTS custom_domains;
DROP TABLE IF EXISTS dns_records;

-- Ingress IP per cluster (for A records in delegated zones)
ALTER TABLE clusters ADD COLUMN ingress_ip TEXT;
UPDATE clusters SET ingress_ip = '46.225.35.234' WHERE region = 'eu-central-1';
ALTER TABLE clusters ALTER COLUMN ingress_ip SET NOT NULL;

CREATE TABLE delegated_zones (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    zone TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'pending_verification',
    verification_token TEXT NOT NULL,
    wildcard_cert_secret TEXT,
    cert_issued_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    delegated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '7 days'),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_zone_status CHECK (
        status IN ('pending_verification','pending_delegation','provisioning','active','failed')
    )
);
CREATE INDEX idx_delegated_zones_user_id ON delegated_zones(user_id);

CREATE TABLE zone_records (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    zone_id TEXT NOT NULL REFERENCES delegated_zones(id) ON DELETE CASCADE,
    service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(zone_id, name)
);
CREATE INDEX idx_zone_records_service_id ON zone_records(service_id);

-- +goose Down
DROP TABLE IF EXISTS zone_records;
DROP TABLE IF EXISTS delegated_zones;
ALTER TABLE clusters DROP COLUMN IF EXISTS ingress_ip;
