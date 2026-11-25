#!/bin/sh
# Runs once, when the Postgres volume is first created (docker-entrypoint-initdb.d).
# Each service gets its own login: API owns database "hexta" (schema public), AI owns schema ai_runtime in it,
# and the connector owns a separate database "hexta_connector" that the API login cannot connect to.
set -eu
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
	-v api_pw="$API_DB_PASSWORD" -v ai_pw="$AI_DB_PASSWORD" -v connector_pw="$CONNECTOR_DB_PASSWORD" <<'SQL'
CREATE ROLE hexta_api LOGIN PASSWORD :'api_pw';
CREATE ROLE hexta_ai LOGIN PASSWORD :'ai_pw';
CREATE ROLE hexta_connector LOGIN PASSWORD :'connector_pw';

CREATE DATABASE hexta OWNER hexta_api;
\connect hexta
ALTER SCHEMA public OWNER TO hexta_api;
CREATE SCHEMA ai_runtime AUTHORIZATION hexta_ai;
CREATE SCHEMA atlas_ai_revisions AUTHORIZATION hexta_ai;
-- pgvector (AI knowledge store) is not a trusted extension: the superuser creates it so the hexta_ai migrator
-- (non-superuser) succeeds. Existing volumes get it from the postgres-extensions job in compose.prod.yml.
CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;
REVOKE CONNECT ON DATABASE hexta FROM PUBLIC;
GRANT CONNECT ON DATABASE hexta TO hexta_api, hexta_ai;

CREATE DATABASE hexta_connector OWNER hexta_connector;
REVOKE CONNECT ON DATABASE hexta_connector FROM PUBLIC;
GRANT CONNECT ON DATABASE hexta_connector TO hexta_connector;
SQL
