-- Create "tenants" table
CREATE TABLE "public"."tenants" (
  "id" character varying(36) NOT NULL,
  "name" character varying(255) NOT NULL,
  "slug" character varying(255) NOT NULL,
  "plan" character varying(50) NOT NULL DEFAULT 'free',
  "status" character varying(50) NOT NULL DEFAULT 'active',
  "owner_id" character varying(255) NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_tenants_slug" to table: "tenants"
CREATE UNIQUE INDEX "idx_tenants_slug" ON "public"."tenants" ("slug");
-- Create index "idx_tenants_owner_id" to table: "tenants"
CREATE INDEX "idx_tenants_owner_id" ON "public"."tenants" ("owner_id");

-- Create "tenant_members" table
CREATE TABLE "public"."tenant_members" (
  "id" bigserial NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "user_id" character varying(255) NOT NULL,
  "role" character varying(50) NOT NULL DEFAULT 'member',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tenants_members" FOREIGN KEY ("tenant_id") REFERENCES "public"."tenants" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_tenant_members_tenant_id" to table: "tenant_members"
CREATE INDEX "idx_tenant_members_tenant_id" ON "public"."tenant_members" ("tenant_id");
-- Create index "idx_tenant_members_user_id" to table: "tenant_members"
CREATE INDEX "idx_tenant_members_user_id" ON "public"."tenant_members" ("user_id");
-- Create index "idx_tenant_user" to table: "tenant_members"
CREATE UNIQUE INDEX "idx_tenant_user" ON "public"."tenant_members" ("tenant_id", "user_id");
