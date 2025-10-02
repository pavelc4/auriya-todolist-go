-- name: CreateProject :one
INSERT INTO projects (
  user_id,
  name
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetProject :one
SELECT * FROM projects
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET name = $3, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1 AND user_id = $2;
