-- name: CreateTask :one
INSERT INTO tasks (title, description, status, priority, due_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListTasks :many
SELECT * FROM tasks
WHERE ($1::text IS NULL OR status = $1)
  AND ($2::timestamptz IS NULL OR due_date <= $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateTask :one
UPDATE tasks
SET 
  title = COALESCE($2, title),
  description = COALESCE($3, description),
  status = COALESCE($4, status),
  priority = COALESCE($5, priority),
  due_date = COALESCE($6, due_date),
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;
