-- Idempotent local-only role provisioning; works with existing demo volumes.
DO $$ BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname='hexta_ai') THEN
        CREATE ROLE hexta_ai LOGIN PASSWORD 'local-ai-only';
    END IF;
END $$;
CREATE SCHEMA IF NOT EXISTS ai_runtime AUTHORIZATION hexta_ai;
