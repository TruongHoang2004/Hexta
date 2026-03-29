-- Rename a column from "username" to "full_name"
ALTER TABLE "public"."users" RENAME COLUMN "username" TO "full_name";
-- Rename a column from "password" to "gender"
ALTER TABLE "public"."users" RENAME COLUMN "password" TO "gender";
-- Modify "users" table
ALTER TABLE "public"."users" DROP CONSTRAINT "uni_users_username", DROP COLUMN "deleted_at", ADD COLUMN "avatar" text NULL, ADD COLUMN "date_of_birth" text NOT NULL;
-- Create "addresses" table
CREATE TABLE "public"."addresses" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "receiver" text NOT NULL,
  "phone" text NOT NULL,
  "city" text NOT NULL,
  "district" text NOT NULL,
  "ward" text NOT NULL,
  "street" text NOT NULL,
  "details" text NOT NULL,
  "is_default" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_addresses" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_addresses_user_id" to table: "addresses"
CREATE INDEX "idx_addresses_user_id" ON "public"."addresses" ("user_id");
