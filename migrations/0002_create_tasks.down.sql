DROP TRIGGER IF EXISTS trg_set_updated_at ON tasks;
DROP FUNCTION IF EXISTS set_updated_at();
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_due_date;
DROP TABLE IF EXISTS tasks;