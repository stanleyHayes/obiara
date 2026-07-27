# Data Classification and Processing Register v0 — Obiara

Status: Initial (v0), pre-P0
Prepared: 2026-07-26
Owner: backend/security workstream
Sources: `agent_plan.md` §8, §15; Obiara Doc 08 (Data Model and AI
Specification, esp. §8 consent map); Obiara Doc 09 (Trust, Safety and
Compliance, esp. §7 retention table)

This document is the data inventory, classification scheme and processing
register required by `agent_plan.md` §15 before P0. It feeds the DPIA
(`dpia-inputs.md`), Ghana DPC registration, and the collection-level design
artifacts (JSON schema validation, retention automation) that land with the
MongoDB conventions in E02-S05.

## 1. Classification levels

| Level                     | Definition                                                                                                                              | Handling baseline                                                                                                                                         |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| C1 Public                 | Marketing/brand content, published market packs                                                                                         | No restriction                                                                                                                                            |
| C2 Internal               | Configuration, feature flags, aggregated/pseudonymized analytics                                                                        | Authenticated staff access; no raw member content                                                                                                         |
| C3 Confidential (PII)     | Member profile, contact data, trust edges, consent records, notification preferences, order/receipt data                                | Encrypted at rest; least-privilege access; no raw values in logs or analytics events                                                                      |
| C4 Restricted (sensitive) | Identity verification artifacts, biometric/liveness media, voice assets, room/doorway content, T&S evidence, care signals, suban events | Isolated collections, envelope/application-level field encryption, narrower DB credentials, break-glass audited access, strict retention + crypto-erasure |
| C5 Secret                 | Provider credentials, signing keys, encryption keys                                                                                     | Secrets manager only; HSM-held where designated; rotation; never in code/logs/fixtures                                                                    |

Rules applying across levels (`agent_plan.md` §8, §14):

- No secret, raw production data or provider credential in code, logs or
  fixtures; structured logging with PII redaction.
- No analytics event contains raw conversation content, voice, or free text
  (Doc 08 §3 event taxonomy is producer-enforced).
- Media lives in object storage; MongoDB holds metadata, ownership,
  retention, checksum and signed-access policy.
- TTL indexes are used only where automatic deletion is legally correct;
  legal-hold paths are preserved.

## 2. Processing register (by collection family)

Lawful-basis shorthand: C = contract, LI = legitimate interest, Cons =
consent (consent-map row), LO = legal obligation.

| Collection family                                            | Class | Purpose (consent-map row)                            | Basis                                      | Notes                                                       |
| ------------------------------------------------------------ | ----- | ---------------------------------------------------- | ------------------------------------------ | ----------------------------------------------------------- |
| `accounts`, `sessions`, `roles`, `admin_access`              | C3    | Service provision; authentication                    | C                                          | Device/session revocation supported                         |
| `identity_verifications`, `identity_bindings`, `device_risk` | C4    | Identity & safety processing (required)              | C+LI                                       | Isolated storage, envelope keys, minimal retained proof set |
| `profiles`, `doorway_questions`                              | C3    | Service provision; matching inputs                   | C / Cons (matching personalization opt-in) | Per-source toggles; full off = rules-only intros            |
| `voice_assets`, `photo_assets`                               | C4    | Service provision; verification                      | C                                          | Owner export anytime; crypto-erasure on expiry              |
| `consent_records`, `privacy_requests`, `legal_holds`         | C3    | Consent enforcement; rights handling                 | LO+C                                       | Versioned, purpose-specific, revocable where permitted      |
| `circles`, `circle_memberships`, `vouches`, `trust_edges`    | C3    | Community/trust features                             | C+LI                                       | Provenance-carrying; member-controlled visibility           |
| `seed_allowance_entries`, `seeds`, `pods`, `doorway_threads` | C3    | Core product loop                                    | C                                          | Immutable allowance ledger                                  |
| `rooms`, `room_events`, `pause_stones`, `theme_progress`     | C4    | Core product loop; scam-arc monitoring (opt-out row) | C / Cons                                   | E2E design pending §6 of threat model                       |
| `cloth_artifacts`, `ceremony_events`, `gate_reviews`         | C4    | Ceremony features; Gate review                       | C+Cons                                     | Pair-owned; reviewer packs expire                           |
| `fires`, `fire_attendance`, `embers`                         | C3    | Community events                                     | C                                          | Recording only per fire consent policy                      |
| `games`, `game_moves`, `ratings`                             | C3    | Games; play-portraits (opt-in row)                   | C / Cons                                   | Skill never influences matching visibility                  |
| `introductions`, `matching_explanations`                     | C3    | Matching (personalization opt-in row)                | Cons                                       | Grounded reasons only; no generated flattery                |
| `suban_events`, `suban_marks`                                | C4    | Character ledger; safety                             | LI                                         | Append-only; recomputable; member-visible events            |
| `reports`, `safety_cases`, `case_evidence`, `panel_rulings`  | C4    | Trust & safety, care                                 | LI+LO                                      | Least-exposure viewer; reporter identity never disclosed    |
| `notification_preferences`, `notification_deliveries`        | C3    | Communications                                       | C                                          | Caps server-enforced; opt-out honored                       |
| `orders`, `ledger_entries`, `payouts`, `reconciliations`     | C3    | Commerce (P1+)                                       | C+LO                                       | Immutable double-entry ledger                               |
| `audit_events`, `outbox`, `inbox`                            | C2/C3 | Auditability; durability                             | LI+LO                                      | Append-only; privileged-action audit mandatory              |
| `analytics_events`                                           | C2    | Product analytics (opt-out row)                      | Cons                                       | Pseudonymous only; pseudonymized at 90 d                    |
| `market_packs`, `feature_flags`, `configuration_changes`     | C2    | Operations                                           | LI                                         | Configuration changes audited                               |

## 3. Binding retention table (Doc 09 §7)

| Data class          | Retention                                                        |
| ------------------- | ---------------------------------------------------------------- |
| ID artifacts        | Verification proof minimal set for account life + 90 days        |
| Liveness raw media  | 30 days                                                          |
| Room voice          | 180 days post-closure (owner export anytime; then crypto-erased) |
| Gate reviewer packs | 30 days                                                          |
| T&S evidence        | Case life + 2 years, or legal hold                               |
| Financial ledgers   | 6 years                                                          |
| Analytics events    | Pseudonymized at 90 days; aggregated at 13 months                |

Retention automation and proof-of-deletion records are a pre-P0 requirement
(`agent_plan.md` §15); under-18 hard blocks require a 24-hour purge proof
(E03-S05).

## 4. Consent map (the master switchboard — Doc 08 §8)

| Purpose                      | Default                        | Member control                                   | Data classes                             |
| ---------------------------- | ------------------------------ | ------------------------------------------------ | ---------------------------------------- |
| Identity & safety processing | Required for service           | Cannot disable                                   | ID, liveness, screening signals          |
| Matching personalization     | Opt-in at S-06                 | Per-source toggles; full off = rules-only intros | Values tags, transcript-derived features |
| Scam-arc room monitoring     | Opt-out available with warning | Per-room override; off = education-only mode     | Metadata + screened signals              |
| Play-portraits               | Opt-in                         | View/delete anytime                              | Game telemetry                           |
| Product analytics            | Opt-out                        | One toggle                                       | Pseudonymous events only                 |

Engineering rule: every feature change that touches member data names its
consent-map row; CI blocks unmapped access (`agent_plan.md` §14). The
consent-map row table above is the authoritative register until the
machine-enforced registry ships with E15-S01/E15-S02.

## 5. Residency and transfer

- Member-content residency posture is unresolved pending the founder/legal
  decision in `agent_plan.md` §24.6 (Render has no African region; S0-011).
- Diaspora (P2) requires GDPR alignment with SCCs for any transfer; Doc 09
  commits to honoring EU/UK rights globally as a single standard.
- Provider data-location scoring is part of every provider selection
  (`agent_plan.md` §11).

## 6. Maintenance

Review on: any new collection family, consent-map row, retention change,
provider boundary, residency decision, or legal-hold process change. The
DPIA (`dpia-inputs.md`) refreshes with each new AI purpose.
