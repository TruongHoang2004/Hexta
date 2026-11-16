---
epic: EPIC-008
session: 2026-11-16_single-host-deployment
status: in-review
completed: false
date: 2026-11-16
---

# Single-host production deployment for hexta.org

## Scope

The owner has one server (2 vCPU, 16 GB RAM) and asked for a runbook to deploy Hexta to hexta.org. A GitOps design (k3s + Argo CD) was proposed in conversation first; for this hardware the owner proceeded with Docker Compose. Code behavior unchanged.

## Outcome and decisions

- `infrastructure/compose.prod.yml` + `scripts/prod.sh`: Caddy (auto TLS, only public ports 80/443), Postgres 17, Redis, single-node Redpanda in production mode (fsync kept), idempotent topic job, Atlas migration jobs per target, api, ingress, ai, connector, web. Memory limits total ~6.5 GB.
- One origin: `/api/*` → api, `/webhooks/facebook` and `/webhooks/telegram/*` → connector, everything else → web. The connector `/internal/*` control plane is never routed.
- Database isolation: `hexta_api` owns database `hexta`, `hexta_ai` owns schema `ai_runtime`, `hexta_connector` owns separate database `hexta_connector` (Atlas apply on a shared database would see foreign tables).
- New `services/connector/Dockerfile` (non-root); API image now also contains the `ingress` consumer binary.
- Secrets: `infrastructure/prod/.env` (git-ignored) and connector secret files under `infrastructure/prod/secrets/` (git-ignored), generated on the server with `openssl rand`.
- Runbook: `infrastructure/README.deploy.md` (server prep, secrets, first deploy, verification, channels, update, rollback, backups, operations, known limits).

## Verification

| Check | Result |
| --- | --- |
| `docker compose -f infrastructure/compose.prod.yml config --quiet` (dummy env) | PASS |
| Local full-stack run with `DOMAIN=localhost`, throwaway random secrets, isolated project `hexta-prod` | PASS: all images built; migrations and topic job exit 0; api/ai/web/postgres/redis/redpanda healthy; web/api/swagger 200; Telegram and Facebook webhooks routed (403 without credentials); `/internal/connections` 404 at the edge; www and http redirects. Stack and volumes removed afterwards. |
| `scripts/check-agent-instructions.py` | PASS |
| Real server, public DNS/certificates, live provider callbacks, backup restore | NOT_RUN |

## Unresolved work and handoff

- Owner runs the runbook on the real server and reports results.
- Follow-ups: separate migrator/runtime DB roles, publish ai/connector images and a runtime-configurable web API URL so GHCR images can replace on-host builds, metrics/alerting, off-host backups.

## Follow-up: CI-built images and release-PR deploys (2026-11-16)

Owner asked not to build on the VPS, then chose a GitOps-style flow without Kubernetes (option A: deploy on approval).

- PR #74: Docker CI/CD publishes `api`, `connector`, `ai`, `web`, `migrator` as `sha-<7>` on every `main` commit; web image is same-origin (empty `NEXT_PUBLIC_API_URL`); Google login links no longer hardcode `localhost:8080`; `compose.prod.yml` pulls by tag.
- Release PR + agent (stacked on PR #74): CI updates one rolling `release/production` PR that sets `infrastructure/prod/release.env`; merging is the approval. `scripts/deploy-agent.sh` (systemd timer, every minute) checks out the released commit, verifies images, backs up, pulls, starts, health-checks through Caddy, and rolls back automatically. Optional Telegram notices.
- Repository setting "Allow GitHub Actions to create and approve pull requests" is off; until the owner enables it, CI pushes the branch and links a manual PR in the run summary.
- Verification: agent decision paths PASS in a Debian container with a throwaway git remote and stubbed Docker (9 cases). Real CI release PR and the agent on a real server NOT_RUN.
