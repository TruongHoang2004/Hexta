-- Modify "messaging_connections" table
ALTER TABLE "messaging_connections" ADD COLUMN "account_username" text NOT NULL DEFAULT '';
