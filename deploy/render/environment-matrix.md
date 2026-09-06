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
| `MONGODB_DATABASE`                         |           yes |      yes | Non-secret Blueprint value. **Name the environment in the database**: `obiara_dev` locally, `obiara_prod` in production. A connection string pointed at the wrong cluster then fails on a missing database instead of quietly reading and writing the wrong one. The live database is still named `obiara_production`; see the note below before changing it |
| `MONGODB_URI` and `MONGODB_URI_ROTATED_AT` |           yes |      yes | Independent `sync: false` values; RFC3339 rotation time under 90 days |
| `RESEND_WEBHOOK_SECRET` and rotation time  |           yes |       no | `sync: false`; verified production webhook signing secret             |
| `LIVENESS_HMAC_SECRET` and rotation time   |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `COMMERCE_HMAC_SECRET` and rotation time   |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `ADMIN_HMAC_SECRET` and rotation time      |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `NNOBOA_INVITE_SECRET` and rotation time   |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `SEED_HMAC_SECRET` and rotation time       |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `CIRCLE_HMAC_SECRET` and rotation time     |           yes |       no | `sync: false`; at least 32 random bytes                               |
| `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET`    | optional pair |       no | Both or neither; `sync: false`                                        |
| `OTP_PROVIDERS=arkesel`                    |           yes |       no | Ordered OTP ladder; the simulator is rejected outside development     |
| `ARKESEL_API_KEY`, `ARKESEL_SENDER_ID`     |           yes |       no | `sync: false`; sender id must be pre-approved and at most 11 chars    |
| `EMAIL_PROVIDER=resend`                    |           yes |       no | Non-secret selection; the simulator is rejected outside development   |
| `RESEND_API_KEY`, `RESEND_FROM_ADDRESS`    |           yes |       no | `sync: false`; the from address must be a verified Resend domain      |
| `RESEND_REPLY_TO`                          |      optional |       no | `sync: false`; optional operator reply address                        |
| `WHATSAPP_PROVIDER=disabled`               |           yes |       no | `disabled` until Cloud API provisioning; never `simulator`            |
| `SAFEGUARDING_HMAC_SECRET` and rotation time |         yes |       no | `sync: false`; at least 32 random bytes. **The API refuses to boot without it.** |
| `OBJECT_STORAGE_*` (region, bucket, keys)  | optional set |      no | All four or none. Without them nobody can record a voice, so the whole introduction, sow and pod chain is absent |
| `PAYSTACK_SECRET_KEY`                      | optional      |      no | `sync: false`. Absent means nothing can be bought and the purchase route is not registered. `sk_live_` moves real money; `sk_test_` logs a warning at boot |
| `PAYSTACK_CALLBACK_URL`                    | optional      |      no | Where Paystack reports outcomes. Must reach `/v1/payments/paystack/webhook` from outside |
| `PAYSTACK_BASE_URL`                        | optional      |      no | Defaults to `https://api.paystack.co`. Set only to point at a sandbox |
| `AFFILIATE_COMMISSION_PESEWAS`             | optional set  |      no | Flat amount one qualified conversion earns. All three affiliate values or none |
| `AFFILIATE_MINIMUM_PAYOUT_PESEWAS`         | optional set  |      no | Floor a balance must clear before it can be requested |
| `AFFILIATE_WITHHOLDING_BASIS_POINTS`       | optional set  |      no | Tax withheld on commission, hundredths of a percent (750 = 7.5%). **No default: an unset rate leaves the whole scheme absent rather than paying gross.** Set it with your accountant |
| `OTEL_EXPORTER_OTLP_ENDPOINT`              |      optional | optional | Credential-free HTTPS URL; `sync: false`                              |
| `SERVICE_VERSION`                          |           yes |      yes | Derived from immutable `RENDER_GIT_COMMIT` at start                   |

### Renaming the production database

The convention above is `obiara_prod`. The live database is currently
`obiara_production`, and the two are not the same string.

MongoDB has no rename. Switching means copying the data to the new name,
pointing `MONGODB_DATABASE` at it, and only then dropping the old one — with a
verified backup first. **Changing the variable alone would point production at
an empty database**, which is why nothing here does it automatically. Until
somebody does that migration deliberately, production keeps its current name
and the convention applies to everything new.

### What is absent rather than broken

Four groups above are optional as a set, and each is composed only when it is
complete. This is deliberate: a half-configured payment rail should read as
"off", not as "on and broken", so a missing key removes the surface rather
than leaving one that always fails.

- **No object storage**: no member can record a voice, so introductions, sows
  and pods are all absent.
- **No `PAYSTACK_SECRET_KEY`**: nothing can be bought. `POST
  /v1/membership/purchases` and the webhook are not registered at all.
- **No affiliate values**: the referral scheme is not composed and nothing
  accrues. A scheme that accrued but could never legally pay out would be a
  liability that only grows.
- Sponsored seats need none of the above beyond an operator to record a
  deposit, so organization funding works as soon as the API runs.

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

## Outbound delivery providers

Delivery adapters are selected explicitly and validated at startup. The API
**refuses to boot** when a channel is left on the simulator outside
development, because the simulator reports every message as delivered while
transmitting nothing. That combination previously shipped to production and
was the cause of members completing sign-up without ever receiving an OTP.

| Channel  | Production value | Effect                                                                      |
| -------- | ---------------- | --------------------------------------------------------------------------- |
| OTP      | `arkesel`        | SMS over Arkesel, the primary Ghanaian transport                            |
| Email    | `resend`         | Admin MFA codes and operational mail over Resend                            |
| WhatsApp | `disabled`       | Every send fails loudly and is delivery-logged; nothing is silently dropped |

`OTP_PROVIDERS` accepts a comma-separated ladder tried in order, for example
`arkesel,twilio`. Each rung is attempted only if the previous one fails, and
every attempt is reported through the redacting logger without the code or the
recipient. To add WhatsApp as an OTP fallback later, provision the Cloud API,
set `WHATSAPP_PROVIDER=meta` with `META_WHATSAPP_PHONE_NUMBER_ID` and
`META_WHATSAPP_ACCESS_TOKEN`, then set `OTP_PROVIDERS=arkesel,whatsapp`.

## Bootstrapping the first admin

Admin enrollment requires an existing stepped-up admin session, so an empty
database has no way to create its first operator. Run the one-off bootstrap
job against the production database instead:

```
MONGODB_URI=...            \
MONGODB_DATABASE=obiara_production \
BOOTSTRAP_ADMIN_EMAIL=...  \
BOOTSTRAP_ADMIN_PASSWORD=... \
go run ./services/api/cmd/bootstrap
```

The command is idempotent: re-running it resets the named operator's roles,
status and password, which also makes it the recovery path for a locked-out
super admin. Credentials are read only from the environment, never from
flags, so they do not appear in process listings or shell history. Every run
appends an `admin.bootstrap.*` entry to the immutable `admin_access` audit.

## Verifying a deployment

Run the member journey against the deployment before opening it to traffic:

```
API_BASE=https://obiara-api-production.onrender.com \
MONGODB_URI=... MONGODB_DATABASE=obiara_production \
go run ./services/api/cmd/smoke
```

It walks health, OTP sign-up, session rotation and its theft check, device
registration, the member surface and the authentication boundary, then
removes everything it created. `MONGODB_URI` is needed for the two things the
API deliberately does not expose: reading back the code a member would
receive by SMS, and cleaning up afterwards.

A capability whose feature flag is off answers `503 feature_unavailable`.
That is the route working, and the walk reports it as such.

## When a channel goes dark

Delivery failures are reported as `503` with a code naming the channel —
`otp_delivery_failed`, `mfa_delivery_failed` — rather than a generic 500, and
the cause is logged under `fault` against the same correlation id the caller
was shown. Searching the logs for that id gives the provider status directly,
for example `status 401, provider error validation_error` for a rejected
Resend key.
