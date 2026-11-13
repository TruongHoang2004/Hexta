-- Modify "messaging_ingress" table
ALTER TABLE "messaging_ingress" ADD COLUMN "external_conversation_key" text NOT NULL, ADD COLUMN "source_identity" text NOT NULL, ADD COLUMN "source_revision" text NOT NULL, ADD COLUMN "reply_to" text NOT NULL, ADD COLUMN "source_timestamp" timestamptz NOT NULL, ADD COLUMN "acquisition" text NOT NULL, ADD COLUMN "purpose" text NOT NULL;
