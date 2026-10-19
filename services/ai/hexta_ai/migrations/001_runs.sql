CREATE TABLE IF NOT EXISTS ai_runtime.runs (
    tenant_id text NOT NULL,
    run_id text NOT NULL,
    request bytea NOT NULL,
    digest bytea NOT NULL,
    result bytea NOT NULL,
    status smallint NOT NULL CHECK (status BETWEEN 1 AND 5),
    attempts integer NOT NULL DEFAULT 0,
    lease_token uuid,
    lease_until timestamptz,
    error_code text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (tenant_id, run_id),
    CHECK ((status = 2) = (lease_token IS NOT NULL AND lease_until IS NOT NULL))
);
CREATE INDEX IF NOT EXISTS runs_queue ON ai_runtime.runs(status, lease_until, created_at);
