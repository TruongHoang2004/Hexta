-- pgvector for ai_runtime knowledge embeddings. Custom migration (atlas migrate new): Atlas Community cannot
-- diff extensions. The extension lives in schema public because an Atlas dev database needs an empty ai_runtime.
-- pgvector is not a trusted extension: when the AI migrator login is not a superuser, a superuser must run
-- `CREATE EXTENSION vector WITH SCHEMA public` in this database first; IF NOT EXISTS then makes this a no-op.
CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;
DO $$
DECLARE installed record;
BEGIN
 SELECT n.nspname, e.extversion INTO installed
 FROM pg_extension e JOIN pg_namespace n ON n.oid = e.extnamespace WHERE e.extname = 'vector';
 IF installed.nspname IS DISTINCT FROM 'public' THEN
   RAISE EXCEPTION 'pgvector must be installed in schema public (found %)', installed.nspname;
 END IF;
 -- Filtered HNSW iterative scans (hnsw.iterative_scan) require pgvector 0.8.0 or newer.
 IF string_to_array(split_part(installed.extversion, '-', 1), '.')::int[] < ARRAY[0, 8, 0] THEN
   RAISE EXCEPTION 'pgvector 0.8.0 or newer is required (found %)', installed.extversion;
 END IF;
END $$;
