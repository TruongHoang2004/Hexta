# Deploy Hexta to a single server (hexta.org)

Runbook for one Linux host (sized for 2 vCPU / 16 GB RAM) running the whole stack with Docker Compose behind Caddy. CI builds every application image; the server only pulls and runs them. The local demo stack (`compose.demo.yml`) is separate and is not used here.

| Path | Purpose |
| --- | --- |
| `infrastructure/compose.prod.yml` | Production stack: Caddy, Postgres, Redis, Redpanda, migrations, api, ingress, ai, connector, web |
| `infrastructure/compose.prod.build.yml` | Local testing only: builds the same images from the working tree |
| `.github/workflows/docker-ci-cd.yml` | Builds and pushes `api`, `connector`, `ai`, `web`, `migrator` to GHCR on every `main` commit |
| `infrastructure/prod/Caddyfile` | HTTPS edge and routing |
| `infrastructure/prod/api.yaml` | API configuration (values come from `.env`) |
| `infrastructure/prod/init-db.sh` | One-time database and role creation |
| `infrastructure/prod/.env.example` | Template for `infrastructure/prod/.env` (never commit the real file) |
| `infrastructure/prod/secrets/` | Connector secret files (git-ignored) |
| `scripts/prod.sh` | `docker compose` wrapper that loads `prod/.env` |

## Architecture

```
Internet ──443──► Caddy ──/api/*──────────────► api:8080 ──► Postgres (db hexta), Redis, Redpanda, ai:8000
                    │──/webhooks/facebook,
                    │  /webhooks/telegram/*──► connector:8091 ──► Postgres (db hexta_connector), Redpanda
                    └──everything else──────► web:3000
ingress (Kafka consumer) ◄──► Redpanda ──► Postgres (db hexta)
```

- One origin (`https://hexta.org`) serves the web app and `/api/*`, so auth cookies stay first-party and no CORS is needed.
- Only ports 80 and 443 are published. The connector's `/internal/*` control plane, AI, Postgres, Redis and Kafka are reachable only on the Compose network.
- Each service has its own database login: `hexta_api` owns database `hexta`, `hexta_ai` owns schema `ai_runtime`, and `hexta_connector` owns the separate database `hexta_connector`.
- Migrations run as one-shot Atlas jobs before the services that need them; no service changes schema at startup.
- Memory limits total about 6.5 GB, leaving headroom for the OS.
- Images come from `ghcr.io/truonghoang2004/hexta/<service>:sha-<7-char commit>`, published by CI for every commit on `main`. `HEXTA_IMAGE_TAG` in `.env` selects the version. The web image calls the API on its own origin, so the same image works on any domain.

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
sed -i "s|^HEXTA_IMAGE_TAG=.*|HEXTA_IMAGE_TAG=sha-$(git -C /opt/hexta rev-parse --short=7 HEAD)|" .env
```

`HEXTA_IMAGE_TAG` must name a `main` commit whose **Docker CI/CD** run succeeded. Check it under the repository's Actions tab before deploying.

Then edit `.env` for the optional settings: `GOOGLE_CLIENT_*`, `AI_PROVIDER` plus its key, `META_APP_ID`. Use hex values for passwords because they are embedded in database URLs.

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

```sh
cd /opt/hexta
./scripts/prod.sh pull
./scripts/prod.sh up -d --wait
./scripts/prod.sh ps -a
```

Expected:
- `migrate-api`, `migrate-ai`, `migrate-connector` and `kafka-topics` show `exited (0)`.
- `postgres`, `redis`, `redpanda`, `api`, `ai`, `web` are `healthy`.
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

If `pull` reports `denied` for an image, that GHCR package is private. Either make it public (GitHub → the package → Package settings → Change visibility), or on the server run `docker login ghcr.io -u <github user>` with a personal access token that has only `read:packages`.

## 7. Connect channels

- **Telegram:** Settings → Channels → paste a @BotFather token. The connector registers `https://hexta.org/webhooks/telegram/<connection id>` automatically. Message the bot from a private chat and check the Inbox.
- **Facebook:**
  1. In the Meta app, set the webhook callback to `https://hexta.org/webhooks/facebook`, using the value in `secrets/meta_verify_token` as the verify token.
  2. Set the valid OAuth redirect to `https://hexta.org/channels/facebook/callback`.
  3. Put the real app secret in `secrets/meta_app_secret` and `META_APP_ID` in `.env`.
  4. Run `./scripts/prod.sh up -d --wait connector api`.
- **Google login:** set the authorized redirect URI to `https://hexta.org/api/v1/auth/google/callback`.

## 8. Update to a new version

```sh
cd /opt/hexta
./scripts/prod.sh exec -T postgres pg_dumpall -U postgres | gzip > ~/hexta-pre-update-$(date +%F-%H%M).sql.gz
grep '^HEXTA_IMAGE_TAG=' infrastructure/prod/.env      # note the running version for rollback
git fetch origin && git checkout <main commit>          # configuration for that version
sed -i "s|^HEXTA_IMAGE_TAG=.*|HEXTA_IMAGE_TAG=sha-$(git rev-parse --short=7 HEAD)|" infrastructure/prod/.env
./scripts/prod.sh pull
./scripts/prod.sh up -d --wait
./scripts/prod.sh ps -a
```

Migration jobs rerun on every `up`; they apply only new migrations and exit 0 when there are none. Keep the checkout and `HEXTA_IMAGE_TAG` on the same commit so configuration and images match.

## 9. Roll back

- **Code only (no new migration):** `git checkout <previous commit>`, set `HEXTA_IMAGE_TAG` back to its `sha-…` tag, then `./scripts/prod.sh pull && ./scripts/prod.sh up -d --wait`.
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

Redpanda holds only in-flight events. The connector and API keep their durable outboxes in Postgres, so a lost Redpanda volume is recreated by `kafka-topics` and the outboxes republish pending events.

## 11. Operations

| Task | Command |
| --- | --- |
| Status | `./scripts/prod.sh ps -a` |
| Logs | `./scripts/prod.sh logs -f --tail=200 api` (or `connector`, `ingress`, `ai`, `web`, `caddy`) |
| Restart one service | `./scripts/prod.sh restart connector` |
| Resource use | `docker stats --no-stream` |
| Disk cleanup (old images) | `docker image prune -f` |
| Postgres shell | `./scripts/prod.sh exec postgres psql -U postgres -d hexta` |

Never run `down -v` on the server: it deletes the database volume. Never print `.env`, `secrets/` or `docker compose config` output into logs or chats; it contains resolved secrets.

## Known limits

- One host, no high availability. Disk loss means restoring from an off-server backup.
- Each service uses one login for migrations and runtime; separate migrator/runtime roles are a follow-up.
- No metrics or alerting are wired; use `ps`, `logs` and `docker stats`.
- Verified locally on 2026-11-16 by running this stack with `DOMAIN=localhost` and images built through `compose.prod.build.yml` (same Dockerfiles and build args as CI), covering routing, health, migrations, topic creation, and register/login on one origin. NOT_RUN: the first CI publish of the `connector` and `ai` images, and on a real server public certificates, DNS, live Facebook/Telegram callbacks and backup restore.
