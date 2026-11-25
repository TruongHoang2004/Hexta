-- Custom data migration (atlas migrate new): enqueue one AI knowledge sync job for every existing active product
-- and active script, so the catalog that predates the knowledge sync outbox reaches the AI knowledge store once.
-- Later changes enqueue themselves in their own transaction. Idempotent: existing jobs are left untouched. The job
-- reads each source's current state after commit, so no business data is copied here.
INSERT INTO "knowledge_sync_jobs" ("tenant_id", "source_type", "source_id", "state", "attempts", "fence", "available_at", "updated_at")
SELECT p."tenant_id", 'product', p."id", 'pending', 0, 1, now(), now()
FROM "products" p
WHERE p."status" = 'active' AND p."deleted_at" IS NULL
ON CONFLICT ("tenant_id", "source_type", "source_id") DO NOTHING;

INSERT INTO "knowledge_sync_jobs" ("tenant_id", "source_type", "source_id", "state", "attempts", "fence", "available_at", "updated_at")
SELECT s."tenant_id", 'script', s."id", 'pending', 0, 1, now(), now()
FROM "scripts" s
WHERE s."status" = 'active'
ON CONFLICT ("tenant_id", "source_type", "source_id") DO NOTHING;
