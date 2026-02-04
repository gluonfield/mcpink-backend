-- name: CreateDNSRecord :one
INSERT INTO dns_records (app_id, cloudflare_record_id, subdomain, full_domain, target_ip)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetDNSRecordByAppID :one
SELECT * FROM dns_records WHERE app_id = $1;

-- name: GetDNSRecordBySubdomain :one
SELECT * FROM dns_records WHERE subdomain = $1;

-- name: GetDNSRecordByCloudflareID :one
SELECT * FROM dns_records WHERE cloudflare_record_id = $1;

-- name: DeleteDNSRecord :exec
DELETE FROM dns_records WHERE id = $1;

-- name: DeleteDNSRecordByAppID :exec
DELETE FROM dns_records WHERE app_id = $1;

-- name: UpdateDNSRecordTarget :one
UPDATE dns_records
SET target_ip = $2
WHERE id = $1
RETURNING *;
