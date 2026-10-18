-- Create the database first (docker compose does this for you):
--   createdb employee_db
-- Then run this file:
--   psql -d employee_db -f migrations/001_create_employees.sql

CREATE TABLE IF NOT EXISTS employees (
  id           BIGSERIAL PRIMARY KEY,
  name         VARCHAR(100) NOT NULL,
  email        VARCHAR(255) NOT NULL,
  department   VARCHAR(100) NOT NULL,
  designation  VARCHAR(100) NOT NULL,
  joining_date DATE         NOT NULL,
  status       VARCHAR(20)  NOT NULL DEFAULT 'active'
               CHECK (status IN ('active', 'inactive')),
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_employees_email UNIQUE (email)
);

CREATE INDEX IF NOT EXISTS idx_employees_department ON employees (department);
CREATE INDEX IF NOT EXISTS idx_employees_status     ON employees (status);
