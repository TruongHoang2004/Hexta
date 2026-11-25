-- Create "script_member_permissions" table
CREATE TABLE "script_member_permissions" (
  "tenant_id" character varying(36) NOT NULL,
  "user_id" character varying(255) NOT NULL,
  "can_create" boolean NOT NULL DEFAULT false,
  "can_edit" boolean NOT NULL DEFAULT false,
  "can_activate" boolean NOT NULL DEFAULT false,
  "can_archive" boolean NOT NULL DEFAULT false,
  "version" bigint NOT NULL,
  "updated_by" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("tenant_id", "user_id"),
  CONSTRAINT "script_member_permissions_version_check" CHECK (version > 0)
);
-- Create "script_versions" table
CREATE TABLE "script_versions" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "script_id" character varying(36) NOT NULL,
  "number" bigint NOT NULL,
  "triggers" jsonb NOT NULL,
  "answer" text NOT NULL,
  "created_by" character varying(255) NOT NULL,
  "updated_by" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "published_at" timestamptz NULL,
  "published_by" character varying(255) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "idx_script_version_number" UNIQUE ("tenant_id", "script_id", "number"),
  CONSTRAINT "script_versions_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "script_versions_tenant_id_script_id_id_key" UNIQUE ("tenant_id", "script_id", "id"),
  CONSTRAINT "script_versions_answer_check" CHECK ((char_length(answer) >= 1) AND (char_length(answer) <= 2000)),
  CONSTRAINT "script_versions_number_check" CHECK (number > 0),
  CONSTRAINT "script_versions_published_check" CHECK ((published_at IS NULL) = (published_by IS NULL)),
  CONSTRAINT "script_versions_triggers_check" CHECK (jsonb_typeof(triggers) = 'array'::text)
);
-- Create index "idx_script_versions_one_draft" to table: "script_versions"
CREATE UNIQUE INDEX "idx_script_versions_one_draft" ON "script_versions" ("tenant_id", "script_id") WHERE (published_at IS NULL);
-- Create "scripts" table
CREATE TABLE "scripts" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "name" character varying(120) NOT NULL,
  "status" character varying(16) NOT NULL,
  "active_version_id" character varying(36) NULL,
  "version" bigint NOT NULL,
  "create_hash" character varying(64) NOT NULL,
  "created_by" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "archived_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "scripts_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "scripts_active_pointer_check" CHECK (((status)::text = 'active'::text) = (active_version_id IS NOT NULL)),
  CONSTRAINT "scripts_archived_check" CHECK (((status)::text = 'archived'::text) = (archived_at IS NOT NULL)),
  CONSTRAINT "scripts_status_check" CHECK ((status)::text = ANY (ARRAY[('inactive'::character varying)::text, ('active'::character varying)::text, ('archived'::character varying)::text])),
  CONSTRAINT "scripts_version_check" CHECK (version > 0)
);
-- Create index "idx_scripts_tenant_updated" to table: "scripts"
CREATE INDEX "idx_scripts_tenant_updated" ON "scripts" ("tenant_id", "updated_at" DESC, "id");
-- Modify "script_member_permissions" table
ALTER TABLE "script_member_permissions" ADD CONSTRAINT "script_member_permissions_member_fkey" FOREIGN KEY ("tenant_id", "user_id") REFERENCES "tenant_members" ("tenant_id", "user_id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "script_versions" table
ALTER TABLE "script_versions" ADD CONSTRAINT "script_versions_tenant_id_script_id_fkey" FOREIGN KEY ("tenant_id", "script_id") REFERENCES "scripts" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "scripts" table
ALTER TABLE "scripts" ADD CONSTRAINT "scripts_active_version_fkey" FOREIGN KEY ("tenant_id", "id", "active_version_id") REFERENCES "script_versions" ("tenant_id", "script_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "scripts_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
