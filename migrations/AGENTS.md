# Database migrations

Read [the canonical Atlas policy](../rules/api-service/05-atlas-migrations.md); it owns detailed requirements and provenance. API/public and AI/ai_runtime have separate desired SQL and versioned histories. Do not duplicate that policy here or silently approve rules labeled proposed.

Changes to desired SQL, history, root `atlas.hcl`, migration scripts/jobs and test provisioning all select this scope. Identify the owning service, affected models and existing-data upgrade before generation. Use the root Makefile/Atlas workflow; runtime startup and tests must not introduce another schema migration mechanism.

Check history integrity and desired/history agreement, plus fresh/upgrade behavior when applicable on explicitly disposable targets. Checks must not apply to normal application databases. Report unavailable DB/Atlas prerequisites as NOT_RUN. Review destructive/locking effects and recovery separately from checksum success.
