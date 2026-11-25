-- Create index "idx_orders_tenant_created" to table: "orders"
CREATE INDEX "idx_orders_tenant_created" ON "orders" ("tenant_id", "created_at");
