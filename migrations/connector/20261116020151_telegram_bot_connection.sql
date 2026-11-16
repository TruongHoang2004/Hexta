-- Modify "connections" table
ALTER TABLE "connector"."connections" DROP CONSTRAINT "connections_platform_check", ADD CONSTRAINT "connections_platform_check" CHECK (platform = ANY (ARRAY['facebook'::text, 'telegram'::text])), ADD CONSTRAINT "connections_check1" CHECK ((webhook_secret_ciphertext IS NULL) OR (credential_ciphertext IS NOT NULL)), ADD COLUMN "webhook_secret_ciphertext" bytea NULL;
