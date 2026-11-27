-- Modify "messaging_turns" table
ALTER TABLE "messaging_turns" DROP CONSTRAINT "messaging_turns_close_lifecycle_check", ADD CONSTRAINT "messaging_turns_close_lifecycle_check" CHECK (((lifecycle)::text = 'closed'::text) = (close_cause IS NOT NULL));
