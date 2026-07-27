# Cloth harvest and weaver handoff

Status: implementation contract
Owner: `codex-root`
Story: E08-S08 / S6-010

## Purpose

Harvest turns a pair-owned, deterministic Cloth recipe into one bounded
production request. It is not a marketplace, public gallery, payment surface
or channel for private reflections. The weaver receives only the information
needed to produce the approved artifact.

## Authority and consent

- A harvest belongs to exactly two pair members represented by keyed opaque
  identifiers.
- Both current members independently approve the same recipe version, format,
  safe production tokens and opaque delivery reference.
- A changed recipe, format, delivery reference or policy version invalidates
  prior approvals.
- The application revalidates pair membership, Cloth ownership, recipe
  integrity and both approvals immediately before handoff.
- Withdrawal before acceptance cancels the request. Acceptance creates an
  immutable production snapshot; later cancellation is a separate audited
  transition, never a rewrite.

## State machine

`draft -> awaiting_pair -> ready -> handed_off -> accepted -> completed`

Terminal alternatives:

- `cancelled` before weaver acceptance;
- `declined` when the provider cannot satisfy the bounded specification;
- `expired` when a ready request is not handed off within seven days.

Every command carries a unique opaque command identifier, expected revision
and fingerprint. Replays return the original result. A mismatched replay or
stale revision fails without appending an event.

## Production envelope

The handoff envelope contains only:

| Field              | Contract                                                  |
| ------------------ | --------------------------------------------------------- |
| `handoffId`        | Random opaque identifier                                  |
| `recipeKey`        | Keyed opaque immutable recipe reference                   |
| `recipeVersion`    | Supported canonical grammar version                       |
| `renderSeed`       | Deterministic 64-character lowercase hex seed             |
| `productionTokens` | Six allowlisted bounded tokens produced by the grammar    |
| `format`           | One of `woven_band`, `framed_cloth`, `digital_archive`    |
| `deliveryRef`      | Envelope-encrypted opaque fulfillment reference           |
| `policyVersion`    | Version that governed member approval                     |
| `expiresAt`        | Server timestamp, no more than seven days after readiness |

It never contains names, handles, phone numbers, email addresses, street
addresses, payment details, voice, transcripts, prompt responses, reviewer
material, relationship labels or circle membership.

## Weaver boundary

- The provider adapter receives one envelope by `handoffId`; it cannot browse
  members, pairs, recipes or other orders.
- Provider callbacks are authenticated, replay-protected and limited to
  accepted, declined and completed states.
- Callback notes are bounded reason codes, not free text.
- The provider loses access after completion, cancellation, decline or expiry.
- Delivery resolution occurs inside Obiara's fulfillment boundary. The
  production provider never receives the raw delivery destination.

## Persistence and retention

- MongoDB stores HMAC-keyed pair/member/recipe references and an append-only
  event/command audit with optimistic revision checks.
- The globally unique handoff and command indexes prevent duplicate orders.
- Production envelopes expire from the provider-access collection after
  completion plus 30 days. The pair retains the Cloth artifact and a
  pseudonymous receipt under the Cloth lifecycle policy.
- Legal hold blocks physical deletion but does not restore provider access.

## Required proofs

1. Same canonical recipe and approvals create the same production payload,
   while a command retry creates no second handoff.
2. Two concurrent final approvals converge on one ready request.
3. A changed recipe, token, format or delivery reference cannot reuse consent.
4. An outsider and a single member cannot hand off, cancel or read the order.
5. Expired, cancelled and completed envelopes cannot be fetched by a provider.
6. Raw MongoDB and provider payload scans contain none of the prohibited
   identity, reflection, payment or address fields.
