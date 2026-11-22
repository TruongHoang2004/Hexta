-- Rows still pending under the pre-release topic names move to their per-event topics so the new check holds.
UPDATE "connector"."outbound_events" SET "topic" = 'hexta.delivery.send-completed.v1' WHERE "topic" = 'hexta.outbound.result.v1';
UPDATE "connector"."outbound_events" SET "topic" = 'hexta.connection.health-changed.v1' WHERE "topic" = 'hexta.connector.health.v1';
-- Modify "outbound_events" table
ALTER TABLE "connector"."outbound_events" DROP CONSTRAINT "outbound_events_topic_check", ADD CONSTRAINT "outbound_events_topic_check" CHECK (topic = ANY (ARRAY['hexta.delivery.send-completed.v1'::text, 'hexta.connection.health-changed.v1'::text]));
