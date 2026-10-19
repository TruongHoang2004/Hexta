-- Preserve missing prices separately from explicitly entered zero prices.
ALTER TABLE "public"."product_variants" ADD COLUMN "price_provided" boolean NOT NULL DEFAULT false;
ALTER TABLE "public"."orders"
  ADD COLUMN "version" bigint NOT NULL DEFAULT 1,
  ADD COLUMN "requires_delivery" boolean NOT NULL DEFAULT false,
  ADD COLUMN "delivery_address" text NOT NULL DEFAULT '';
ALTER TABLE "public"."order_items" ADD COLUMN "price_provided" boolean NOT NULL DEFAULT false;
-- Existing drafts need explicit price review before confirmation. Historical
-- non-draft records retain their legacy price interpretation.
UPDATE "public"."order_items" AS item SET "price_provided" = true
FROM "public"."orders" AS ord WHERE item.order_id = ord.id AND ord.status <> 'draft';

DROP INDEX "public"."idx_customers_tenant_phone";
CREATE UNIQUE INDEX "idx_customers_tenant_phone" ON "public"."customers" ("tenant_id", "phone") WHERE phone <> '' AND deleted_at IS NULL;
