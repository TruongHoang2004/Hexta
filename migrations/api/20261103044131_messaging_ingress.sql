-- Modify "messaging_messages" table
ALTER TABLE "messaging_messages" DROP CONSTRAINT "messaging_messages_status_check", ADD CONSTRAINT "messaging_messages_status_check" CHECK ((status)::text = ANY ((ARRAY['simulated'::character varying, 'received'::character varying])::text[]));
-- Create "messaging_connections" table
CREATE TABLE "messaging_connections" (
  "id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "version" bigint NOT NULL,
  "platform" text NOT NULL,
  "account_id" text NOT NULL,
  "enabled" boolean NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "messaging_connections_platform_account_id_key" UNIQUE ("platform", "account_id"),
  CONSTRAINT "messaging_connections_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "messaging_connections_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "messaging_connections_version_check" CHECK (version > 0)
);
-- Create "messaging_ingress" table
CREATE TABLE "messaging_ingress" (
  "identity" text NOT NULL,
  "event_id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "connection_id" uuid NOT NULL,
  "message_id" text NOT NULL,
  "received_at" timestamptz NOT NULL,
  "outcome" text NOT NULL,
  PRIMARY KEY ("identity"),
  CONSTRAINT "messaging_ingress_tenant_id_connection_id_fkey" FOREIGN KEY ("tenant_id", "connection_id") REFERENCES "messaging_connections" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
