-- name: CreateTask :one
INSERT INTO tasks (title, description, status, priority, due_date, user_id, project_id)
VALUES (
  sqlc.arg('title'),
  sqlc.narg('description'),
  COALESCE(sqlc.narg('status'), 'pending'),
  COALESCE(sqlc.narg('priority'), 1),
  sqlc.narg('due_date'),
  sqlc.arg('user_id'),
  sqlc.narg('project_id')
)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id');

-- name: ListTasks :many
SELECT * FROM tasks
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
  AND (sqlc.narg('due_before')::timestamptz IS NULL OR due_date <= sqlc.narg('due_before')::timestamptz)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListTasksByProject :many
SELECT * FROM tasks
WHERE user_id = sqlc.arg('user_id') AND project_id = sqlc.arg('project_id')
ORDER BY created_at DESC;

-- name: UpdateTask :one
UPDATE tasks
SET
  title       = COALESCE(sqlc.narg('title'), title),
  description = COALESCE(sqlc.narg('description'), description),
  status      = COALESCE(sqlc.narg('status'), status),
  priority    = COALESCE(sqlc.narg('priority'), priority),
  due_date    = COALESCE(sqlc.narg('due_date'), due_date),
  project_id  = COALESCE(sqlc.narg('project_id'), project_id)
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id');
