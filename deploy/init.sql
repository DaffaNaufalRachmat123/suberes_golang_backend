-- deploy/init.sql
-- ─────────────────────────────────────────────────────────────────────────────
-- Executed once by PostgreSQL on first container start (via entrypoint-initdb.d).
-- The database and superuser are already created by Docker via POSTGRES_DB /
-- POSTGRES_USER / POSTGRES_PASSWORD env vars.
-- This file only enables required PostgreSQL extensions.
-- ─────────────────────────────────────────────────────────────────────────────

-- PostGIS: spatial queries for mitra search radius (requires postgis/postgis image)
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- UUID generation helper
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Full-text search helpers (optional but useful)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;
