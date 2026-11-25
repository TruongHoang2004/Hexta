-- Modify "domain_events" table
ALTER TABLE "domain_events" DROP CONSTRAINT "chk_domain_events_state", ADD CONSTRAINT "chk_domain_events_state" CHECK ((state)::text = ANY (ARRAY[('pending'::character varying)::text, ('leased'::character varying)::text, ('published'::character varying)::text, ('failed'::character varying)::text, ('cancelled'::character varying)::text]));
