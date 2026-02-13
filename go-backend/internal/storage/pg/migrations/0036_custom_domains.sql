-- +goose Up
CREATE TABLE custom_domains (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    domain TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending_dns',
    expected_record_target TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
    verified_at TIMESTAMPTZ,
    last_checked_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('pending_dns', 'provisioning', 'active', 'failed')),
    CHECK (domain = lower(domain))
);

CREATE UNIQUE INDEX idx_custom_domains_service ON custom_domains(service_id);
CREATE UNIQUE INDEX idx_custom_domains_domain ON custom_domains(lower(domain));

-- +goose Down
DROP TABLE IF EXISTS custom_domains;
