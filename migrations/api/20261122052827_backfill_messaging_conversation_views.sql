-- Seed the conversation read model from the authoritative thread and open-turn rows.
INSERT INTO "messaging_conversation_views" (
  "tenant_id", "thread_id", "external_key", "display_name", "thread_created_at", "thread_updated_at",
  "connection_id", "recipient_id", "turn_id", "lifecycle", "control", "assignee_id", "needs_human",
  "context_revision", "control_generation", "turn_version", "last_activity_at"
)
SELECT th."tenant_id", th."id", th."external_key", th."display_name", th."created_at", th."updated_at",
  th."connection_id", th."recipient_id", t."id", t."lifecycle", t."control", t."assignee_id", t."needs_human",
  t."context_revision", t."control_generation", t."version", t."last_activity_at"
FROM "messaging_threads" th
JOIN "messaging_turns" t ON t."tenant_id" = th."tenant_id" AND t."thread_id" = th."id" AND t."lifecycle" = 'open'
ON CONFLICT ("tenant_id", "thread_id") DO NOTHING;
