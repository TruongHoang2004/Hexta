-- Modify "order_items" table
ALTER TABLE "order_items" DROP CONSTRAINT "fk_orders_items", ADD CONSTRAINT "fk_orders_items_tenant" FOREIGN KEY ("tenant_id", "order_id") REFERENCES "orders" ("tenant_id", "id") ON UPDATE CASCADE ON DELETE CASCADE;
-- Create index "idx_products_tenant_identity" to table: "products"
CREATE UNIQUE INDEX "idx_products_tenant_identity" ON "products" ("tenant_id", "id");
-- Modify "product_variants" table
ALTER TABLE "product_variants" DROP CONSTRAINT "fk_products_variants", ADD CONSTRAINT "fk_products_variants_tenant" FOREIGN KEY ("tenant_id", "product_id") REFERENCES "products" ("tenant_id", "id") ON UPDATE CASCADE ON DELETE CASCADE;
