
-- +migrate Up
CREATE TABLE IF NOT EXISTS "todos" (
  "id" serial PRIMARY KEY,
  "name" varchar(50) NOT NULL UNIQUE,
  "category" varchar(50) NOT NULL,
  "description" text NULL,
  "created_at" timestamp NULL DEFAULT (now()),
  "updated_at" timestamp NULL DEFAULT (now())
);

-- +migrate Down
DROP TABLE IF EXISTS "todos";