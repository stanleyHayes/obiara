# Render backend production environment

- Owner: Platform engineering
- Blueprint: [`../../render.yaml`](../../render.yaml)
- Last reviewed: 2026-08-05
- Repository configuration: **ready**
- Provisioning and launch: **manual external action**

Render hosts only the stateless Go API and background worker. Marketing,
member web and admin are Vercel projects and must never be recreated from this
Blueprint. Production services are protected, network-isolated and configured
with `autoDeployTrigger: off`; creating the Blueprint cannot publish a new Git
revision automatically.

## Service matrix

| Service                    | Render type                  | Health/readiness                                               | Durable access                            |
| -------------------------- | ---------------------------- | -------------------------------------------------------------- | ----------------------------------------- |
| `obiara-api-production`    | Web, native Go               | Render checks `/live`; operators use dependency-aware `/ready` | Independent least-privilege `MONGODB_URI` |
| `obiara-worker-production` | Background worker, native Go | Process supervision and dead-letter job handling               | Independent least-privilege `MONGODB_URI` |

Both services use the same logical `obiara_production` database so the worker
can consume API outbox records, but each receives its own database credential.
Neither service uses a persistent Render disk; MongoDB remains the durable
source of truth.

## Variable ownership

| Variable                                   |           API |   Worker | Rule                                                                  |
| ------------------------------------------ | ------------: | -------: | --------------------------------------------------------------------- |
| `APP_ENV=production`                       |           yes |      yes | Shared non-secret environment group                                   |
| `MONGO_CONNECT_TIMEOUT=10s`                |           yes |      yes | Shared runtime limit                                                  |
| `SHUTDOWN_TIMEOUT=20s`                     |           yes |      yes | Shared runtime limit                                                  |
| `FEATURE_SOW_ENABLED=true`                 |           yes |       no | Explicit composed capability                                          |
| `FEATURE_FIRES_ENABLED=true`               |           yes |       no | Explicit composed capability                                          |
| `FEATURE_GATE_ENABLED=true`                |           yes |       no | Explicit composed capability                                          |
| `FEATURE_PAYMENTS_ENABLED=false`           |           yes |       no | Fail closed until live payment approval and credentials               |
| `FEATURE_AI_ENABLED=false`                 |           yes |       no | Fail closed until an approved AI runtime exists                       |
| `MONGODB_DATABASE=obiara_production`       |           yes |      yes | Non-secret Blueprint value                                            |
| `MONGODB_URI` and `MONGODB_URI_ROTATED_AT` |           yes |      yes | Independent `sync: false` values; RFC3339 rotation time under 90 days |
| `RESEND_WEBHOOK_SECRET` and rotation time  |           yes |       no | `sync: false`; verified production webhook signing secret             |
| `LIVENESS_HMAC_SECRET` and rotation time   |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `COMMERCE_HMAC_SECRET` and rotation time   |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `ADMIN_HMAC_SECRET` and rotation time      |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `NNOBOA_INVITE_SECRET` and rotation time   |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `SEED_HMAC_SECRET` and rotation time       |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `CIRCLE_HMAC_SECRET` and rotation time     |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET`    | optional pair |       no | Both or neither; `sync: false`                                        |
| `OTEL_EXPORTER_OTLP_ENDPOINT`              |      optional | optional | Credential-free HTTPS URL; `sync: false`                              |
| `SERVICE_VERSION`                          |           yes |      yes | Derived from immutable `RENDER_GIT_COMMIT` at start                   |

Use `services/api/.env.production` and `services/worker/.env.production` as
local copy/paste worksheets. They are ignored by Git. Their tracked `.example`
counterparts document every field without carrying a credential.

## Deployment sequence

1. Provision the approved MongoDB production database, least-privilege API and
   worker users, backups and restore evidence.
2. Generate independent random HMAC values and record one current RFC3339 UTC
   activation time for each secret.
3. Create/sync the Blueprint, fill every required `sync: false` value, and keep
   both services suspended from traffic.
4. Deploy the exact verified `main` SHA manually. Confirm `/live`, then `/ready`.
5. Configure each Vercel app with its backend URL and verify server-side BFF
   calls, waitlist submission, authentication and admin access.
6. Complete production legal, provider, physical-device, backup/restore and
   human go/no-go gates before enabling member traffic.

A successful Blueprint validation or build proves configuration only. It does
not prove provider contracts, residency approval, credentials, live database
readiness, store acceptance, or production launch approval.
