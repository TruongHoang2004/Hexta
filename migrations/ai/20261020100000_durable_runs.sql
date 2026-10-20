-- Existing schema owners need not receive database-wide CREATE privileges.
DO $$ BEGIN
 IF NOT EXISTS (SELECT FROM pg_namespace WHERE nspname='ai_runtime') THEN
   EXECUTE 'CREATE SCHEMA ai_runtime';
 END IF;
END $$;
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

-- Adopt only the exact legacy table shape; abort on incompatible existing state.
CREATE TEMP TABLE hexta_expected_runs (
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
DO $$
DECLARE actual oid := 'ai_runtime.runs'::regclass;
        expected oid := 'pg_temp.hexta_expected_runs'::regclass;
BEGIN
 IF EXISTS (
   WITH actual_columns AS (
     SELECT a.attname, a.atttypid, a.atttypmod, a.attnotnull,
            pg_get_expr(d.adbin,d.adrelid) AS default_expression
     FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum
     WHERE a.attrelid=actual AND a.attnum>0 AND NOT a.attisdropped
   ), expected_columns AS (
     SELECT a.attname, a.atttypid, a.atttypmod, a.attnotnull,
            pg_get_expr(d.adbin,d.adrelid) AS default_expression
     FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum
     WHERE a.attrelid=expected AND a.attnum>0 AND NOT a.attisdropped
   )
   SELECT * FROM ((TABLE actual_columns EXCEPT TABLE expected_columns)
                  UNION ALL (TABLE expected_columns EXCEPT TABLE actual_columns)) differences
 ) OR EXISTS (
   WITH actual_constraints AS (
     SELECT contype, pg_get_constraintdef(oid) definition FROM pg_constraint WHERE conrelid=actual
   ), expected_constraints AS (
     SELECT contype, pg_get_constraintdef(oid) definition FROM pg_constraint WHERE conrelid=expected
   )
   SELECT * FROM ((TABLE actual_constraints EXCEPT TABLE expected_constraints)
                  UNION ALL (TABLE expected_constraints EXCEPT TABLE actual_constraints)) differences
 ) THEN
   RAISE EXCEPTION 'Existing ai_runtime.runs is incompatible; review its schema before Atlas adoption';
 END IF;
 IF NOT EXISTS (
   SELECT 1 FROM pg_index WHERE indexrelid='ai_runtime.runs_queue'::regclass
   AND indisvalid AND NOT indisunique AND indpred IS NULL AND indexprs IS NULL
   AND pg_get_indexdef(indexrelid) LIKE '% USING btree (status, lease_until, created_at)'
 ) THEN
   RAISE EXCEPTION 'Existing AI queue index is incompatible; review before Atlas adoption';
 END IF;
END $$;
DROP TABLE hexta_expected_runs;
