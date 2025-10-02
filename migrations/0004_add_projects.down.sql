-- Revert the changes from 0004_add_projects.up.sql

-- Drop the index on project_id
DROP INDEX IF EXISTS tasks_project_id_idx;

-- Drop the foreign key constraint from tasks
ALTER TABLE "tasks" DROP CONSTRAINT IF EXISTS tasks_project_id_fkey;

-- Drop the project_id column from tasks
ALTER TABLE "tasks" DROP COLUMN IF EXISTS "project_id";

-- Drop the projects table
DROP TABLE IF EXISTS "projects";
