-- Create index "idx_orders_tenant_identity" to table: "orders"
CREATE UNIQUE INDEX "idx_orders_tenant_identity" ON "orders" ("tenant_id", "id");
-- Create "workflow_definitions" table
CREATE TABLE "workflow_definitions" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "name" character varying(120) NOT NULL,
  "version" bigint NOT NULL,
  "create_hash" character varying(64) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "workflow_definitions_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "workflow_definitions_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "workflow_definitions_version_check" CHECK (version >= 0)
);
-- Create index "idx_workflow_definitions_tenant_id" to table: "workflow_definitions"
CREATE INDEX "idx_workflow_definitions_tenant_id" ON "workflow_definitions" ("tenant_id");
-- Create "workflow_versions" table
CREATE TABLE "workflow_versions" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "definition_id" character varying(36) NOT NULL,
  "number" bigint NOT NULL,
  "name" character varying(120) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "idx_workflow_version_number" UNIQUE ("tenant_id", "definition_id", "number"),
  CONSTRAINT "workflow_versions_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "workflow_versions_tenant_id_definition_id_fkey" FOREIGN KEY ("tenant_id", "definition_id") REFERENCES "workflow_definitions" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "workflow_versions_number_check" CHECK (number > 0)
);
-- Create "workflow_steps" table
CREATE TABLE "workflow_steps" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "workflow_version_id" character varying(36) NOT NULL,
  "ordinal" integer NOT NULL,
  "name" character varying(120) NOT NULL,
  "handoff_mode" character varying(16) NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "idx_workflow_step_ordinal" UNIQUE ("workflow_version_id", "ordinal"),
  CONSTRAINT "workflow_steps_tenant_id_workflow_version_id_ordinal_key" UNIQUE ("tenant_id", "workflow_version_id", "ordinal"),
  CONSTRAINT "workflow_steps_tenant_id_workflow_version_id_fkey" FOREIGN KEY ("tenant_id", "workflow_version_id") REFERENCES "workflow_versions" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "workflow_steps_handoff_mode_check" CHECK ((handoff_mode)::text = ANY ((ARRAY['direct'::character varying, 'acceptance'::character varying])::text[])),
  CONSTRAINT "workflow_steps_ordinal_check" CHECK ((ordinal >= 0) AND (ordinal < 30))
);
-- Create "work_items" table
CREATE TABLE "work_items" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "order_id" character varying(36) NOT NULL,
  "workflow_version_id" character varying(36) NOT NULL,
  "current_step" integer NOT NULL,
  "responsible_id" character varying(255) NULL,
  "recipient_id" character varying(255) NULL,
  "state" character varying(16) NOT NULL,
  "version" bigint NOT NULL,
  "created_by" character varying(255) NOT NULL,
  "request_key" character varying(36) NOT NULL,
  "create_hash" character varying(64) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "idx_work_create_key" UNIQUE ("tenant_id", "created_by", "request_key"),
  CONSTRAINT "work_items_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "work_items_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "work_items_tenant_id_order_id_fkey" FOREIGN KEY ("tenant_id", "order_id") REFERENCES "orders" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "work_items_tenant_id_workflow_version_id_current_step_fkey" FOREIGN KEY ("tenant_id", "workflow_version_id", "current_step") REFERENCES "workflow_steps" ("tenant_id", "workflow_version_id", "ordinal") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "work_items_tenant_id_workflow_version_id_fkey" FOREIGN KEY ("tenant_id", "workflow_version_id") REFERENCES "workflow_versions" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "work_items_check" CHECK ((((state)::text = 'waiting'::text) AND (responsible_id IS NOT NULL) AND (recipient_id IS NOT NULL) AND ((recipient_id)::text <> (responsible_id)::text)) OR (((state)::text <> 'waiting'::text) AND (recipient_id IS NULL))),
  CONSTRAINT "work_items_check1" CHECK (((state)::text <> 'completed'::text) OR (responsible_id IS NOT NULL)),
  CONSTRAINT "work_items_state_check" CHECK ((state)::text = ANY ((ARRAY['active'::character varying, 'waiting'::character varying, 'completed'::character varying])::text[])),
  CONSTRAINT "work_items_version_check" CHECK (version > 0)
);
-- Create index "idx_work_items_tenant_created" to table: "work_items"
CREATE INDEX "idx_work_items_tenant_created" ON "work_items" ("tenant_id", "created_at" DESC, "id");
-- Create "work_events" table
CREATE TABLE "work_events" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "work_item_id" character varying(36) NOT NULL,
  "version" bigint NOT NULL,
  "actor_id" character varying(255) NOT NULL,
  "action" character varying(32) NOT NULL,
  "from_responsible_id" character varying(255) NULL,
  "responsible_id" character varying(255) NULL,
  "recipient_id" character varying(255) NULL,
  "from_step" integer NOT NULL,
  "current_step" integer NOT NULL,
  "state" character varying(16) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "idx_work_event_version" UNIQUE ("work_item_id", "version"),
  CONSTRAINT "work_events_tenant_id_work_item_id_fkey" FOREIGN KEY ("tenant_id", "work_item_id") REFERENCES "work_items" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "work_events_action_check" CHECK ((action)::text = ANY ((ARRAY['created'::character varying, 'handoff'::character varying, 'accept'::character varying, 'assign'::character varying, 'complete'::character varying])::text[])),
  CONSTRAINT "work_events_state_check" CHECK ((state)::text = ANY ((ARRAY['active'::character varying, 'waiting'::character varying, 'completed'::character varying])::text[])),
  CONSTRAINT "work_events_version_check" CHECK (version > 0)
);
