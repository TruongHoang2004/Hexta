-- Create "merchants" table
CREATE TABLE "public"."merchants" (
  "id" text NOT NULL,
  "owner_id" text NOT NULL,
  "name" text NOT NULL,
  "email" text NOT NULL,
  "phone" text NOT NULL,
  "description" text NULL,
  "logo_url" character varying(500) NULL,
  "status" text NOT NULL DEFAULT 'pending',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_merchants_email" UNIQUE ("email"),
  CONSTRAINT "uni_merchants_name" UNIQUE ("name"),
  CONSTRAINT "uni_merchants_phone" UNIQUE ("phone")
);
-- Create index "idx_merchants_deleted_at" to table: "merchants"
CREATE INDEX "idx_merchants_deleted_at" ON "public"."merchants" ("deleted_at");
-- Create index "idx_merchants_owner_id" to table: "merchants"
CREATE INDEX "idx_merchants_owner_id" ON "public"."merchants" ("owner_id");
