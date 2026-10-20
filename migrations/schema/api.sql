-- Create "sessions" table
CREATE TABLE "sessions" (
  "id" bigserial NOT NULL,
  "user_id" text NOT NULL,
  "token" text NOT NULL,
  "provider" text NOT NULL,
  "device_info" character varying(500) NULL,
  "ip_address" character varying(45) NULL,
  "user_agent" character varying(1000) NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "audit_logs" table
CREATE TABLE "audit_logs" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "user_id" character varying(255) NOT NULL,
  "action" character varying(64) NOT NULL,
  "entity_type" character varying(64) NOT NULL,
  "entity_id" character varying(36) NOT NULL,
  "source" character varying(32) NOT NULL DEFAULT 'web_form',
  "draft_id" character varying(36) NULL,
  "before_state" jsonb NULL,
  "after_state" jsonb NOT NULL,
  "ip_address" character varying(45) NULL,
  "user_agent" text NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_audit_logs_action" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_action" ON "audit_logs" ("action");
-- Create index "idx_audit_logs_created_at" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_created_at" ON "audit_logs" ("created_at");
-- Create index "idx_audit_logs_draft_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_draft_id" ON "audit_logs" ("draft_id");
-- Create index "idx_audit_logs_entity_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_entity_id" ON "audit_logs" ("entity_id");
-- Create index "idx_audit_logs_entity_type" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_entity_type" ON "audit_logs" ("entity_type");
-- Create index "idx_audit_logs_source" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_source" ON "audit_logs" ("source");
-- Create index "idx_audit_logs_tenant_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_tenant_id" ON "audit_logs" ("tenant_id");
-- Create index "idx_audit_logs_user_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_user_id" ON "audit_logs" ("user_id");
-- Create "customers" table
CREATE TABLE "customers" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "full_name" character varying(255) NOT NULL,
  "phone" character varying(32) NULL,
  "email" character varying(255) NULL,
  "address" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_customers_deleted_at" to table: "customers"
CREATE INDEX "idx_customers_deleted_at" ON "customers" ("deleted_at");
-- Create index "idx_customers_email" to table: "customers"
CREATE INDEX "idx_customers_email" ON "customers" ("email");
-- Create index "idx_customers_full_name" to table: "customers"
CREATE INDEX "idx_customers_full_name" ON "customers" ("full_name");
-- Create index "idx_customers_phone" to table: "customers"
CREATE INDEX "idx_customers_phone" ON "customers" ("phone");
-- Create index "idx_customers_tenant_id" to table: "customers"
CREATE INDEX "idx_customers_tenant_id" ON "customers" ("tenant_id");
-- Create index "idx_customers_tenant_phone" to table: "customers"
CREATE UNIQUE INDEX "idx_customers_tenant_phone" ON "customers" ("tenant_id", "phone") WHERE (((phone)::text <> ''::text) AND (deleted_at IS NULL));
-- Create "identities" table
CREATE TABLE "identities" (
  "id" bigserial NOT NULL,
  "user_id" text NOT NULL,
  "provider" character varying(50) NOT NULL,
  "identifier" character varying(255) NOT NULL,
  "password" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_identities_provider_identifier" to table: "identities"
CREATE UNIQUE INDEX "idx_identities_provider_identifier" ON "identities" ("provider", "identifier");
-- Create index "idx_identities_user_id" to table: "identities"
CREATE INDEX "idx_identities_user_id" ON "identities" ("user_id");
-- Create "inventory_items" table
CREATE TABLE "inventory_items" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "variant_id" character varying(36) NOT NULL,
  "on_hand_qty" bigint NOT NULL DEFAULT 0,
  "reserved_qty" bigint NOT NULL DEFAULT 0,
  "available_qty" bigint NOT NULL DEFAULT 0,
  "safety_threshold" bigint NOT NULL DEFAULT 5,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_inventory_items_tenant_id" to table: "inventory_items"
CREATE INDEX "idx_inventory_items_tenant_id" ON "inventory_items" ("tenant_id");
-- Create index "idx_inventory_items_variant_id" to table: "inventory_items"
CREATE INDEX "idx_inventory_items_variant_id" ON "inventory_items" ("variant_id");
-- Create index "idx_inventory_tenant_variant" to table: "inventory_items"
CREATE UNIQUE INDEX "idx_inventory_tenant_variant" ON "inventory_items" ("tenant_id", "variant_id");
-- Create "ai_drafts" table
CREATE TABLE "ai_drafts" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "user_id" character varying(255) NOT NULL,
  "intent" character varying(64) NOT NULL,
  "prompt" text NOT NULL,
  "draft_payload" jsonb NOT NULL,
  "status" character varying(32) NOT NULL DEFAULT 'proposed',
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_ai_drafts_expires_at" to table: "ai_drafts"
CREATE INDEX "idx_ai_drafts_expires_at" ON "ai_drafts" ("expires_at");
-- Create index "idx_ai_drafts_intent" to table: "ai_drafts"
CREATE INDEX "idx_ai_drafts_intent" ON "ai_drafts" ("intent");
-- Create index "idx_ai_drafts_status" to table: "ai_drafts"
CREATE INDEX "idx_ai_drafts_status" ON "ai_drafts" ("status");
-- Create index "idx_ai_drafts_tenant_id" to table: "ai_drafts"
CREATE INDEX "idx_ai_drafts_tenant_id" ON "ai_drafts" ("tenant_id");
-- Create index "idx_ai_drafts_user_id" to table: "ai_drafts"
CREATE INDEX "idx_ai_drafts_user_id" ON "ai_drafts" ("user_id");
-- Create "orders" table
CREATE TABLE "orders" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "order_number" character varying(64) NOT NULL,
  "customer_id" character varying(36) NULL,
  "employee_id" character varying(36) NULL,
  "status" character varying(32) NOT NULL DEFAULT 'draft',
  "payment_status" character varying(32) NOT NULL DEFAULT 'unpaid',
  "subtotal" numeric(15,2) NOT NULL DEFAULT 0,
  "discount" numeric(15,2) NOT NULL DEFAULT 0,
  "total_amount" numeric(15,2) NOT NULL DEFAULT 0,
  "source" character varying(32) NOT NULL DEFAULT 'web_form',
  "draft_id" character varying(36) NULL,
  "notes" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "version" bigint NOT NULL DEFAULT 1,
  "requires_delivery" boolean NOT NULL DEFAULT false,
  "delivery_address" text NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "idx_orders_customer_id" to table: "orders"
CREATE INDEX "idx_orders_customer_id" ON "orders" ("customer_id");
-- Create index "idx_orders_deleted_at" to table: "orders"
CREATE INDEX "idx_orders_deleted_at" ON "orders" ("deleted_at");
-- Create index "idx_orders_draft_id" to table: "orders"
CREATE INDEX "idx_orders_draft_id" ON "orders" ("draft_id");
-- Create index "idx_orders_employee_id" to table: "orders"
CREATE INDEX "idx_orders_employee_id" ON "orders" ("employee_id");
-- Create index "idx_orders_payment_status" to table: "orders"
CREATE INDEX "idx_orders_payment_status" ON "orders" ("payment_status");
-- Create index "idx_orders_source" to table: "orders"
CREATE INDEX "idx_orders_source" ON "orders" ("source");
-- Create index "idx_orders_status" to table: "orders"
CREATE INDEX "idx_orders_status" ON "orders" ("status");
-- Create index "idx_orders_tenant_id" to table: "orders"
CREATE INDEX "idx_orders_tenant_id" ON "orders" ("tenant_id");
-- Create index "idx_orders_tenant_identity" to table: "orders"
CREATE UNIQUE INDEX "idx_orders_tenant_identity" ON "orders" ("tenant_id", "id");
-- Create index "idx_orders_tenant_number" to table: "orders"
CREATE UNIQUE INDEX "idx_orders_tenant_number" ON "orders" ("tenant_id", "order_number");
-- Create "platform_operators" table
CREATE TABLE "platform_operators" (
  "user_id" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("user_id")
);
-- Create "stock_movements" table
CREATE TABLE "stock_movements" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "inventory_id" character varying(36) NOT NULL,
  "movement_type" character varying(32) NOT NULL,
  "quantity" bigint NOT NULL,
  "balance_after" bigint NOT NULL,
  "reference_type" character varying(64) NULL,
  "reference_id" character varying(36) NULL,
  "notes" text NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_stock_movements_created_at" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_created_at" ON "stock_movements" ("created_at");
-- Create index "idx_stock_movements_inventory_id" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_inventory_id" ON "stock_movements" ("inventory_id");
-- Create index "idx_stock_movements_movement_type" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_movement_type" ON "stock_movements" ("movement_type");
-- Create index "idx_stock_movements_reference_id" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_reference_id" ON "stock_movements" ("reference_id");
-- Create index "idx_stock_movements_tenant_id" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_tenant_id" ON "stock_movements" ("tenant_id");
-- Create "order_items" table
CREATE TABLE "order_items" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "order_id" character varying(36) NOT NULL,
  "variant_id" character varying(36) NOT NULL,
  "quantity" bigint NOT NULL,
  "price" numeric(15,2) NOT NULL DEFAULT 0,
  "line_total" numeric(15,2) NOT NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "price_provided" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_orders_items" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_order_items_order_id" to table: "order_items"
CREATE INDEX "idx_order_items_order_id" ON "order_items" ("order_id");
-- Create index "idx_order_items_tenant_id" to table: "order_items"
CREATE INDEX "idx_order_items_tenant_id" ON "order_items" ("tenant_id");
-- Create index "idx_order_items_variant_id" to table: "order_items"
CREATE INDEX "idx_order_items_variant_id" ON "order_items" ("variant_id");
-- Create "products" table
CREATE TABLE "products" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "category" character varying(100) NULL,
  "status" character varying(32) NOT NULL DEFAULT 'active',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_products_category" to table: "products"
CREATE INDEX "idx_products_category" ON "products" ("category");
-- Create index "idx_products_deleted_at" to table: "products"
CREATE INDEX "idx_products_deleted_at" ON "products" ("deleted_at");
-- Create index "idx_products_name" to table: "products"
CREATE INDEX "idx_products_name" ON "products" ("name");
-- Create index "idx_products_status" to table: "products"
CREATE INDEX "idx_products_status" ON "products" ("status");
-- Create index "idx_products_tenant_id" to table: "products"
CREATE INDEX "idx_products_tenant_id" ON "products" ("tenant_id");
-- Create "product_variants" table
CREATE TABLE "product_variants" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "product_id" character varying(36) NOT NULL,
  "sku" character varying(64) NOT NULL,
  "variant_name" character varying(128) NOT NULL,
  "price" numeric(15,2) NOT NULL DEFAULT 0,
  "cost_price" numeric(15,2) NOT NULL DEFAULT 0,
  "barcode" character varying(64) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "price_provided" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_products_variants" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_product_variants_barcode" to table: "product_variants"
CREATE INDEX "idx_product_variants_barcode" ON "product_variants" ("barcode");
-- Create index "idx_product_variants_deleted_at" to table: "product_variants"
CREATE INDEX "idx_product_variants_deleted_at" ON "product_variants" ("deleted_at");
-- Create index "idx_product_variants_product_id" to table: "product_variants"
CREATE INDEX "idx_product_variants_product_id" ON "product_variants" ("product_id");
-- Create index "idx_product_variants_tenant_id" to table: "product_variants"
CREATE INDEX "idx_product_variants_tenant_id" ON "product_variants" ("tenant_id");
-- Create index "idx_variants_tenant_sku" to table: "product_variants"
CREATE UNIQUE INDEX "idx_variants_tenant_sku" ON "product_variants" ("tenant_id", "sku");
-- Create "tenants" table
CREATE TABLE "tenants" (
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
-- Create index "idx_tenants_owner_id" to table: "tenants"
CREATE INDEX "idx_tenants_owner_id" ON "tenants" ("owner_id");
-- Create index "idx_tenants_slug" to table: "tenants"
CREATE UNIQUE INDEX "idx_tenants_slug" ON "tenants" ("slug");
-- Create "tenant_invitations" table
CREATE TABLE "tenant_invitations" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "email" character varying(255) NOT NULL,
  "role" character varying(50) NOT NULL DEFAULT 'member',
  "invited_by" character varying(255) NOT NULL,
  "token_hash" character varying(64) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "expires_at" timestamptz NOT NULL,
  "accepted_by" character varying(255) NOT NULL DEFAULT '',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "tenant_invitations_token_hash_key" UNIQUE ("token_hash"),
  CONSTRAINT "tenant_invitations_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "tenant_invitations_role_check" CHECK ((role)::text = 'member'::text),
  CONSTRAINT "tenant_invitations_status_check" CHECK ((status)::text = ANY ((ARRAY['pending'::character varying, 'accepted'::character varying, 'declined'::character varying, 'revoked'::character varying])::text[]))
);
-- Create index "idx_tenant_invitations_tenant_id" to table: "tenant_invitations"
CREATE INDEX "idx_tenant_invitations_tenant_id" ON "tenant_invitations" ("tenant_id");
-- Create index "idx_tenant_pending_invitation_email" to table: "tenant_invitations"
CREATE UNIQUE INDEX "idx_tenant_pending_invitation_email" ON "tenant_invitations" ("tenant_id", "email") WHERE ((status)::text = 'pending'::text);
-- Create "tenant_members" table
CREATE TABLE "tenant_members" (
  "id" bigserial NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "user_id" character varying(255) NOT NULL,
  "role" character varying(50) NOT NULL DEFAULT 'member',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tenants_members" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_tenant_members_tenant_id" to table: "tenant_members"
CREATE INDEX "idx_tenant_members_tenant_id" ON "tenant_members" ("tenant_id");
-- Create index "idx_tenant_members_user_id" to table: "tenant_members"
CREATE INDEX "idx_tenant_members_user_id" ON "tenant_members" ("user_id");
-- Create index "idx_tenant_user" to table: "tenant_members"
CREATE UNIQUE INDEX "idx_tenant_user" ON "tenant_members" ("tenant_id", "user_id");
-- Create "tenant_registration_requests" table
CREATE TABLE "tenant_registration_requests" (
  "id" character varying(36) NOT NULL,
  "applicant_id" character varying(255) NOT NULL,
  "name" character varying(100) NOT NULL,
  "slug" character varying(100) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "tenant_id" character varying(36) NULL,
  "reviewed_by" character varying(255) NOT NULL DEFAULT '',
  "review_reason" text NOT NULL DEFAULT '',
  "reviewed_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "tenant_registration_requests_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "tenant_registration_requests_status_check" CHECK ((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[]))
);
-- Create index "idx_tenant_registration_active_slug" to table: "tenant_registration_requests"
CREATE UNIQUE INDEX "idx_tenant_registration_active_slug" ON "tenant_registration_requests" ("slug") WHERE ((status)::text <> 'rejected'::text);
-- Create index "idx_tenant_registration_requests_applicant_id" to table: "tenant_registration_requests"
CREATE INDEX "idx_tenant_registration_requests_applicant_id" ON "tenant_registration_requests" ("applicant_id");
-- Create index "idx_tenant_registration_requests_status" to table: "tenant_registration_requests"
CREATE INDEX "idx_tenant_registration_requests_status" ON "tenant_registration_requests" ("status");
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
  CONSTRAINT "workflow_steps_handoff_mode_check" CHECK ((handoff_mode)::text = ANY (ARRAY[('direct'::character varying)::text, ('acceptance'::character varying)::text])),
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
  CONSTRAINT "work_items_state_check" CHECK ((state)::text = ANY (ARRAY[('active'::character varying)::text, ('waiting'::character varying)::text, ('completed'::character varying)::text])),
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
  CONSTRAINT "work_events_action_check" CHECK ((action)::text = ANY (ARRAY[('created'::character varying)::text, ('handoff'::character varying)::text, ('accept'::character varying)::text, ('assign'::character varying)::text, ('complete'::character varying)::text])),
  CONSTRAINT "work_events_state_check" CHECK ((state)::text = ANY (ARRAY[('active'::character varying)::text, ('waiting'::character varying)::text, ('completed'::character varying)::text])),
  CONSTRAINT "work_events_version_check" CHECK (version > 0)
);
