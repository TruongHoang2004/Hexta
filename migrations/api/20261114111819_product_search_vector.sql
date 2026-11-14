-- Modify "products" table
ALTER TABLE "products" ADD COLUMN "search_vector" tsvector NULL GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, (((((COALESCE(name, ''::character varying))::text || ' '::text) || (COALESCE(category, ''::character varying))::text) || ' '::text) || COALESCE(description, ''::text)))) STORED;
-- Create index "idx_products_search_vector" to table: "products"
CREATE INDEX "idx_products_search_vector" ON "products" USING GIN ("search_vector");
