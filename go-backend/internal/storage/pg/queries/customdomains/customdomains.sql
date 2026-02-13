-- name: CreateCustomDomain :one
INSERT INTO custom_domains (service_id, domain, expected_record_target)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetByServiceID :one
SELECT * FROM custom_domains WHERE service_id = $1;

-- name: GetByDomain :one
SELECT * FROM custom_domains WHERE lower(domain) = lower($1);

-- name: UpdateStatus :one
UPDATE custom_domains SET status = $2, last_checked_at = NOW(), updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: UpdateVerified :one
UPDATE custom_domains SET status = 'active', verified_at = NOW(), updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: UpdateError :one
UPDATE custom_domains SET last_error = $2, last_checked_at = NOW(), updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: Delete :exec
DELETE FROM custom_domains WHERE id = $1;

-- name: DeleteByServiceID :exec
DELETE FROM custom_domains WHERE service_id = $1;

-- name: ExpireStale :exec
UPDATE custom_domains SET status = 'failed', last_error = 'expired', updated_at = NOW()
WHERE status = 'pending_dns' AND expires_at < NOW();
