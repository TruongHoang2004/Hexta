-- Create "products" table
CREATE TABLE "public"."products" (
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
CREATE INDEX "idx_products_tenant_id" ON "public"."products" ("tenant_id");
CREATE INDEX "idx_products_name" ON "public"."products" ("name");
CREATE INDEX "idx_products_category" ON "public"."products" ("category");
CREATE INDEX "idx_products_status" ON "public"."products" ("status");
CREATE INDEX "idx_products_deleted_at" ON "public"."products" ("deleted_at");

-- Create "product_variants" table
CREATE TABLE "public"."product_variants" (
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
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_products_variants" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE INDEX "idx_product_variants_tenant_id" ON "public"."product_variants" ("tenant_id");
CREATE INDEX "idx_product_variants_product_id" ON "public"."product_variants" ("product_id");
CREATE UNIQUE INDEX "idx_variants_tenant_sku" ON "public"."product_variants" ("tenant_id", "sku");
CREATE INDEX "idx_product_variants_barcode" ON "public"."product_variants" ("barcode");
CREATE INDEX "idx_product_variants_deleted_at" ON "public"."product_variants" ("deleted_at");

-- Create "inventory_items" table
CREATE TABLE "public"."inventory_items" (
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
CREATE INDEX "idx_inventory_items_tenant_id" ON "public"."inventory_items" ("tenant_id");
CREATE INDEX "idx_inventory_items_variant_id" ON "public"."inventory_items" ("variant_id");
CREATE UNIQUE INDEX "idx_inventory_tenant_variant" ON "public"."inventory_items" ("tenant_id", "variant_id");

-- Create "stock_movements" table
CREATE TABLE "public"."stock_movements" (
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
CREATE INDEX "idx_stock_movements_tenant_id" ON "public"."stock_movements" ("tenant_id");
CREATE INDEX "idx_stock_movements_inventory_id" ON "public"."stock_movements" ("inventory_id");
CREATE INDEX "idx_stock_movements_movement_type" ON "public"."stock_movements" ("movement_type");
CREATE INDEX "idx_stock_movements_reference_id" ON "public"."stock_movements" ("reference_id");
CREATE INDEX "idx_stock_movements_created_at" ON "public"."stock_movements" ("created_at");

-- Create "customers" table
CREATE TABLE "public"."customers" (
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
CREATE INDEX "idx_customers_tenant_id" ON "public"."customers" ("tenant_id");
CREATE INDEX "idx_customers_full_name" ON "public"."customers" ("full_name");
CREATE INDEX "idx_customers_phone" ON "public"."customers" ("phone");
CREATE UNIQUE INDEX "idx_customers_tenant_phone" ON "public"."customers" ("tenant_id", "phone");
CREATE INDEX "idx_customers_email" ON "public"."customers" ("email");
CREATE INDEX "idx_customers_deleted_at" ON "public"."customers" ("deleted_at");

-- Create "orders" table
CREATE TABLE "public"."orders" (
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
  PRIMARY KEY ("id")
);
CREATE INDEX "idx_orders_tenant_id" ON "public"."orders" ("tenant_id");
CREATE UNIQUE INDEX "idx_orders_tenant_number" ON "public"."orders" ("tenant_id", "order_number");
CREATE INDEX "idx_orders_customer_id" ON "public"."orders" ("customer_id");
CREATE INDEX "idx_orders_employee_id" ON "public"."orders" ("employee_id");
CREATE INDEX "idx_orders_status" ON "public"."orders" ("status");
CREATE INDEX "idx_orders_payment_status" ON "public"."orders" ("payment_status");
CREATE INDEX "idx_orders_source" ON "public"."orders" ("source");
CREATE INDEX "idx_orders_draft_id" ON "public"."orders" ("draft_id");
CREATE INDEX "idx_orders_deleted_at" ON "public"."orders" ("deleted_at");

-- Create "order_items" table
CREATE TABLE "public"."order_items" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "order_id" character varying(36) NOT NULL,
  "variant_id" character varying(36) NOT NULL,
  "quantity" bigint NOT NULL,
  "price" numeric(15,2) NOT NULL DEFAULT 0,
  "line_total" numeric(15,2) NOT NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_orders_items" FOREIGN KEY ("order_id") REFERENCES "public"."orders" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE INDEX "idx_order_items_tenant_id" ON "public"."order_items" ("tenant_id");
CREATE INDEX "idx_order_items_order_id" ON "public"."order_items" ("order_id");
CREATE INDEX "idx_order_items_variant_id" ON "public"."order_items" ("variant_id");

-- Create "audit_logs" table
CREATE TABLE "public"."audit_logs" (
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
CREATE INDEX "idx_audit_logs_tenant_id" ON "public"."audit_logs" ("tenant_id");
CREATE INDEX "idx_audit_logs_user_id" ON "public"."audit_logs" ("user_id");
CREATE INDEX "idx_audit_logs_action" ON "public"."audit_logs" ("action");
CREATE INDEX "idx_audit_logs_entity_type" ON "public"."audit_logs" ("entity_type");
CREATE INDEX "idx_audit_logs_entity_id" ON "public"."audit_logs" ("entity_id");
CREATE INDEX "idx_audit_logs_source" ON "public"."audit_logs" ("source");
CREATE INDEX "idx_audit_logs_draft_id" ON "public"."audit_logs" ("draft_id");
CREATE INDEX "idx_audit_logs_created_at" ON "public"."audit_logs" ("created_at");

-- Create "ai_drafts" table
CREATE TABLE "public"."ai_drafts" (
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
CREATE INDEX "idx_ai_drafts_tenant_id" ON "public"."ai_drafts" ("tenant_id");
CREATE INDEX "idx_ai_drafts_user_id" ON "public"."ai_drafts" ("user_id");
CREATE INDEX "idx_ai_drafts_intent" ON "public"."ai_drafts" ("intent");
CREATE INDEX "idx_ai_drafts_status" ON "public"."ai_drafts" ("status");
CREATE INDEX "idx_ai_drafts_expires_at" ON "public"."ai_drafts" ("expires_at");
