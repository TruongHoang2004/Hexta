// Explicit versioned migrations. Runtime applications never apply desired schemas.
env "api" {
  url = getenv("API_DATABASE_URL")
  dev = getenv("API_DEV_URL") != "" ? getenv("API_DEV_URL") : "docker://postgres/17.7-alpine3.23/dev?search_path=public"
  src = "file://migrations/schema/api.sql"
  migration {
    dir = "file://migrations/api"
    revisions_schema = "atlas_schema_revisions"
  }
}

env "ai" {
  url = getenv("AI_DATABASE_URL")
  // The AI schema needs pgvector: a disposable pgvector container per command. migrations/schema/ai.sql creates the
  // extension in a DO block because Atlas drops it from the dev database after each diff. A custom AI_DEV_URL must
  // point at a disposable pgvector database whose login may run CREATE EXTENSION.
  dev = getenv("AI_DEV_URL") != "" ? getenv("AI_DEV_URL") : "docker+postgres://_/pgvector/pgvector:0.8.1-pg17-bookworm/dev"
  src = "file://migrations/schema/ai.sql"
  migration {
    dir = "file://migrations/ai"
    revisions_schema = "atlas_ai_revisions"
  }
}

env "connector" {
 url = getenv("CONNECTOR_DATABASE_URL")
 dev = getenv("CONNECTOR_DEV_URL") != "" ? getenv("CONNECTOR_DEV_URL") : "docker://postgres/17.7-alpine3.23/dev"
 src = "file://migrations/schema/connector.sql"
 migration {
  dir = "file://migrations/connector"
  revisions_schema = "atlas_connector_revisions"
 }
}
