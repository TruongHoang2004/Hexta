-- Create "identities" table
CREATE TABLE "public"."identities" (
  "id" bigserial NOT NULL,
  "user_id" text NOT NULL,
  "provider" character varying(50) NOT NULL,
  "identifier" character varying(255) NOT NULL,
  "password" character varying(255) NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_identities_identifier" to table: "identities"
CREATE UNIQUE INDEX "idx_identities_identifier" ON "public"."identities" ("identifier");
-- Create index "idx_identities_user_id" to table: "identities"
CREATE INDEX "idx_identities_user_id" ON "public"."identities" ("user_id");
-- Create "sessions" table
CREATE TABLE "public"."sessions" (
  "id" bigserial NOT NULL,
  "user_id" text NOT NULL,
  "token" text NOT NULL,
  "provider" text NOT NULL,
  "device_info" character varying(500) NULL,
  "ip_address" character varying(45) NULL,
  "user_agent" character varying(1000) NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
