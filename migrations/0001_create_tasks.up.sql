CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    priority INT NOT NULL DEFAULT 1,
    due_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE tasks
    ADD CONSTRAINT status_check
    CHECK (status IN ('pending','in_progress','done'));

-- Index untuk status
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- Index untuk due_date
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);

-- Function untuk auto-update updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger pakai function di atas
CREATE TRIGGER trg_set_updated_at
BEFORE UPDATE ON tasks
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
