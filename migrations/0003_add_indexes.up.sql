-- Add indexes for performance optimization

-- Index for user_id on tasks table for fast task lookups by user
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON "tasks" ("user_id");

-- Unique index on email for users table for fast login lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON "users" ("email");

-- Composite unique index on provider and provider_user_id for fast OAuth lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider_provider_user_id ON "users" ("provider", "provider_user_id");
