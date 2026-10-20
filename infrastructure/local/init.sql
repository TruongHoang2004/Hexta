-- Idempotent local-only role provisioning; works with existing demo volumes.
DO $$ BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname='hexta_ai') THEN
        CREATE ROLE hexta_ai LOGIN PASSWORD 'local-ai-only';
    END IF;
END $$;
CREATE SCHEMA IF NOT EXISTS ai_runtime AUTHORIZATION hexta_ai;

-- Independent Atlas metadata for the AI owner, separate from API revisions.
CREATE SCHEMA IF NOT EXISTS atlas_ai_revisions AUTHORIZATION hexta_ai;
-- Disposable developer databases are never application databases.
SELECT 'CREATE DATABASE hexta_atlas_dev_api' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname='hexta_atlas_dev_api')
\gexec
SELECT 'CREATE DATABASE hexta_atlas_dev_ai' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname='hexta_atlas_dev_ai')
\gexec
