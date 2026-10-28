# Logging, metrics and health

Scope: safe logs, trace/metrics wiring and bounded readiness probes. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| OBS-001 | E | Application logging MUST use common/log (Zap-backed) with request context/trace correlation. No access/refresh token, secret, raw credential-bearing URL or unredacted sensitive payload may be logged. Startup stderr diagnostics are allowed only with safe content. | S logging paths; R redaction |
| OBS-002 | O | Middleware/metrics/traces MUST be initialized through composition/lifecycle. Failure handling MUST not expose internal details or authorize business actions. Health/readiness dependency probes MUST use bounded caller context. | S wiring; T dependency/timeout |
