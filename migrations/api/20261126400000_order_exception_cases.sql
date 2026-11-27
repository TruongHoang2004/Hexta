-- Create "order_cases" table
CREATE TABLE "order_cases" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "order_id" character varying(36) NOT NULL,
  "case_type" character varying(16) NOT NULL,
  "status" character varying(16) NOT NULL,
  "handler_id" character varying(255) NOT NULL,
  "opened_by" character varying(255) NOT NULL,
  "progress_note" text NOT NULL DEFAULT '',
  "finalized_by" character varying(255) NULL,
  "finalized_at" timestamptz NULL,
  "version" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_cases_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "order_cases_tenant_id_id_order_id_key" UNIQUE ("tenant_id", "id", "order_id"),
  CONSTRAINT "order_cases_tenant_id_order_id_key" UNIQUE ("tenant_id", "order_id"),
  CONSTRAINT "order_cases_tenant_id_order_id_fkey" FOREIGN KEY ("tenant_id", "order_id") REFERENCES "orders" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "order_cases_case_type_check" CHECK ((case_type)::text = ANY (ARRAY[('cancellation'::character varying)::text, ('return'::character varying)::text])),
  CONSTRAINT "order_cases_finalized_check" CHECK (((status)::text = 'finalized'::text) = ((finalized_by IS NOT NULL) AND (finalized_at IS NOT NULL))),
  CONSTRAINT "order_cases_status_check" CHECK ((status)::text = ANY (ARRAY[('open'::character varying)::text, ('sent_back'::character varying)::text, ('finalized'::character varying)::text])),
  CONSTRAINT "order_cases_version_check" CHECK (version > 0)
);
-- Create "order_case_events" table
CREATE TABLE "order_case_events" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "case_id" character varying(36) NOT NULL,
  "version" bigint NOT NULL,
  "actor_id" character varying(255) NOT NULL,
  "action" character varying(16) NOT NULL,
  "status" character varying(16) NOT NULL,
  "note" text NOT NULL DEFAULT '',
  "lines" jsonb NOT NULL DEFAULT '[]',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_case_events_case_id_version_key" UNIQUE ("case_id", "version"),
  CONSTRAINT "order_case_events_case_fkey" FOREIGN KEY ("tenant_id", "case_id") REFERENCES "order_cases" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "order_case_events_action_check" CHECK ((action)::text = ANY (ARRAY[('opened'::character varying)::text, ('updated'::character varying)::text, ('sent_back'::character varying)::text, ('finalized'::character varying)::text])),
  CONSTRAINT "order_case_events_status_check" CHECK ((status)::text = ANY (ARRAY[('open'::character varying)::text, ('sent_back'::character varying)::text, ('finalized'::character varying)::text])),
  CONSTRAINT "order_case_events_version_check" CHECK (version > 0)
);
-- Create index "idx_order_items_case_reference" to table: "order_items"
CREATE UNIQUE INDEX "idx_order_items_case_reference" ON "order_items" ("tenant_id", "order_id", "id", "quantity");
-- Create "order_case_lines" table
CREATE TABLE "order_case_lines" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "case_id" character varying(36) NOT NULL,
  "order_id" character varying(36) NOT NULL,
  "order_item_id" character varying(36) NOT NULL,
  "original_quantity" bigint NOT NULL,
  "quantity" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_case_lines_case_id_order_item_id_key" UNIQUE ("case_id", "order_item_id"),
  CONSTRAINT "order_case_lines_case_fkey" FOREIGN KEY ("tenant_id", "case_id", "order_id") REFERENCES "order_cases" ("tenant_id", "id", "order_id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "order_case_lines_item_fkey" FOREIGN KEY ("tenant_id", "order_id", "order_item_id", "original_quantity") REFERENCES "order_items" ("tenant_id", "order_id", "id", "quantity") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "order_case_lines_quantity_check" CHECK ((quantity > 0) AND (quantity <= original_quantity))
);
