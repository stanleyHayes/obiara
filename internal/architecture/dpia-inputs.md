# DPIA Inputs v0 — Obiara

Status: Initial inputs for the pre-P0 DPIA, not the signed assessment
Prepared: 2026-07-26
Owner: backend/security workstream → DPO (to be appointed pre-P0)
Sources: `agent_plan.md` §15; Obiara Doc 08 (Data Model and AI
Specification); Obiara Doc 09 (Trust, Safety and Compliance);
`data-classification.md`; `threat-model-v0.md`

The Data Protection Act 2012 (Act 843) requires DPC registration and a DPO
before P0, and a full DPIA covering identity, biometrics, voice, matching,
safety and AI — refreshed per new AI purpose (Doc 09 §7). This document
assembles the engineering inputs the DPO needs: systematic description of
each high-risk processing operation, necessity and proportionality analysis,
risk assessment, and planned mitigations. It does not replace the DPO's
assessment or any required DPC consultation.

## 1. Processing operations under assessment

### 1.1 Identity verification (Ghana Card / provider)

- **Description:** Government-ID capture and provider verification at
  registration; fallback human review queue; ID-derived date of birth drives
  the age gate (FR-104); minimal proof set retained for account life + 90 d.
- **Necessity/proportionality:** Verified-only reach is the core safety
  promise; no less-intrusive means achieves age assurance and one-person-
  one-account. Collection is minimized to the proof set; raw artifacts are
  isolated, envelope-encrypted, under narrower DB credentials.
- **Risks to members:** breach exposure of national-ID data; exclusion from
  wrongful provider rejection; insider misuse.
- **Mitigations:** isolation + encryption (C4 handling), fallback human queue,
  access audit, break-glass controls, 30-day-replaceable provider adapters,
  breach runbook with 72-hour reporting clock.
- **Residual/open:** provider data-location scoring pending vendor selection
  (founder decision §24.5); residency decision §24.6.

### 1.2 Biometric liveness (face + voice)

- **Description:** Face+voice anti-spoof ensemble at enrollment and step-ups;
  deepfake screen; uncertainty routes to a human queue — never a silent pass
  (Doc 08 §7). Raw media retained 30 days.
- **Necessity/proportionality:** Defends against fake-identity and
  alternate-account re-entry after Tier-A bans; biometric blocklist is the
  only durable re-entry control. Retention is short and purpose-bound.
- **Risks:** biometric data is irrevocable if breached; false rejects
  disproportionately affecting some groups; function creep.
- **Mitigations:** 30-day raw-media retention with crypto-erasure, human
  review on uncertainty, fairness monitoring of reject rates, purpose
  limitation enforced by consent-map row (identity & safety — required,
  cannot disable, disclosed).
- **Residual/open:** fairness baseline must be measured per demographic proxy
  once a vendor PoC exists; blocklist matching thresholds need documented
  review.

### 1.3 Voice content (Voice of Introduction, pods, room voice)

- **Description:** Member-recorded voice is the primary introduction and
  conversation medium; room voice retained 180 days post-closure, owner
  export anytime, then crypto-erased.
- **Necessity/proportionality:** "Voice over text" is a product law; voice is
  also a character signal text cannot provide. Listening eligibility and
  strict alternation bound exposure; nothing is public or feed-ranked.
- **Risks:** intimate-content exposure via breach or insider access;
  non-consensual capture/sharing by counterparts; voice used for purposes
  beyond the product loop.
- **Mitigations:** signed-access media policy, watermarking on sensitive
  surfaces, C4 encryption, retention automation with proof of deletion,
  transcription for T&S review only under case scope with original retained
  per legal hold, no voice in analytics events (producer-enforced taxonomy).
- **Residual/open:** E2E encryption design decision (threat model §6) may
  change server-side access assumptions; counterpart-capture risk is only
  partially technical (deterrence via watermarking and conduct policy).

### 1.4 Matching intelligence

- **Description:** Cold start is rules + trust-path graph only (first ~5k
  members). Model-assisted ranking (two-tower + gradient-boosted ranker) is
  gated behind consent, data and fairness readiness (P1+). Inputs are
  consent-gated per source; hard-constraint filters are compiled, not
  learned; skin-tone and tribe are never features.
- **Necessity/proportionality:** Introduction-over-discovery requires some
  ranking; the opt-in consent-map row (full off = rules-only intros) keeps
  personalization genuinely optional.
- **Risks:** exposure unfairness across ethnicity proxies; colorism drift;
  manipulation of intimate life choices; inferred sensitive attributes.
- **Mitigations:** quarterly fairness/exposure audits with correction,
  grounded explainability (template-grounded true reasons, no generated
  flattery), negative penalty on predicted-decline exposure, counsel content
  excluded from matching features, model cards and versioned vendor register.
- **Residual/open:** fairness gates and offline evaluation criteria must be
  defined before any model ships (E11-S05); DPIA refresh required at that
  point regardless of timing.

### 1.5 Safety monitoring (Sentinel: screening, scam-arc, syndicate detection)

- **Description:** Pre-delivery multilingual screening of sows/pods
  (contact-exfiltration, payment/gift language, sexual content, threats);
  consent-mapped scam-arc sequence monitoring in rooms; graph syndicate and
  vouch-ring detection; Sika Shield payment-language interception.
- **Necessity/proportionality:** Romance-scam harm is the dominant member
  safety risk in-market; the action ladder (silent watch → education →
  friction → case) applies the least-intrusive effective measure first.
  Scam-arc room monitoring is an opt-out row with per-room override; off =
  education-only mode.
- **Risks:** surveillance of intimate conversations; false positives
  stigmatizing innocent members; multilingual classifier bias
  (Twi/Pidgin code-switch).
- **Mitigations:** consent-map governance with warning-on-off, ≥95% precision
  gate before auto-friction (below it: human review), false-positive appeal
  overturn target <8%, education-first ladder, local annotation partner under
  worker-welfare terms, E2E design decision before room processing.
- **Residual/open:** the E2E-vs-monitoring architecture decision (threat
  model §6) is the single largest open privacy question; legal review of
  honeypot doorways required before operation.

### 1.6 Okyeame (AI companion)

- **Description:** Hosted LLM behind the AI gateway; capability whitelist
  (draft counsel, cultural briefings, theme facilitation, translation,
  proverbs, Gate relay wording, closure help); hard bans on initiating
  romance, impersonation, autonomous sending, therapy claims, and exposing
  other members' private data.
- **Necessity/proportionality:** Each capability maps to a member-facing
  feature; small-model routing and caching limit data sent to vendors.
- **Risks:** members forming attachment to or being manipulated by AI;
  counsel content leaking into matching; vendor egress of intimate drafts.
- **Mitigations:** AI-assisted drafts marked to their author, assist-rate
  throttle per thread, refusal/escalation on self-harm and abuse signals per
  the care protocol, counsel isolation from matching enforced at the gateway,
  per-language prompt packs with red-team suites.
- **Residual/open:** vendor data-policy terms must be verified at selection;
  Twi TTS quality gate before ship.

### 1.7 Suban character ledger

- **Description:** Append-only event ledger (meeting follow-through, kind
  closures, vouch outcomes, conduct findings) producing thresholded marks —
  never a number to members; positives decay (12-month half-life), negatives
  per rehabilitation table.
- **Necessity/proportionality:** "Reputation is capital" requires a durable
  character record; thresholded marks and decay bound the impact; members
  can view every event behind their marks and appeal.
- **Risks:** a de facto social-credit score; permanent stigma; gaming or
  collusion distorting marks.
- **Mitigations:** no numeric display, rehabilitation decay, recomputable
  marks (no cached authority), anti-collusion graph checks, Mpanyimfo
  consultation + published changelog for any formula change, member
  explanation/appeal view.
- **Residual/open:** appeal SLA and panel workload modeling needed before P1
  productization.

### 1.8 Analytics

- **Description:** Producer-enforced event taxonomy (`zone.object_action`);
  no event carries conversation content, voice or free text; pseudonymous
  identifiers; opt-out via a single toggle; pseudonymized at 90 days,
  aggregated at 13 months.
- **Risks:** re-identification from pseudonymous behavioral data; scope creep
  in event properties.
- **Mitigations:** schema registry blocks unmapped/free-text properties in CI,
  one-toggle opt-out, retention automation.

## 2. Cross-cutting risks and consultations

| Item | Status | Owner / trigger |
|---|---|---|
| Member-content residency (Render has no African region) | Open — production blocked | Founder + DPO/legal; plan §24.6 |
| E2E encryption vs safety processing | Open — must close before E07 | Architecture + legal; threat model §6 |
| DPC registration and DPO appointment | Required pre-P0 | Founder/ops; Doc 09 §7 |
| CERT-GH liaison and 72-hour incident runbook | Required P0 | Security workstream |
| Ghana DPC consultation on biometric processing | Assess with DPO once vendor PoC exists | DPO |
| Diaspora GDPR/SCC posture | P2 scope; single-standard commitment recorded | DPO/legal |
| DPIA refresh per new AI purpose | Standing rule | DPO; Doc 09 §7 |

## 3. Rights handling commitments

- Export and deletion with legal-hold behavior; proof-of-deletion records
  (pre-P0, `agent_plan.md` §15).
- Under-18 hard block with 24-hour purge proof (E03-S05).
- Consent receipts and purpose-bound enforcement at data-access boundaries.
- Member visibility into suban events and matching explanations; appeal paths
  for safety and moderation actions.
- Care-protocol guarantee: no punitive action attaches to seeking help.
