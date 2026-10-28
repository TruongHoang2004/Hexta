# Verification and future hook contract

Status: review specification only. No hook/checker is implemented or activated. The rule catalog applies to the entire API codebase; delivery policy for existing failures is an owner decision, not an excuse to label them PASS.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| QA-001 | E | Every changed module MUST run applicable formatting, vet, tests and contract/migration checks. Results MUST include exact scope, command/outcome and skipped prerequisites. No absent integration test may be described as verified. | R evidence; T checks |
| QA-002 | P | Whole-codebase assessment MUST report each applicable rule as PASS/FAIL/NOT_RUN/NOT_APPLICABLE with evidence and source locations. Static pattern matching alone MUST NOT establish semantic authorization/transaction correctness. | R report schema/claims |
| QA-003 | P | A rule-check result MUST identify source revision/working diff, rule version/content and test context. Any subsequent source/rule change MUST invalidate affected evidence. No previous green result may cover new edits implicitly. | T checker freshness when implemented |
| QA-004 | P | Future hooks MUST inspect all API source for rule compliance, reporting pre-existing and introduced violations separately. They MUST NOT silently ignore old failures, modify rules/baselines, apply migrations to application DBs, send messages or add business privileges. | R hook design; T safety/freshness |
| QA-005 | P | A missing DB/tool/permission or hook timeout MUST be NOT_RUN with a reason. Continuation MUST be bounded; after repeated environmental failure, report unresolved checks instead of endless retries or a false PASS. | T missing prerequisites/loop handling |
| QA-006 | E | Behavioral changes MUST have meaningful tests for affected authorization, validation, rollback, conflicts/replays and failure paths. Cross-layer integration tests are allowed. Tests MUST NOT use runtime DDL/AutoMigrate or production data. | R test relevance; T isolated execution |

## Checks available now

From repository root:

```sh
sh scripts/check-gofmt.sh
go vet ./services/api/... ./packages/shared/...
go test -race ./services/api/... ./packages/shared/...
make migrate-validate
```

- Configure `API_TEST_DATABASE_URL` privately and ensure Atlas is available for DB scenarios. Fixtures create disposable databases. Absence causes skips, not DB acceptance.
- Endpoint changes require `make swagger` and generated-contract verification. Do not regenerate without a contract change merely to manufacture evidence.
- Schema changes require the Atlas integrity/drift/fresh/upgrade checks described in MIG-008, against explicit disposable targets. Root CI currently contains desired-history drift and isolated AI checks; those are not automatically run by a documentation edit.
- Provider/RPC changes require adapter/contract tests with fake/local fixtures first. No paid/live side effects are implied.
- Repository architecture AST/import checks are proposed, not existing commands.

## Required reasoning pass

For a scoped implementation, apply the following checks only where relevant to the changed use-case paths and selected rules. A full-codebase audit performs these checks sequentially per module; it does not require every rule body in one context:

1. Trace route → controller/validator → aggregate mapper → service → repository/read adapter → infrastructure and confirm ownership. Check each aggregate mapper for purity and boundary-specific imports.
2. Trace actor/tenant/resource scope, including negative and revoked-state paths.
3. Identify the transaction, every participating write/audit and the version/replay behavior; inspect partial failure.
4. Trace public response/error/log data for ORM coupling and sensitive data.
5. Review migration source/history/provenance and test isolation when schema is affected.
6. For cache changes, inspect policy scope, DB-only authorization, after-commit invalidation, aliases, fill fencing, outage TTL bounds and measured benefit.
7. Report failures and unknowns by rule ID; do not infer a rule is satisfied merely because the function/file is named correctly.

## Proposed assessment record

```text
Source: commit + working-tree fingerprint
Rules: version + content fingerprint
Rule: REP-003
Result: FAIL
Evidence: services/api/internal/repository/session_repository.go:29
Finding: request context is not attached to the query
Classification: pre-existing | introduced | worsened
Checks: command/outcome or NOT_RUN reason
```

Also record the selected rule modules/IDs and why path/behavior triggers applied. Unselected scope is not checked. Full audits must retain coverage between batches so selective context does not silently narrow the audit.

This is the interface for a later hook/report implementation, not a requirement to create another global session-context file or an assertion that every rule can be machine-proven.
