-- Create "messaging_ingress_quarantine" table
CREATE TABLE "messaging_ingress_quarantine" (
  "identity" text NOT NULL,
  "reason" text NOT NULL,
  "received_at" timestamptz NOT NULL,
  PRIMARY KEY ("identity")
);
