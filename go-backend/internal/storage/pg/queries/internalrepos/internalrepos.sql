-- name: CreateInternalRepo :one
INSERT INTO internal_repos (user_id, provider, repo_id, full_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetInternalRepoByID :one
SELECT * FROM internal_repos WHERE id = $1;

-- name: GetInternalRepoByFullName :one
SELECT * FROM internal_repos WHERE full_name = $1;

-- name: ListInternalReposByUserID :many
SELECT * FROM internal_repos
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteInternalRepo :exec
DELETE FROM internal_repos WHERE id = $1;

-- name: DeleteInternalRepoByFullName :exec
DELETE FROM internal_repos WHERE full_name = $1;
