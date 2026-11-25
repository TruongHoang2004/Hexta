-- Drop index "idx_messaging_conversation_views_waiting" from table: "messaging_conversation_views"
DROP INDEX "idx_messaging_conversation_views_waiting";
-- Modify "messaging_conversation_views" table
ALTER TABLE "messaging_conversation_views" ADD COLUMN "review_pending" boolean NOT NULL DEFAULT false;
-- Create index "idx_messaging_conversation_views_waiting" to table: "messaging_conversation_views"
CREATE INDEX "idx_messaging_conversation_views_waiting" ON "messaging_conversation_views" ("tenant_id", "thread_id" DESC) WHERE (((lifecycle)::text = 'open'::text) AND (assignee_id IS NULL) AND (((control)::text = 'sales'::text) OR review_pending));
-- Modify "messaging_turns" table
ALTER TABLE "messaging_turns" DROP CONSTRAINT "messaging_turns_needs_human_reason_check", ADD CONSTRAINT "messaging_turns_needs_human_reason_check" CHECK ((needs_human_reason)::text = ANY ((ARRAY['customer_requested_staff'::character varying, 'insufficient_information'::character varying, 'ai_error'::character varying, 'business_review'::character varying, 'staff_handoff'::character varying, 'bot_off'::character varying])::text[])), ADD CONSTRAINT "messaging_turns_review_pending_check" CHECK ((NOT review_pending) OR (((lifecycle)::text = 'open'::text) AND ((control)::text = 'system'::text) AND (assignee_id IS NULL) AND (NOT needs_human))), ADD COLUMN "review_pending" boolean NOT NULL DEFAULT false;
-- Modify "tenant_reply_settings" table
ALTER TABLE "tenant_reply_settings" DROP CONSTRAINT "tenant_reply_settings_mode_check", ADD CONSTRAINT "tenant_reply_settings_mode_check" CHECK ((reply_mode)::text = ANY (ARRAY[('off'::character varying)::text, ('suggest'::character varying)::text, ('auto_send'::character varying)::text]));
