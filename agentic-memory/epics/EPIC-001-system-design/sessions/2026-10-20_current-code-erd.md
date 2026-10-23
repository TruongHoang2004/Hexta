---
epic: EPIC-001
session: 2026-10-20_current-code-erd
status: in-progress
completed: false
date: 2026-10-20
---

# Current code ERD

## Scope and starting evidence

Owner requested drawing the current code ERD after D4 commit af06cf4 on codex/order-workflows. Preserved unrelated working-tree reconciliation and prior sales evidence. Read canonical architecture/database modules, current GORM schema loader and models, all seven ordered Atlas API migrations and Python AI migration 001.

## Outcome and decisions

Added [current ERD](../../../designs/system/09-current-erd.md) with a complete relationship overview and four detailed domain diagrams, plus editable Mermaid overview source. Includes 21 public business tables and ai_runtime.runs; excludes migration metadata and future conversation/event/exception tables. Solid lines explicitly mean physical migration FKs; dashed lines mean application/domain references, not constraints. Explains account user_id without inventing a users table, nullable draft/customer references, composite workflow FKs, SQL price versus JSON unit_price, uniqueness and physical inventory/legacy AI boundaries. Canonical design index links the ERD.

## Verification and evidence

Parsed CREATE TABLE blocks and applied column overlays from all seven API migrations plus Python 001; checked 22 tables/215 columns and complete domain-diagram coverage. FK relationships inspected from migration SQL; GORM files reviewed for mapping differences. This is a code/migration view, not a fresh live-schema introspection or application test rerun. No credentials or runtime rows read. No schema/API/code changes or external tracker changes.

## Remaining work and handoff

Documentation is local/uncommitted/unmerged. Review rendered Mermaid and deliver/merge separately before marking this session done. D4 browser acceptance and implementation epic completion are unchanged. Update this diagram alongside future schema changes; migration SQL remains authoritative for constraints.
