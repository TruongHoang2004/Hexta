-- Create "categories" table
CREATE TABLE "public"."categories" (
  "id" text NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "products" table
CREATE TABLE "public"."products" (
  "id" text NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "price" numeric(10,2) NOT NULL,
  "stock" integer NOT NULL DEFAULT 0,
  "category_id" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_products_category_id" to table: "products"
CREATE INDEX "idx_products_category_id" ON "public"."products" ("category_id");
-- Create index "idx_products_name" to table: "products"
CREATE INDEX "idx_products_name" ON "public"."products" ("name");
