-- deploy/init.sql
-- ─────────────────────────────────────────────────────────────────────────────
-- Executed once by PostgreSQL on first container start (via entrypoint-initdb.d).
-- Creates a least-privilege application user and configures extensions.
-- ─────────────────────────────────────────────────────────────────────────────

-- Enable PostGIS (spatial queries used by mitra search)
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ── Application database user ─────────────────────────────────────────────────
-- The actual password is set via POSTGRES_PASSWORD env var for the superuser;
-- the app user password must be changed to a strong secret before going live.
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'suberes_app') THEN
        CREATE ROLE suberes_app LOGIN PASSWORD 'CHANGE_ME_STRONG_PASSWORD';
    END IF;
END
$$;

-- Grant the app user exactly what it needs — no superuser, no create-db
GRANT CONNECT ON DATABASE suberes TO suberes_app;
GRANT USAGE   ON SCHEMA public TO suberes_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES    IN SCHEMA public TO suberes_app;
GRANT USAGE, SELECT                  ON ALL SEQUENCES IN SCHEMA public TO suberes_app;
GRANT EXECUTE                        ON ALL FUNCTIONS IN SCHEMA public TO suberes_app;

-- Ensure future tables/sequences created by GORM auto-migrate inherit these grants
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES    TO suberes_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT                  ON SEQUENCES TO suberes_app;

-- ── Connection pool recommendation ────────────────────────────────────────────
-- With pgBouncer (transaction-mode) pool of 20, set in postgresql.conf:
--   max_connections = 100
--   shared_buffers  = 256MB      (25 % of RAM on a 1 GB VPS)
--   work_mem        = 4MB
-- Without pgBouncer, keep GORM pool ≤ 25 connections.
