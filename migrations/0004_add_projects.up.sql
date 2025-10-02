-- Create the projects table
CREATE TABLE "projects" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "name" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

-- Add foreign key constraint to projects table
ALTER TABLE "projects" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

-- Add the project_id column to the tasks table
ALTER TABLE "tasks" ADD COLUMN "project_id" bigint;

-- Add foreign key constraint to tasks table for project_id
ALTER TABLE "tasks" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON DELETE SET NULL;

-- Add an index on project_id for faster lookups
CREATE INDEX ON "tasks" ("project_id");
