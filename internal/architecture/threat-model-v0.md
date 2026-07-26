# Threat Model v0 — Obiara Platform

Status: Initial (v0), pre-P0
Prepared: 2026-07-26
Owner: backend/security workstream
Sources: `agent_plan.md` §4, §7, §8, §11, §15, §23; Obiara Doc 09 (Trust,
Safety and Compliance); Obiara Doc 08 (Data Model and AI Specification)

This is the standing threat model for the Go modular monolith (`api` +
`worker`), MongoDB operational store, member/admin web clients, Expo mobile
client, and all managed provider boundaries. It is a living document: any
change that adds a data purpose, a trust boundary, or a privileged/irreversible
action requires a threat-model delta recorded here (see `agent_plan.md` §14
review gates).

## 1. Security priorities

When controls conflict, the Safety Constitution order applies (Doc 09 §1):

1. Member physical safety
2. Member dignity and privacy
3. Platform integrity
4. Growth

Product-law invariants that are also security invariants (`agent_plan.md` §4):

- Age below 18 is a hard block with required data purge.
- Seeds are never sold; no member-to-member money path exists.
- Conversations strictly alternate; clients render state and never
  authoritatively decide it.
- Consent is purpose-specific, versioned, and enforced at data-access
  boundaries.
- No analytics event contains raw conversation content, voice, or free text.
- All privileged and irreversible actions are auditable.

## 2. Assets and security objectives

| Asset | Objective |
|---|---|
| Verified identity artifacts (Ghana Card data, verification proofs) | Confidentiality + integrity; isolated storage, envelope encryption, minimal retained proof set |
| Biometric/liveness media (face + voice) | Confidentiality; 30-day raw-media retention; human review on uncertainty |
| Voice of Introduction and in-room voice/content | Confidentiality; owner-controlled export; crypto-erasure on retention expiry |
| Account/session credentials (OTP, tokens, refresh sessions) | Confidentiality + integrity; takeover resistance |
| State-machine truth (allowance, alternation, doorway, room, pause, water, Gate) | Integrity; server-authoritative, race-safe, idempotent |
| Ledger and suban events | Integrity; append-only, immutable, recomputable |
| Consent records | Integrity + availability; the enforcement switchboard (Doc 08 §8) |
| T&S evidence and safety cases | Confidentiality + integrity; least-exposure, audit, legal hold |
| Admin plane | Confidentiality + integrity; least privilege, MFA, four-eyes |
| Provider secrets and signing keys | Confidentiality; HSM/secrets manager, rotation |
| Audit trail | Integrity + non-repudiation; append-only |

## 3. Trust boundaries

1. Member clients (mobile/web) ↔ `api` — untrusted input boundary; TLS;
   short-lived access tokens.
2. Admin web ↔ `api` admin surface — separate auth plane (SSO/MFA before
   production), RBAC/ABAC, deny-by-default.
3. `api`/`worker` ↔ MongoDB Atlas — authenticated, network-restricted,
   narrower credentials for isolated identity/biometric collections.
4. `api`/`worker` ↔ providers (SMS/WhatsApp, Resend, identity/liveness,
   storage/CDN, LiveKit, MoMo, push, AI gateway) — each an independent
   compromise scenario; provider SDK types never cross into domain models.
5. Provider webhooks → `api` ingress — unauthenticated network input until
   signature/timestamp verification.
6. Object storage/CDN ↔ clients — signed-access policy only; no public media.
7. AI gateway ↔ model vendors — no direct model calls from clients; prompts
   and outputs are data-egress paths.

## 4. Threat actors

- Romance-scam syndicates (documented sakawa/yahoo playbooks), including
  coordinated vouch rings and device farms.
- Opportunistic fraudsters and harassers; married-status deceivers.
- Account takeover attackers (SIM-swap, OTP phishing, credential stuffing).
- Malicious insiders (admin evidence access, support tooling abuse).
- Compromised or negligent providers (data egress, webhook forgery source).
- Curious/abusive members probing API and state-machine boundaries.

## 5. Threat register

Format: threat — abuse case — primary mitigations (epic that must deliver).

### 5.1 Account takeover

- OTP interception/SIM-swap — attacker takes over a verified account to farm
  trust — phone OTP with rate limits, short-lived access tokens, rotated
  refresh sessions, device binding and `device_risk` signals, session
  revocation, step-up verification on sensitive actions (E03-S01; plan §11).
- Session theft on shared devices — shared-device collisions expose a member —
  secure credential storage on mobile, cache controls preventing member-state
  leaks on web, shared-device/collision review cases (E03-S12).
- OTP enumeration/bruteforce — per-actor/device/IP/endpoint-class rate limits
  (E02, E03-S01).

### 5.2 Identity fraud

- Fake or borrowed Ghana Card — underage/deceptive personas enter romantic
  surfaces — provider verification + fallback human queue, age gate with
  hard block and 24-hour purge proof (E03-S03, E03-S05; FR-104).
- Liveness spoofing/deepfake at enrollment or step-up — face+voice anti-spoof
  ensemble, deepfake screen, uncertainty routes to human review, never a
  silent pass (E11-S08; Doc 08 §7).
- Alternate-account re-entry after Tier-A ban — device/biometric blocklist
  with propagated controls across alternate accounts (E12-S04; Doc 09 §2).
- One verified identity, multiple active accounts — strict unique index on
  active account per verified identity (plan §7.4).

### 5.3 Media and content

- Photo vault/veil bypass — direct object-URL guessing or leaked signed URLs —
  media in object storage with metadata/checksum/signed-access policy in
  MongoDB, short-lived signed URLs, watermarking on sensitive surfaces
  (E02-S10, E07-S10).
- Screenshot/exfiltration of room voice or Gate packs — watermarking, reviewer
  links with OTP and expiry (E08-S05), retention-limited exposure windows.
- Malicious upload (oversize, polyglot, malware) — media screening pipeline,
  type/size validation, quarantine-before-publish (E06-S04, E02-S10).

### 5.4 State engine integrity

- Race to create duplicate rooms on mutual water — unique indexes plus
  single-room-creation-under-races design and tests (E06-S08).
- Replay/duplicate sends breaking strict alternation — idempotency keys on
  every write, optimistic concurrency/version fields, event-sourced room
  projection, property/fuzz tests (E07-S02, E07-S03; plan §7.4).
- Allowance counter tampering or double-spend of seeds — immutable weekly
  allowance ledger (debit/credit rows, never a mutable counter), ledger
  balance property tests (E06-S01, E06-S12).
- Time-based transition manipulation — server time only, explicit IANA time
  zones, worker-driven expiry transitions (plan §7.4).
- Direct-API bypass of product laws (sow without listening, consecutive
  sends, under-tier romantic access) — deny-by-default authorization tested
  at route and use-case boundaries; server-verified 20-second listening
  eligibility (E06-S03; plan §11).

### 5.5 Admin plane

- Privilege escalation or token theft — SSO/MFA before production, RBAC/ABAC
  least privilege, session hardening, break-glass with mandatory audit
  (E16-S01; plan §9.2, §15).
- Four-eyes bypass on irreversible actions — four-eyes controls for pricing,
  policy and evidence actions, adversarial tests (E16-S08).
- Evidence viewer over-exposure — privacy-redacted least-exposure viewer,
  auto-redaction of third parties, purpose-scoped just-in-time access,
  immutable access audit (E12-S03; Doc 09 §3).

### 5.6 Webhooks and async processing

- Forged provider webhooks (payment/verification/notification) — signature
  verification, timestamp/replay checks, persisted and deduplicated via
  inbox records, dead-lettered on failure (plan §11; E14-S04).
- Duplicate/out-of-order delivery causing double effects — at-least-once-safe
  consumers with inbox/dedup records; chaos tests for duplicate and
  out-of-order delivery (plan §7.4, §13).
- Outbox loss splitting domain change from its event — durable outbox
  committed with the domain change in one transaction (plan §7.4).

### 5.7 Provider compromise

- Identity/liveness provider breach exposing member PII — minimal retained
  proof set, envelope encryption, isolated collections with narrower
  credentials, 30-day liveness raw-media retention (plan §8; Doc 09 §7).
- AI vendor egress of member content — AI gateway with vendor/model register,
  purpose-scoped data policy, counsel content excluded from matching, no raw
  content in analytics (E11-S01, E11-S07).
- Communications provider abuse (SMS/WhatsApp/Resend) — template governance,
  delivery observability, provider fallback, 30-day replaceability scoring
  (plan §11; E13-S08).
- Payment aggregator fraud — operate via licensed PSP with no own float,
  immutable double-entry ledger, daily reconciliation, suspicious-transaction
  referral runbook (E14; Doc 09 §7).

### 5.8 Insider access

- Employee access to identity/biometric/voice assets — application-level
  field encryption for the most sensitive PII, envelope/per-user key design,
  HSM-held keys, break-glass audited access (plan §8, §15).
- Support/tooling snooping on members — audited access, access reviews,
  least-privilege role matrix (E16-S01).
- Moderation workforce harm — exposure limits, counseling, rotation as
  contractual terms for any BPO/annotation partner (E12-S10; Doc 09 §7).

### 5.9 Platform-specific abuse (Sentinel-facing)

- Scam-arc progression (accelerated affection, emergency narrative,
  off-platform pull, ask patterns) — consent-mapped sequence signals, action
  ladder silent-watch → education → friction → case (E11-S11; Doc 08 §7).
- Contact/payment exfiltration in sows and rooms — pre-delivery multilingual
  screening incl. Twi/Pidgin code-switch; Sika Shield with ≥95% precision
  gate before auto-friction (E11-S09, E11-S10).
- Syndicate/vouch-ring/device anomalies — graph clustering, vouch-ring and
  mass-sow detection, legally reviewed honeypot doorways (E11-S12).

## 6. Critical unresolved decision: E2E encryption vs consented safety processing

Room content confidentiality and Sentinel scam-arc processing are in tension
(plan §23, Doc 09 §7 compliance pack). Current recorded position: encrypted
room content, scam-arc processing on consent-mapped signals only, keys
HSM-held, access break-glass audited — with a dedicated architecture/legal
threat-model decision required **before room implementation (E07)** and an
architecture review pre-P1. Until that decision is recorded, E07 stories must
not assume plaintext server access to room content.

## 7. Residency posture

Render has no African region (S0-011 finding). Production hosting remains
founder/privacy/legal-blocked until the residency interpretation question
(plan §24.6) is answered. All threats involving data location assume the
final residency decision; this model must be revisited when it lands.

## 8. Maintenance and update triggers

Update this document when any of the following occurs:

- New consent-map row, data purpose, or provider boundary is added.
- E2E/safety-processing decision or residency decision is recorded.
- A new privileged, irreversible, or payment-path action is introduced.
- An incident or red-team finding reveals an unlisted threat.
- Penetration-test or DAST results land (pre-launch gates, plan §13).
