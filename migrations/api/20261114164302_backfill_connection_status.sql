-- Connections disabled before status existed were operator-paused bindings; keep them non-sending.
UPDATE "messaging_connections" SET "status" = 'paused' WHERE NOT "enabled" AND "status" = 'active';
