-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByGitHubID :one
SELECT * FROM users WHERE github_id = $1;

-- name: CreateUser :one
INSERT INTO users (id, github_id, github_username, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users
SET github_username = $2, avatar_url = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: SetCoolifyGitHubAppUUID :one
UPDATE users
SET coolify_github_app_uuid = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ClearCoolifyGitHubAppUUID :one
UPDATE users
SET coolify_github_app_uuid = NULL, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetGiteaUsername :one
UPDATE users
SET gitea_username = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUserByGiteaUsername :one
SELECT * FROM users WHERE gitea_username = $1;

-- name: CreateFirebaseUser :one
INSERT INTO users (id, email, display_name, avatar_url) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: LinkGitHub :one
UPDATE users
SET github_id = $2, github_username = $3, avatar_url = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;
