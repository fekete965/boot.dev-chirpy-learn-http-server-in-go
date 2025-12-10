-- name: CreateChirp :one
INSERT INTO chirps (id, user_id, body, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;
