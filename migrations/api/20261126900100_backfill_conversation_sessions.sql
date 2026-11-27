-- Data backfill for conversation sessions (custom migration; no schema change).
-- Every turn closed before close_cause existed was closed by the 24-hour inactivity expiry.
UPDATE "messaging_turns" SET "close_cause" = 'expired' WHERE "lifecycle" = 'closed' AND "close_cause" IS NULL;
-- The current assignee is the last known handler of existing turns.
UPDATE "messaging_turns" SET "last_handler_id" = "assignee_id" WHERE "assignee_id" IS NOT NULL AND "last_handler_id" IS NULL;
-- Keep the rebuildable conversation read model in step with its source turns.
UPDATE "messaging_conversation_views" AS v SET "closed_at" = t."closed_at", "close_cause" = t."close_cause", "outcome" = t."outcome"
FROM "messaging_turns" AS t WHERE t."tenant_id" = v."tenant_id" AND t."id" = v."turn_id";
