# Outbound client adapters

Scope: external HTTP/RPC client behavior, timeouts and effect retries. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| INF-002 | O | Outbound clients MUST use caller context, bounded timeouts, close response bodies and validate status/decoding. Provider/protocol details MUST stay in adapters and be mapped/sanitized before business/public errors. | S call/close paths; T timeout/status/malformed response |
| INF-003 | P | External effects MUST NOT be automatically retried unless their idempotency/reconciliation semantics are defined. Client callbacks and streaming errors MUST propagate; no ignored error may claim completion. | R retry/effect analysis; T interruption/failure |
