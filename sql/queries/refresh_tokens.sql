-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expires_at, revoked_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
returning *;

-- name: FindRefreshToken :one
SELECT rt.*
FROM refresh_tokens rt
INNER JOIN users u ON u.id = rt.user_id
WHERE token = $1
LIMIT 1;

-- name: RevokeRefreshToken :execrows
UPDATE refresh_tokens SET revoked_at = $1, updated_at = $2 WHERE token = $3;

-- name: ExpireRefreshToken :execrows
UPDATE refresh_tokens SET expires_at = $1, updated_at = $2 WHERE token = $3;
