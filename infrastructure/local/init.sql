-- Idempotent local-only role provisioning; works with existing demo volumes.
DO $$ BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname='hexta_ai') THEN
        CREATE ROLE hexta_ai LOGIN PASSWORD 'local-ai-only';
    END IF;
END $$;
CREATE SCHEMA IF NOT EXISTS ai_runtime AUTHORIZATION hexta_ai;
-- pgvector for the AI knowledge store. It is not a trusted extension, so the superuser creates it here; the
-- non-superuser AI migration then finds it (CREATE EXTENSION IF NOT EXISTS is a no-op).
CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;

-- Independent Atlas metadata for the AI owner, separate from API revisions.
CREATE SCHEMA IF NOT EXISTS atlas_ai_revisions AUTHORIZATION hexta_ai;
-- Disposable developer databases are never application databases.
SELECT 'CREATE DATABASE hexta_atlas_dev_api' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname='hexta_atlas_dev_api')
\gexec
SELECT 'CREATE DATABASE hexta_atlas_dev_ai' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname='hexta_atlas_dev_ai')
\gexec

-- Connector owns a separate database that the API login cannot reach (mirrors production isolation).
DO $$ BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname='hexta_connector') THEN
        CREATE ROLE hexta_connector LOGIN PASSWORD 'local-connector-only';
    END IF;
END $$;
SELECT 'CREATE DATABASE hexta_connector OWNER hexta_connector' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname='hexta_connector')
\gexec
REVOKE CONNECT ON DATABASE hexta_connector FROM PUBLIC;
GRANT CONNECT ON DATABASE hexta_connector TO hexta_connector;
SELECT 'CREATE DATABASE hexta_atlas_dev_connector' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname='hexta_atlas_dev_connector')
\gexec
