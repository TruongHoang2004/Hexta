-- Modify "messaging_conversation_views" table
ALTER TABLE "messaging_conversation_views" ADD COLUMN "needs_human_reason" character varying(32) NULL;
-- Create index "idx_messaging_conversation_views_waiting" to table: "messaging_conversation_views"
CREATE INDEX "idx_messaging_conversation_views_waiting" ON "messaging_conversation_views" ("tenant_id", "thread_id" DESC) WHERE (((lifecycle)::text = 'open'::text) AND ((control)::text = 'sales'::text) AND (assignee_id IS NULL));
-- Modify "messaging_turns" table
ALTER TABLE "messaging_turns" ADD CONSTRAINT "messaging_turns_closed_at_check" CHECK (((lifecycle)::text = 'closed'::text) = (closed_at IS NOT NULL)), ADD CONSTRAINT "messaging_turns_needs_human_reason_check" CHECK ((needs_human_reason)::text = ANY ((ARRAY['customer_requested_staff'::character varying, 'insufficient_information'::character varying, 'ai_error'::character varying, 'business_review'::character varying, 'staff_handoff'::character varying])::text[])), ADD COLUMN "needs_human_reason" character varying(32) NULL, ADD COLUMN "closed_at" timestamptz NULL;
-- Create index "idx_messaging_turn_expiry" to table: "messaging_turns"
CREATE INDEX "idx_messaging_turn_expiry" ON "messaging_turns" ("last_activity_at") WHERE (((lifecycle)::text = 'open'::text) AND (needs_human = false));
