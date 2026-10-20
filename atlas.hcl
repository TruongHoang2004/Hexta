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
  dev = getenv("AI_DEV_URL") != "" ? getenv("AI_DEV_URL") : "docker://postgres/17.7-alpine3.23/dev"
  src = "file://migrations/schema/ai.sql"
  migration {
    dir = "file://migrations/ai"
    revisions_schema = "atlas_ai_revisions"
  }
}
