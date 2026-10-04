-- Drop index "idx_identities_identifier" from table: "identities"
DROP INDEX "public"."idx_identities_identifier";
-- Modify "identities" table
ALTER TABLE "public"."identities" ALTER COLUMN "password" DROP NOT NULL;
-- Create index "idx_identities_provider_identifier" to table: "identities"
CREATE UNIQUE INDEX "idx_identities_provider_identifier" ON "public"."identities" ("provider", "identifier");
