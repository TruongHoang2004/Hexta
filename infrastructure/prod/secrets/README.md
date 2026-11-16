# Connector secret files

Created on the server only (never committed); the files in this directory are git-ignored. Each file holds one value
with no trailing newline requirement. See the "Secrets" step in `infrastructure/README.deploy.md`.

| File | Content |
| --- | --- |
| `connector_management_token` | Same value as `CONNECTOR_MANAGEMENT_TOKEN` in `.env` |
| `connector_credential_key` | 64 hex characters (`openssl rand -hex 32`); encrypts stored Page tokens and Bot Tokens |
| `meta_app_secret` | Meta app secret, or a random placeholder if Facebook is not used |
| `meta_verify_token` | Random token entered in the Meta webhook settings |
