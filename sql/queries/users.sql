-- name: CreateUser :one
INSERT INTO users (email, full_name, provider, provider_user_id, avatar_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByProvider :one
SELECT * FROM users WHERE provider = $1 AND provider_user_id = $2;

-- name: UpdateLastLogin :one
UPDATE users 
SET last_login = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;
