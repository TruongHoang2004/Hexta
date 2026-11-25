# Deploy Hexta to a single server (hexta.org)

Runbook for one Linux host (sized for 2 vCPU / 16 GB RAM) running the whole stack with Docker Compose behind Caddy. Deploys are GitOps-style: CI builds every image and opens a release PR; merging it is the approval, and a deploy agent on the server applies it within a minute. The server never builds. The local demo stack (`compose.demo.yml`) is separate and is not used here.

| Path | Purpose |
| --- | --- |
| `infrastructure/compose.prod.yml` | Production stack: Caddy, Postgres, Redis, Kafka (Confluent `cp-kafka`, KRaft), migrations, api, ingress, ai, connector, web |
| `infrastructure/compose.prod.build.yml` | Local testing only: builds the same images from the working tree |
| `.github/workflows/docker-ci-cd.yml` | Builds and pushes `api`, `connector`, `ai`, `web`, `migrator` to GHCR on every `main` commit, then opens or updates the release PR |
| `infrastructure/prod/release.env` | Desired production image tag (`sha-<7 chars>`); changed only by the release PR |
| `scripts/deploy-agent.sh` + `infrastructure/prod/systemd/` | Server agent and its one-minute systemd timer |
| `infrastructure/prod/Caddyfile` | HTTPS edge and routing |
| `infrastructure/prod/api.yaml` | API configuration (values come from `.env`) |
| `infrastructure/prod/init-db.sh` | One-time database and role creation |
| `infrastructure/prod/.env.example` | Template for `infrastructure/prod/.env` (never commit the real file) |
| `infrastructure/prod/secrets/` | Connector secret files (git-ignored) |
| `scripts/prod.sh` | `docker compose` wrapper that loads `prod/.env` |

## Architecture

```
Internet ──443──► Caddy ──/api/*──────────────► api:8080 ──► Postgres (db hexta), Redis, Kafka, ai:8000
                    │──/webhooks/facebook,
                    │  /webhooks/telegram/*──► connector:8091 ──► Postgres (db hexta_connector), Kafka
                    └──everything else──────► web:3000
ingress (Kafka consumer) ◄──► Kafka ──► Postgres (db hexta)
```

- One origin (`https://hexta.org`) serves the web app and `/api/*`, so auth cookies stay first-party and no CORS is needed.
- Only ports 80 and 443 are published. The connector's `/internal/*` control plane, AI, Postgres, Redis and Kafka are reachable only on the Compose network.
- Each service has its own database login: `hexta_api` owns database `hexta`, `hexta_ai` owns schema `ai_runtime`, and `hexta_connector` owns the separate database `hexta_connector`.
- Migrations run as one-shot Atlas jobs before the services that need them; no service changes schema at startup.
- Memory limits total about 6.5 GB, leaving headroom for the OS.
- Images come from `ghcr.io/truonghoang2004/hexta/<service>:sha-<7-char commit>`, published by CI for every commit on `main`. The web image calls the API on its own origin, so the same image works on any domain.

### Release flow

```
merge to main ─► Docker CI/CD builds 5 images (sha-X) ─► release PR "Release sha-X to production"
                                                            │ you review and merge (= approval)
server: hexta-deploy.timer (every minute) ─► deploy-agent.sh sees release.env = sha-X on origin/main
   ─► checks out commit X (configuration matches the images) ─► all 5 images exist? ─► pg_dumpall backup
   ─► pull ─► up -d --wait ─► https health check ─► success: record sha-X ─► Telegram notice (optional)
                                                └─► failure: back to the previous commit + tag, notify, never retry sha-X
```

- One rolling release PR (`release/production`): a newer `main` build replaces its contents, so merging always ships the newest approved images.
- Merging the release PR does not rebuild anything (`release.env` is excluded from the build trigger).
- The agent's state lives in `infrastructure/prod/state/` (git-ignored): `deployed` = running tag, `failed` = last rejected tag.

## 1. Prerequisites

- A server running Ubuntu 24.04 LTS (or similar) with a public IPv4 address, SSH access and a sudo user.
- DNS for `hexta.org`: `A hexta.org → <server IP>` and `A www.hexta.org → <server IP>` (add `AAAA` records too if the server has IPv6). Wait until `dig +short hexta.org` returns the server IP; Caddy needs this to get certificates.
- Optional: a Meta app (Facebook channel), Google OAuth client, AI provider key. Telegram needs nothing up front; owners add bot tokens in the app.

## 2. Prepare the server

```sh
sudo apt-get update && sudo apt-get -y upgrade
sudo apt-get -y install ca-certificates curl git ufw unattended-upgrades
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker "$USER"   # log out and back in afterwards

# Firewall: SSH and web only.
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 443/udp
sudo ufw enable

# 4 GB swap as memory headroom.
sudo fallocate -l 4G /swapfile && sudo chmod 600 /swapfile
sudo mkswap /swapfile && sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

Docker publishes ports itself (it bypasses `ufw`), which is why only Caddy publishes any.

## 3. Get the code

```sh
sudo mkdir -p /opt/hexta && sudo chown "$USER" /opt/hexta
git clone https://github.com/TruongHoang2004/Hexta.git /opt/hexta
cd /opt/hexta
git checkout main
```

The checkout supplies only configuration (Compose file, Caddyfile, API config, init script); images come from GHCR. For a private repository, use a read-only deploy key.

## 4. Configuration and secrets

Never paste secret values into shell history, chats or tickets. Generate them on the server:

```sh
cd /opt/hexta/infrastructure/prod
cp .env.example .env && chmod 600 .env
for k in POSTGRES_PASSWORD API_DB_PASSWORD AI_DB_PASSWORD CONNECTOR_DB_PASSWORD \
         JWT_ACCESS_TOKEN_SECRET JWT_REFRESH_TOKEN_SECRET AI_INTERNAL_TOKEN CONNECTOR_MANAGEMENT_TOKEN; do
  sed -i "s|^$k=.*|$k=$(openssl rand -hex 32)|" .env
done
sed -i 's|^ACME_EMAIL=.*|ACME_EMAIL=<your email>|' .env
```

Then edit `.env` for the optional settings: `GOOGLE_CLIENT_*`, `META_APP_ID`, and `DEPLOY_NOTIFY_TELEGRAM_TOKEN` / `DEPLOY_NOTIFY_TELEGRAM_CHAT_ID` for deploy notifications. AI stays off (`AI_PROVIDER=disabled`) until you enable it as described in [AI settings](#ai-settings). Use hex values for passwords because they are embedded in database URLs.

### AI settings

Owner decision 2026-11-26: production AI is **off** until it is fully configured. With the defaults
(`AI_PROVIDER=disabled`) the bot sends only a unique exact match of an active script and hands every other message
to staff. No paid model, Jev or embedding call is made; knowledge sync keeps running with the offline hashing
embeddings (free). The `ai` service logs one startup line, `AI runtime DISABLED … provider=disabled`.

To enable AI, set **all** of these in `prod/.env` together (Claude answers, Jev routing and Gemini embeddings are
one feature: without Jev the Claude classifier never routes to knowledge answers):

```sh
AI_PROVIDER=anthropic
AI_MODEL=claude-sonnet-5-5
ANTHROPIC_API_KEY=<key>
AI_JEV_ENABLED=true
TYPESAFE_API_KEY=<key>
AI_EMBEDDING_PROVIDER=gemini
GEMINI_API_KEY=<key>
```

Apply and check, then re-embed the knowledge with the new model:

```sh
./scripts/prod.sh up -d --wait ai
./scripts/prod.sh logs ai | grep "AI runtime"   # AI runtime enabled: provider=anthropic model=claude-sonnet-5-5 router=jev embeddings=gemini-embedding-001
./scripts/prod.sh exec ingress ./knowledge-resync -all-tenants
```

Configuration problems never stop the service (so they never fail a deploy); they are logged and AI stays safe:

| Setting | Result |
| --- | --- |
| Real `AI_PROVIDER` without its key, without `AI_MODEL`, with another provider's model (`anthropic` + `gpt-4o-mini`, `openai` + `claude-…`), invalid numbers, or an unknown provider | Error `AI disabled: <reason>`; AI disabled as above |
| `AI_JEV_ENABLED=true` without `TYPESAFE_API_KEY`, or invalid `AI_JEV_*` values | Error `Jev router off: …`; Jev stays off |
| `AI_EMBEDDING_PROVIDER=gemini` without `GEMINI_API_KEY` | Error; offline hashing embeddings |
| AI disabled, or `AI_PROVIDER=fake` (offline simulation, demo/tests only), while Jev or Gemini embeddings are configured | Warning; both stay off (no paid call); `fake` runs are labelled simulated (`AI runtime SIMULATED …`) |
| Invalid `AI_RAG_*` values | Error `Knowledge retrieval disabled: …`; no knowledge answers (handed to staff) |

Turn AI off again with `AI_PROVIDER=disabled` and `./scripts/prod.sh up -d --wait ai`. Retrieval only uses chunks
embedded with the current embedding model, so after any change of `AI_EMBEDDING_PROVIDER`/`AI_EMBEDDING_MODEL` run
`./scripts/prod.sh exec ingress ./knowledge-resync -all-tenants` (or `-tenant <workspace id>`); until the sync job
has re-sent every source, answers hand off for lack of knowledge. `./knowledge-resync -status` prints the job counts
(pending, stuck, failed, done).

Connector secret files (the connector runs as uid 10001):

```sh
cd /opt/hexta/infrastructure/prod/secrets
grep '^CONNECTOR_MANAGEMENT_TOKEN=' ../.env | cut -d= -f2- | tr -d '\n' > connector_management_token
openssl rand -hex 32 | tr -d '\n' > connector_credential_key
openssl rand -hex 24 | tr -d '\n' > meta_verify_token
openssl rand -hex 24 | tr -d '\n' > meta_app_secret   # replace with the real Meta app secret if Facebook is used
sudo chown 10001:10001 connector_* meta_*
sudo chmod 400 connector_* meta_*
```

Back up `.env` and `secrets/` to a password manager or another offline store. Losing `connector_credential_key` makes stored Facebook Page tokens and Telegram Bot Tokens unreadable, and owners must reconnect.

## 5. First deployment

One-time GitHub setup:
1. Repository → Settings → Actions → General → Workflow permissions: enable **Allow GitHub Actions to create and approve pull requests**. Without it, CI still pushes the `release/production` branch and its run summary links to open the PR by hand.
2. After a `main` commit's **Docker CI/CD** run succeeds, review and merge the **Release sha-… to production** PR. Before the first merge, `release.env` is empty and the agent does nothing.

On the server, install the agent (runs as `deploy`) and watch it:

```sh
sudo cp /opt/hexta/infrastructure/prod/systemd/hexta-deploy.* /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now hexta-deploy.timer
journalctl -u hexta-deploy -f        # "deployed sha-…" when done
cd /opt/hexta && ./scripts/prod.sh ps -a
```

The agent checks out the released commit in detached mode; keep the checkout free of local edits (secrets, `.env`, `state/` and `backups/` are git-ignored). To deploy without waiting for the timer: `sudo systemctl start hexta-deploy`.

Expected:
- `postgres-collation-guard`, `postgres-extensions`, `migrate-api`, `migrate-ai`, `migrate-connector` and `kafka-topics` show `exited (0)`.
- `postgres`, `redis`, `kafka`, `api`, `ai`, `web` are `healthy`.
- `connector`, `ingress` and `caddy` are `running`.

The first start also issues TLS certificates; check with `./scripts/prod.sh logs caddy | grep -i certificate`.

## 6. Verify

```sh
curl -s -o /dev/null -w '%{http_code}\n' https://hexta.org/                         # 200
curl -s -o /dev/null -w '%{http_code}\n' https://hexta.org/api/v1/health            # 200
curl -s -o /dev/null -w '%{http_code}\n' -X PUT https://hexta.org/internal/connections  # 404 (never the connector)
curl -s -o /dev/null -w '%{http_code}\n' -X POST -d '{}' \
  https://hexta.org/webhooks/telegram/00000000-0000-0000-0000-000000000000          # 403
```

Then in a browser: open `https://hexta.org/register`, create the owner account and a workspace.

If the agent logs `image … not available yet` for a tag whose CI run succeeded, that GHCR package is private. Either make it public (GitHub → the package → Package settings → Change visibility), or on the server run `docker login ghcr.io -u <github user>` with a personal access token that has only `read:packages`.

## 7. Connect channels

- **Telegram:** Settings → Channels → paste a @BotFather token. The connector registers `https://hexta.org/webhooks/telegram/<connection id>` automatically. Message the bot from a private chat and check the Inbox.
- **Facebook:**
  1. In the Meta app, set the webhook callback to `https://hexta.org/webhooks/facebook`, using the value in `secrets/meta_verify_token` as the verify token.
  2. Set the valid OAuth redirect to `https://hexta.org/channels/facebook/callback`.
  3. Put the real app secret in `secrets/meta_app_secret` and `META_APP_ID` in `.env`.
  4. Run `./scripts/prod.sh up -d --wait connector api` (configuration changes in `.env` or `secrets/` are applied by hand; the agent only reacts to new releases).
- **Google login:** set the authorized redirect URI to `https://hexta.org/api/v1/auth/google/callback`.

## 8. Update to a new version

Merge the release PR. Nothing to run on the server: within about a minute the agent backs up the database (`infrastructure/prod/backups/pre-<tag>-*.sql.gz`, kept 14 days), deploys and health-checks. Follow it with `journalctl -u hexta-deploy -f` or the Telegram notices.

Migration jobs rerun on every deploy; they apply only new migrations and exit 0 when there are none.

## 9. Roll back

- **Automatic:** a deploy that fails to start or fails the health check is rolled back to the previous release; the failed tag is not retried until `release.env` changes.
- **Deliberate (code only, no new migration):** revert the release PR's merge commit on `main` (GitHub "Revert" button). The agent deploys the previous tag within a minute.
- **Emergency, from the server:** `sudo systemctl stop hexta-deploy.timer`, then `git checkout <previous commit> && HEXTA_IMAGE_TAG=sha-<previous> ./scripts/prod.sh up -d --wait && echo sha-<previous> > infrastructure/prod/state/deployed`. Re-enable the timer only after `release.env` on `main` points to a good tag.
- **Across the pgvector switch (section 10a):** never roll back to a release from before the switch with the commands above: its Alpine Postgres image would start on the glibc-indexed volume and corrupt text indexes again. Use the rollback in section 10a (restore the pre-switch volume copy or dump). The deploy agent refuses such a release while the volume carries the marker.
- **With migrations:** migrations are forward-only and additive by policy, so the previous code normally still runs on the newer schema. Roll back the code as above. Restore a backup (section 10) only for a destructive migration, and accept losing data written since that backup.

## 10. Backups

A nightly dump, kept 14 days locally. Copy dumps off the server too (another machine or object storage); a backup on the same disk does not survive disk loss.

```sh
mkdir -p ~/hexta-backups
crontab -e
# 03:15 every night
15 3 * * * cd /opt/hexta && ./scripts/prod.sh exec -T postgres pg_dumpall -U postgres | gzip > ~/hexta-backups/hexta-$(date +\%F).sql.gz && find ~/hexta-backups -name '*.sql.gz' -mtime +14 -delete
```

**Test a restore before you rely on it.** On a scratch machine with the same `.env` and `secrets/`:
1. Start Postgres with `./scripts/prod.sh up -d postgres`.
2. Run `gunzip -c <dump> | ./scripts/prod.sh exec -T postgres psql -U postgres`.
3. Run `./scripts/prod.sh up -d --wait`.

Kafka holds only in-flight events. The connector and API keep their durable outboxes in Postgres, so a lost Kafka volume is recreated by `kafka-topics` and the outboxes republish pending events.

## 10a. One-time switch to the pgvector Postgres image

The release that introduces AI knowledge answers changes `postgres` from `postgres:17.7-alpine3.23` (musl) to
`pgvector/pgvector:0.8.1-pg17-bookworm` (Debian 12, glibc) and adds two one-shot jobs: `postgres-collation-guard`
(before `postgres`) and `postgres-extensions` (`CREATE EXTENSION vector` as the superuser, before `migrate-ai`).
Both images are PostgreSQL 17, so the data directory is reused as is, but **text sort order differs between musl
and glibc**: every B-tree index on `text`/`varchar` columns (including the `varchar(36)` IDs used as primary keys)
is silently wrong under glibc until it is rebuilt. Lookups can miss existing rows and unique checks can let
duplicates in. Alpine-created databases record no collation version (`datcollversion` is NULL), so PostgreSQL
prints no warning.

**The switch cannot happen unattended.** `postgres-collation-guard` mounts the data volume read-only and lets
`postgres` start only when the volume is empty (a fresh install; `init-db.sh` then writes the marker under glibc)
or carries `$PGDATA/HEXTA_GLIBC_COLLATION_OK`, which only the steps below write after REINDEX and amcheck. If the
deploy agent reaches this release first, Compose replaces the postgres container but never starts it, the guard
logs `REFUSING TO START POSTGRES … section 10a`, the deploy fails and the agent rolls back to the previous
release. Expect a short outage in that case; the data is not touched. Do not delete or hand-write the marker.

**Never start the Alpine image again on a switched volume.** After step 6 the only way back is the rollback below
(the pre-switch volume copy or the dump). Releases from before this change have no guard, so Compose cannot stop
them; the deploy agent from this release on refuses to deploy or roll back to such a release while the volume
carries the marker (`release … rejected: its Alpine (musl) Postgres image must not start …`).

### Before the window

1. Note the release commit X: the `main` commit whose images the release PR names ("Release sha-X to
   production"; seven hex characters). Every command below uses it explicitly; do not source `release.env` from
   another commit (commit X itself still holds the previous tag).
2. Pause automatic deploys, then merge the release PR (the agent stays idle while paused):
   `sudo systemctl stop hexta-deploy.timer` and wait until `systemctl is-active hexta-deploy` prints `inactive`.
3. AI settings: nothing is required, AI stays off by default (see [AI settings](#ai-settings)). An existing
   `.env` copied from the older template contains `AI_PROVIDER=fake` and `AI_MODEL=gpt-4o-mini`: change them to
   `AI_PROVIDER=disabled` and an empty `AI_MODEL=`, or, to enable AI now, edit **both** `AI_PROVIDER` and
   `AI_MODEL` together with the keys listed there. A leftover `gpt-4o-mini` with `anthropic` keeps AI disabled.
4. Preflight on the running (old) stack:

```sh
cd /opt/hexta
mkdir -p ~/hexta-backups
docker pull pgvector/pgvector:0.8.1-pg17-bookworm
# Superuser login over TCP, as postgres-extensions uses it (POSTGRES_PASSWORD from .env). Prints 1.
./scripts/prod.sh exec -T postgres sh -c 'PGPASSWORD="$POSTGRES_PASSWORD" psql -h 127.0.0.1 -U postgres -d hexta -Atc "select 1"'
# Only if that fails with "password authentication failed": align the stored password with .env, then re-check.
./scripts/prod.sh exec -T postgres sh -c 'psql -U postgres -q -v pw="$POSTGRES_PASSWORD"' <<'SQL'
ALTER ROLE postgres PASSWORD :'pw';
SQL
# Disk: the dump goes to ~/hexta-backups, the volume copy and REINDEX use the Docker disk. Free space on each must
# exceed the data size; keep at least 2.5x the data size free on the Docker disk.
./scripts/prod.sh exec -T postgres du -sh /var/lib/postgresql/data
df -h /var/lib/docker ~/hexta-backups
```

### Maintenance window (writers stopped, a few minutes of downtime)

```sh
cd /opt/hexta
X=<7-char commit>                      # from step 1
PG=pgvector/pgvector:0.8.1-pg17-bookworm
# 1. Stop every writer; keep postgres running for the dump.
./scripts/prod.sh stop caddy web api messaging-command ingress connector ai
# 2. Logical backup, verified; copy it off the server before continuing.
F=~/hexta-backups/pre-pgvector-$(date +%F-%H%M).sql.gz
./scripts/prod.sh exec -T postgres pg_dumpall -U postgres | gzip > "$F" && gunzip -t "$F" && ls -lh "$F"
# 3. File-level copy of the stopped volume for a fast rollback. The copy volume must be new and empty.
./scripts/prod.sh stop postgres
docker volume create hexta-prod_postgres_data_alpine17
docker run --rm --network none -v hexta-prod_postgres_data:/from:ro -v hexta-prod_postgres_data_alpine17:/to \
  --entrypoint sh "$PG" -c '[ -z "$(ls -A /to)" ] || { echo "copy volume not empty: stop" >&2; exit 1; }; cp -a /from/. /to/ && du -s /from /to'
# 4. Check out the release commit and pin its tag explicitly.
git fetch origin && git checkout --detach "$(git rev-parse --verify "$X^{commit}")"
export HEXTA_IMAGE_TAG=sha-$X
./scripts/prod.sh pull --quiet
# 5. Rebuild on glibc in an isolated container (no network: no application can connect; Compose stays blocked).
docker run -d --name hexta-pg-reindex --network none -v hexta-prod_postgres_data:/var/lib/postgresql/data "$PG"
until docker exec hexta-pg-reindex pg_isready -U postgres -q; do sleep 1; done
P="docker exec -i -u postgres hexta-pg-reindex psql -U postgres -v ON_ERROR_STOP=1"
$P -Atc "SELECT datname, datcollate, coalesce(datcollversion, 'NULL') FROM pg_database ORDER BY 1"
# 5a. REINDEX every database, template0 included (opened only for this).
$P -qc "ALTER DATABASE template0 ALLOW_CONNECTIONS true"
for db in $($P -Atc "SELECT datname FROM pg_database ORDER BY 1"); do
  $P -d "$db" -qc "REINDEX DATABASE \"$db\"" -c "REINDEX SYSTEM \"$db\"" || { echo "REINDEX failed in $db: STOP, restore"; break; }
  echo "reindexed $db"
done
# 5b. Refresh the recorded collation version only where one is recorded (Alpine databases record none: no-op).
for db in $($P -Atc "SELECT datname FROM pg_database WHERE datcollversion IS NOT NULL"); do
  $P -qc "ALTER DATABASE \"$db\" REFRESH COLLATION VERSION"
done
# 5c. Verify every valid B-tree index with amcheck (heap included), then remove amcheck again.
for db in $($P -Atc "SELECT datname FROM pg_database ORDER BY 1"); do
  n=$($P -d "$db" -qAt <<'SQL'
SET client_min_messages = warning;
CREATE EXTENSION IF NOT EXISTS amcheck;
SELECT count(bt_index_check(c.oid, true)) FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
  JOIN pg_am am ON am.oid = c.relam WHERE am.amname = 'btree' AND i.indisvalid AND c.relpersistence <> 't';
DROP EXTENSION amcheck;
SQL
) || { echo "amcheck FAILED in $db: STOP, restore"; break; }
  echo "amcheck $db: $n indexes OK"
done
$P -qc "ALTER DATABASE template0 ALLOW_CONNECTIONS false"
# 5d. Only when every database printed "reindexed" and "indexes OK": write the marker, stop the container.
docker exec -u postgres hexta-pg-reindex sh -c 'echo "glibc collation: REINDEX + amcheck by $(postgres --version) on $(date -u +%FT%TZ) (README.deploy.md 10a)" > "$PGDATA/HEXTA_GLIBC_COLLATION_OK"'
docker stop hexta-pg-reindex && docker rm hexta-pg-reindex
# 6. Start the stack: guard (marker found), postgres, postgres-extensions (vector), migrate-api / migrate-ai, apps.
./scripts/prod.sh up -d --wait --remove-orphans
./scripts/prod.sh logs --no-log-prefix postgres-collation-guard | tail -1
./scripts/prod.sh exec -T postgres psql -U postgres -d hexta -Atc "SELECT extname, extversion FROM pg_extension WHERE extname = 'vector'"
# 7. Record the deployed tag yourself, then resume the agent (it sees release.env = sha-X = deployed: no-op).
echo "$HEXTA_IMAGE_TAG" > infrastructure/prod/state/deployed
sudo systemctl start hexta-deploy.timer
```

Deploy this release by hand as above; do not let the agent perform the first glibc start (an agent from an older
release would roll back to the Alpine image if a later health check failed). A `REINDEX` failure on a unique index
means duplicate rows were written under a mismatched order, and an amcheck error means an index is still wrong: in
both cases stop and roll back (below) instead of deleting rows, and do not write the marker.

Verify: section 6 health checks; `SELECT state, count(*) FROM knowledge_sync_jobs GROUP BY 1` in `hexta` drains
from `pending` to `done` (migration 20261127100001 enqueues every active product and script once; `failed` rows
carry a reason in `last_error`). `./scripts/prod.sh logs messaging-command | grep "Knowledge sync jobs need attention"`
reports failed or stuck jobs every ten minutes while there are any.

### Rollback of the switch

Never start the previous release on the switched volume, and never copy into a non-empty volume. Restore the
musl copy into a recreated, empty volume (or restore the dump into an empty volume as in section 10):

```sh
cd /opt/hexta
sudo systemctl stop hexta-deploy.timer
PG=pgvector/pgvector:0.8.1-pg17-bookworm
# Keep what was written since the switch (to re-enter by hand later), then remove the switched volume.
./scripts/prod.sh exec -T postgres pg_dumpall -U postgres | gzip > ~/hexta-backups/post-pgvector-$(date +%F-%H%M).sql.gz
./scripts/prod.sh down                                   # never down -v
docker volume rm hexta-prod_postgres_data
# The previous release recreates the empty volume (with its Compose labels) without starting postgres.
git checkout --detach <previous release commit>
export HEXTA_IMAGE_TAG=sha-<previous>
./scripts/prod.sh up --no-start postgres
docker run --rm --network none -v hexta-prod_postgres_data_alpine17:/from:ro -v hexta-prod_postgres_data:/to \
  --entrypoint sh "$PG" -c '[ -z "$(ls -A /to)" ] || { echo "target volume not empty: stop" >&2; exit 1; }; cp -a /from/. /to/'
./scripts/prod.sh up -d --wait --remove-orphans
echo "$HEXTA_IMAGE_TAG" > infrastructure/prod/state/deployed
```

Then revert the release PR on `main` before re-enabling the timer, or the agent deploys sha-X again (the guard stops
it, since the restored copy has no marker). The new tables and the `vector` extension are additive; the previous
code ignores them. Delete `hexta-prod_postgres_data_alpine17` only after the new release has run cleanly for a while.

## 11. Operations

| Task | Command |
| --- | --- |
| Status | `./scripts/prod.sh ps -a`; running tag: `cat infrastructure/prod/state/deployed` |
| Deploy agent | `systemctl list-timers hexta-deploy.timer`, `journalctl -u hexta-deploy --since today` |
| Pause automatic deploys | `sudo systemctl stop hexta-deploy.timer` (start it again to resume) |
| Logs | `./scripts/prod.sh logs -f --tail=200 api` (or `connector`, `ingress`, `ai`, `web`, `caddy`) |
| Restart one service | `./scripts/prod.sh restart connector` |
| Resource use | `docker stats --no-stream` |
| Disk cleanup (old images) | `docker image prune -f` |
| Postgres shell | `./scripts/prod.sh exec postgres psql -U postgres -d hexta` |
| AI knowledge re-sync (embedding model switch) | `./scripts/prod.sh exec ingress ./knowledge-resync -all-tenants`; counts only: `… ./knowledge-resync -status` |

Never run `down -v` on the server: it deletes the database volume. Never print `.env`, `secrets/` or `docker compose config` output into logs or chats; it contains resolved secrets.

## Shared host behind an existing reverse proxy

Use this when another app's nginx already owns ports 80/443 (for example the eztech host, where a shared nginx serves other sites on the Docker network `infra-net`). Caddy then publishes no port, joins that network as `hexta-edge` and keeps Hexta's routing over plain HTTP; nginx terminates TLS. The overlay also lowers the Kafka heap for ~4 GB hosts.

1. DNS on Cloudflare: `A app.hexta.org → <server IP>`, proxied. Set `DOMAIN=app.hexta.org` and `HEXTA_EDGE_NETWORK=infra-net` in `prod/.env`; `ACME_EMAIL` must still be non-empty but is unused.
2. Start the stack with the overlay on every command (`prod.sh` takes extra `-f` files before the Compose command):
   ```sh
   echo sha-<7 chars> > infrastructure/prod/state/deployed
   ./scripts/prod.sh -f infrastructure/compose.prod.shared-edge.yml pull
   ./scripts/prod.sh -f infrastructure/compose.prod.shared-edge.yml up -d --wait
   ```
3. Origin certificate: Cloudflare → SSL/TLS → Origin Server → Create certificate (covers `hexta.org` and `*.hexta.org`, 15 years). Save the certificate and key on the server as `cert.pem`/`key.pem` (key mode 600) in the directory the proxy mounts at `/etc/letsencrypt/hexta-origin/`, then set SSL/TLS mode to **Full (strict)**. Cloudflare serves the public certificate, so certbot is not needed.
4. Copy [prod/nginx-shared-edge.conf](prod/nginx-shared-edge.conf) into the proxy's `conf.d`, run `nginx -t`, and reload.

The deploy agent does not pass the overlay; until it does, update by hand (step 2 with the new tag). Upstreams in `Caddyfile.shared-edge` are container names (`hexta-prod-<service>-1`) because the shared network can hold other apps' `api`/`postgres`/`redis` aliases.

## Known limits

- One host, no high availability. Disk loss means restoring from an off-server backup.
- Each service uses one login for migrations and runtime; separate migrator/runtime roles are a follow-up.
- No metrics or alerting are wired beyond deploy notices; use `ps`, `logs` and `docker stats`.
- The pgvector switch (section 10a) was rehearsed locally on 2026-11-26 with real containers (Compose v5.6): fresh volume starts, an Alpine-created volume is refused by the guard, and the runbook (REINDEX, amcheck, marker) plus its rollback succeed. NOT_RUN on a server.
- Verified locally on 2026-11-16 by running this stack with `DOMAIN=localhost` and images built through `compose.prod.build.yml` (same Dockerfiles and build args as CI), covering routing, health, migrations, topic creation, and register/login on one origin. The deploy agent was tested in a Linux container against a throwaway git remote with stubbed Docker/Compose: empty release, invalid tag, images not yet published, first deploy, no-op repeat, failed deploy with automatic rollback and backup, no retry of a failed tag, and recovery on the next release. NOT_RUN: the first CI publish of the `connector` and `ai` images and of the release PR, the agent against real Docker on a server, public certificates, DNS, live Facebook/Telegram callbacks and backup restore.
