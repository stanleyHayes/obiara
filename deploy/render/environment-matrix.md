# Render environment and release matrix

- Owner: Platform engineering
- Blueprint: [`../../render.yaml`](../../render.yaml)
- Last reviewed: 2026-07-27
- Production status: **blocked**

This matrix implements ADR-0006 without implying that Render is approved for
production member workloads. The repository Blueprint contains only the
protected, synthetic-data `staging` environment. Production resources are
deliberately absent until the residency, DPIA, provider-contract, latency,
backup, recovery and cost gates in
[`provider-residency-feasibility.md`](provider-residency-feasibility.md) are
closed.

## Environment boundaries

| Boundary       | Local development                             | Render staging                                                                                | Production                                                          |
| -------------- | --------------------------------------------- | --------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| Purpose        | Developer feedback and deterministic tests    | Integration, UAT and operations rehearsal using synthetic personas                            | Member traffic only after formal production-readiness approval      |
| Data           | Local synthetic fixtures only                 | Synthetic fixtures only; never copied from production                                         | Not provisioned                                                     |
| Compute        | Developer machine and Testcontainers          | Frankfurt API, worker, web and admin services                                                 | Blocked; topology must follow an approved replacement/follow-up ADR |
| MongoDB        | Local/Testcontainers database                 | Managed external staging cluster; API and worker receive separate least-privilege credentials | Blocked pending Atlas region, backup, support and deletion evidence |
| Deploy trigger | Manual                                        | `checksPass`; protected Render environment                                                    | Off/not defined                                                     |
| Preview        | Local or manually requested, seven-day expiry | Manual only                                                                                   | Prohibited                                                          |
| Secrets        | Untracked local environment                   | Render Dashboard/approved secret store via `sync: false`                                      | Not created                                                         |
| Durable disks  | None                                          | None; services are stateless                                                                  | Must not use ephemeral service disks as durable truth               |
| Observability  | Optional local collector                      | Credential-free HTTPS OTLP endpoint injected at runtime                                       | Blocked pending retention and residency approval                    |

## Service matrix

| Service                 | Render type                  | Health/readiness                                                            | Scale floor | Durable access                                           |
| ----------------------- | ---------------------------- | --------------------------------------------------------------------------- | ----------- | -------------------------------------------------------- |
| `obiara-api-staging`    | Web, native Go               | Render checks `/live`; `/ready` remains dependency-aware for operators      | 1           | Separate API MongoDB credential through `MONGODB_URI`    |
| `obiara-worker-staging` | Background worker, native Go | Process supervision; job failures go through scheduler/dead-letter behavior | 1           | Separate worker MongoDB credential through `MONGODB_URI` |
| `obiara-web-staging`    | Web, Node/Next.js            | `/`                                                                         | 1           | None                                                     |
| `obiara-admin-staging`  | Web, Node/Next.js            | `/`                                                                         | 1           | None                                                     |

API and worker intentionally use the same logical staging database so the
worker can process API outbox records, but the URI is injected independently
for least-privilege database users. Neither service has a persistent disk.

## Variable ownership

| Variable                           | Classification             | API | Worker | Web | Admin | Provisioning rule                                      |
| ---------------------------------- | -------------------------- | --: | -----: | --: | ----: | ------------------------------------------------------ |
| `APP_ENV`                          | Non-secret                 | yes |    yes |  no |    no | Staging environment group                              |
| `FEATURE_SOW_ENABLED`              | Runtime release control    | yes |     no |  no |    no | Explicit `true` for the composed staging capability    |
| `FEATURE_FIRES_ENABLED`            | Runtime release control    | yes |     no |  no |    no | Explicit `true` for the composed staging capability    |
| `FEATURE_PAYMENTS_ENABLED`         | Runtime release control    | yes |     no |  no |    no | Explicit `true` for the composed staging capability    |
| `FEATURE_GATE_ENABLED`             | Runtime release control    | yes |     no |  no |    no | Explicit `true` for the composed staging capability    |
| `FEATURE_AI_ENABLED`               | Runtime release control    | yes |     no |  no |    no | Explicit `false` until an AI runtime is composed       |
| `MONGODB_DATABASE`                 | Non-secret                 | yes |    yes |  no |    no | Blueprint value `obiara_staging`                       |
| `MONGODB_URI`                      | Secret                     | yes |    yes |  no |    no | Independent `sync: false` values; never shared in Git  |
| `MONGODB_URI_ROTATED_AT`           | Rotation metadata          | yes |    yes |  no |    no | Independent RFC3339 activation evidence; `sync: false` |
| `RESEND_WEBHOOK_SECRET`            | Secret                     | yes |     no |  no |    no | `sync: false`; staging-only Resend webhook             |
| `RESEND_WEBHOOK_SECRET_ROTATED_AT` | Rotation metadata          | yes |     no |  no |    no | RFC3339 activation evidence; `sync: false`             |
| `LIVENESS_HMAC_SECRET`             | Secret                     | yes |     no |  no |    no | `sync: false`; keys retained liveness proof references |
| `LIVENESS_HMAC_SECRET_ROTATED_AT`  | Rotation metadata          | yes |     no |  no |    no | RFC3339 activation evidence; `sync: false`             |
| `COMMERCE_HMAC_SECRET`             | Secret                     | yes |     no |  no |    no | `sync: false`; derives scoped commerce member keys     |
| `COMMERCE_HMAC_SECRET_ROTATED_AT`  | Rotation metadata          | yes |     no |  no |    no | RFC3339 activation evidence; `sync: false`             |
| `ADMIN_HMAC_SECRET`                | Secret                     | yes |     no |  no |    no | `sync: false`; derives opaque admin subject references |
| `ADMIN_HMAC_SECRET_ROTATED_AT`     | Rotation metadata          | yes |     no |  no |    no | RFC3339 activation evidence; `sync: false`             |
| `NNOBOA_INVITE_SECRET`             | Secret                     | yes |     no |  no |    no | signs one-time kin invitation authority                |
| `NNOBOA_INVITE_SECRET_ROTATED_AT`  | Rotation metadata          | yes |     no |  no |    no | RFC3339 activation evidence; `sync: false`             |
| `SEED_HMAC_SECRET`                 | Secret                     | yes |     no |  no |    no | keys private seed and garden projections               |
| `SEED_HMAC_SECRET_ROTATED_AT`      | Rotation metadata          | yes |     no |  no |    no | RFC3339 activation evidence; `sync: false`             |
| `CIRCLE_HMAC_SECRET`               | Secret                     | yes |     no |  no |    no | keys circle-room actors and invitation authority       |
| `CIRCLE_HMAC_SECRET_ROTATED_AT`    | Rotation metadata          | yes |     no |  no |    no | RFC3339 activation evidence; `sync: false`             |
| `OTEL_EXPORTER_OTLP_ENDPOINT`      | Configuration              | yes |    yes |  no |    no | `sync: false`; absolute credential-free HTTPS URL      |
| `SERVICE_VERSION`                  | Build identity             | yes |    yes |  no |    no | Derived at start from immutable `RENDER_GIT_COMMIT`    |
| `NEXT_PUBLIC_API_BASE_URL`         | Public build configuration |  no |     no | yes |   yes | `sync: false`; staging API URL, no credentials         |

No secret may be referenced by a Docker build argument or committed value.
The native Go and Node runtimes build directly from the pinned repository
toolchains and lockfiles.

## Promotion and rollback

1. CI and security checks must pass before Render can deploy staging.
2. Staging receives only the exact Git commit that passed checks.
3. `/live` must pass and `/ready` must confirm MongoDB before UAT begins.
4. A failed staging deploy rolls back to the last known-good Render deploy;
   database changes must remain backward compatible.
5. Production is not a promotion target in this Blueprint. Creating it requires
   the signed residency/DPIA decision, provider evidence, recovery rehearsal,
   production topology ADR and explicit founder release approval.

The current Blueprint syntax follows Render's
[Blueprint specification](https://render.com/docs/blueprint-spec),
[monorepo guidance](https://render.com/docs/monorepo-support), and
[secret configuration guidance](https://render.com/docs/configure-environment-variables).
