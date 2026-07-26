# Obiara Traceability Matrix

Status: Seeded (v1), Sprint 0
Prepared: 2026-07-26
Owner: product/engineering traceability (claimed under S0-013)
Sources: SRS (`Obiara_02_SRS.docx` §2–4), PRD module map (`agent_plan.md`
§20), screen contract (`Obiara_06_UX_Flows_Screens.docx`), consent map
(`Obiara_08_Data_Model_AI_Spec.docx` §8), sprint roadmap (`agent_plan.md` §21)

SRS §4 rule: **no requirement ships unlinked; no unlinked code ships.** This
file is the seed of that matrix. It lives in the repository (not an external
wiki) so it can be CI-checked and reviewed with the code it traces.

## 1. Linking format

Every row links:

`Req ID → PRD module → Screens (S-xx) → Epic/Story (agent_plan §20) →
Consent-map row → Test cases → Test class → Release/Sprint → Status`

Conventions:

- Requirement IDs are the SRS FR/NFR IDs; they are stable and never reused.
- Screen IDs follow Doc 06 (`S-01`…`S-90`); `—` means no member-facing screen
  (pure backend/ops requirement).
- Epic/story IDs are the stable local planning IDs in `agent_plan.md` §20
  until Jira keys are assigned (`OBI-*`).
- Consent-map rows are from Doc 08 §8: `identity-safety` (required),
  `matching-personalization` (opt-in), `scam-arc-monitoring` (opt-out),
  `play-portraits` (opt-in), `product-analytics` (opt-out), or `—` when no
  member personal data is processed.
- Test case IDs: `TC-<story>-<nn>` (unit/contract), `TI-<story>-<nn>`
  (Testcontainers integration), `TP-<story>-<nn>` (property/fuzz/chaos),
  `TE-<story>-<nn>` (end-to-end/device). A row is DONE only when its test
  cases exist and pass.
- Status: `LINKED` (mapping complete, not built), `IN PROGRESS`, `VERIFIED`
  (tests green on `main`), `RELEASED`.

## 2. P0 functional requirements

### FR-100 Identity & Account (PRD M1) — test class: integration + manual ID lab

| Req    | Requirement (abbreviated)                                               | Screens                          | Epic/Story                          | Consent row     | Test cases                   | Release        | Status |
| ------ | ----------------------------------------------------------------------- | -------------------------------- | ----------------------------------- | --------------- | ---------------------------- | -------------- | ------ |
| FR-101 | No romantic-surface access below Tier 1; no sowing below Tier 2         | S-20–S-23, S-30–S-33, S-40, S-60 | E03-S06, E06-S04                    | identity-safety | TI-E03-S06-01, TP-E03-S06-01 | P0 / Sprint 2  | LINKED |
| FR-102 | Verify national ID against issuer; exactly one active account per ID    | S-04, S-04b                      | E03-S03                             | identity-safety | TI-E03-S03-01, TC-E03-S03-02 | P0 / Sprint 2  | LINKED |
| FR-103 | Voice+face liveness; reject presented media; human queue on uncertainty | S-05                             | E03-S04, E11-S08                    | identity-safety | TI-E03-S04-01, TE-E03-S04-01 | P0 / Sprint 2  | LINKED |
| FR-104 | Age ≥18 from ID; hard block; purge biometric/ID artifacts ≤24 h         | S-04, S-05                       | E03-S05                             | identity-safety | TI-E03-S05-01, TP-E03-S05-01 | P0 / Sprint 2  | LINKED |
| FR-105 | Vouch attestations immutable, timestamped, linked to voucher suban      | S-13                             | E05-S07 (P0 assisted), E05-S08 (P1) | identity-safety | TI-E05-S08-01                | P1 (P0 manual) | LINKED |
| FR-106 | Export ≤72 h machine-readable; deletion ≤30 d with crypto-erasure       | — (settings)                     | E03-S10                             | —               | TI-E03-S10-01, TI-E03-S10-02 | P0 / Sprint 3  | LINKED |

### FR-200 Seed Economy (PRD M4) — test class: API contract + fuzz

| Req    | Requirement (abbreviated)                                                          | Screens    | Epic/Story       | Consent row     | Test cases                   | Release       | Status                 |
| ------ | ---------------------------------------------------------------------------------- | ---------- | ---------------- | --------------- | ---------------------------- | ------------- | ---------------------- |
| FR-201 | Allowance enforced server-side; no purchase path in code/config/SKU                | S-22       | E06-S01, E14-S01 | —               | TP-E06-S01-01, TC-E14-S01-01 | P0 / Sprint 5 | LINKED                 |
| FR-202 | Sow requires ≥20 s server-verified VoI playback                                    | S-21, S-22 | E06-S03          | —               | TI-E06-S03-01, TP-E06-S03-01 | P0 / Sprint 5 | LINKED (see §4 sample) |
| FR-203 | Declined-doorway re-sow blocked 90 days at API level                               | S-32       | E06-S06          | —               | TI-E06-S06-01, TP-E06-S06-01 | P0 / Sprint 6 | LINKED                 |
| FR-204 | Sow audio Sentinel-screened ≤10 s p95; rejects return seed unspent + reason code   | S-22       | E06-S04, E11-S09 | identity-safety | TI-E06-S04-01, TC-E11-S09-01 | P0 / Sprint 5 | LINKED                 |
| FR-205 | Pod playback receipts never expose partial-listen events to sowers                 | S-31       | E06-S05          | —               | TC-E06-S05-01                | P0 / Sprint 6 | LINKED                 |
| FR-206 | Three-exchange doorway cap + mutual-water room creation; races resolve to one room | S-33       | E06-S07, E06-S08 | —               | TI-E06-S08-01, TP-E06-S08-01 | P0 / Sprint 6 | LINKED                 |

### FR-300 Rooms & Alternation (PRD M5) — test class: state-machine + chaos

| Req    | Requirement (abbreviated)                                                                     | Screens    | Epic/Story       | Consent row     | Test cases                   | Release       | Status |
| ------ | --------------------------------------------------------------------------------------------- | ---------- | ---------------- | --------------- | ---------------------------- | ------------- | ------ |
| FR-301 | Consecutive sends impossible at API layer incl. offline queue, retries, multi-device          | S-40       | E07-S02, E07-S03 | —               | TI-E07-S02-01, TP-E07-S02-01 | P0 / Sprint 7 | LINKED |
| FR-302 | Response windows, pause stones, re-light, closure per room state machine; transitions audited | S-41, S-46 | E07-S04, E07-S05 | —               | TI-E07-S04-01, TP-E07-S04-01 | P0 / Sprint 7 | LINKED |
| FR-303 | Honesty count computed server-side; not suppressible by tier or purchase                      | S-40       | E07-S06          | —               | TI-E07-S06-01                | P0 / Sprint 8 | LINKED |
| FR-304 | In-app calls never reveal phone numbers; call metadata per retention table                    | S-43       | E09-S09          | identity-safety | TI-E09-S09-01                | P0 / Sprint 9 | LINKED |

### FR-400 Fires (PRD M9) — test class: load + field 3G

| Req    | Requirement (abbreviated)                                                  | Screens   | Epic/Story       | Consent row     | Test cases                   | Release       | Status |
| ------ | -------------------------------------------------------------------------- | --------- | ---------------- | --------------- | ---------------------------- | ------------- | ------ |
| FR-401 | Fire entry Tier 1+; capacity/waitlist/eject host-controllable in real time | S-60–S-62 | E09-S01, E09-S03 | identity-safety | TI-E09-S01-01, TE-E09-S03-01 | P0 / Sprint 9 | LINKED |
| FR-402 | Ember: one per attendee per fire, 24 h redemption, co-attendees only       | S-65      | E09-S07          | —               | TI-E09-S07-01, TP-E09-S07-01 | P0 / Sprint 9 | LINKED |
| FR-403 | Live audio degrades to listen-only, never drops below bandwidth floor      | S-62      | E09-S04          | —               | TE-E09-S04-01                | P0 / Sprint 9 | LINKED |

### FR-500 Games (PRD M8)

| Req    | Requirement (abbreviated)                                                   | Screens   | Epic/Story       | Consent row    | Test cases                   | Release       | Status |
| ------ | --------------------------------------------------------------------------- | --------- | ---------------- | -------------- | ---------------------------- | ------------- | ------ |
| FR-501 | Oware legality server-validated; Glicko-2 ratings per Doc 08                | S-70–S-73 | E10-S01, E10-S03 | play-portraits | TC-E10-S01-01, TP-E10-S01-01 | P0 / Sprint 9 | LINKED |
| FR-502 | Game conduct feeds suban only per consent map; skill never affects matching | S-70–S-73 | E10-S04          | play-portraits | TC-E10-S04-01                | P0 / Sprint 9 | LINKED |

### FR-700 Companions (PRD M12) — P0 subset: WhatsApp OTP and pod alerts only

| Req    | Requirement (abbreviated)                                                                       | Screens | Epic/Story          | Consent row     | Test cases    | Release        | Status |
| ------ | ----------------------------------------------------------------------------------------------- | ------- | ------------------- | --------------- | ------------- | -------------- | ------ |
| FR-701 | WhatsApp flows OTP-gated; no room content beyond consented Gate packs / own-number pod forwards | —       | E13-S05 (P0 subset) | identity-safety | TI-E13-S05-01 | P0 / Sprint 10 | LINKED |

### FR-800 Admin (PRD M13)

| Req    | Requirement (abbreviated)                                                                               | Screens   | Epic/Story       | Consent row | Test cases                   | Release        | Status |
| ------ | ------------------------------------------------------------------------------------------------------- | --------- | ---------------- | ----------- | ---------------------------- | -------------- | ------ |
| FR-801 | Admin actions least-privilege, MFA-gated, immutably audited; T&S evidence auto-redacts non-case parties | admin app | E16-S01, E12-S03 | —           | TI-E16-S01-01, TP-E16-S01-01 | P0 / Sprint 10 | LINKED |

## 3. P0 non-functional requirements

| Req family  | Binding targets (abbreviated)                                                                                                                       | Epic/Story                                                                                                                  | Test class                       | Release           | Status |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- | -------------------------------- | ----------------- | ------ |
| NFR-101–106 | Reference-device/3G budgets: cold start ≤3.5 s p90, Opus voice budgets, fire latency ≤400 ms p90, ≤15 MB/day, offline reconcile without duplication | E01, E09-S10; S0-010 spike evidence                                                                                         | Device-lab benchmarks            | P0 / Sprints 9–10 | LINKED |
| NFR-201–203 | 99.9% core / 99.95% fire-window; RPO ≤5 min, RTO ≤60 min                                                                                            | E17-S01, E17-S02, E17-S07                                                                                                   | Load + restore rehearsal         | P0 / Sprint 10    | LINKED |
| NFR-301–303 | TLS 1.3, AES-256, envelope-encrypted voice/biometric, ASVS/MASVS L2, rate limits, device attestation                                                | E02-S03, E02-S04, E17-S09; design inputs in `internal/architecture/threat-model-v0.md`                                      | Pen test + SAST/DAST             | P0 / Sprint 10    | LINKED |
| NFR-401–403 | Act 843 registration, in-region residency, purpose-bound processing, biometric templates never for matching                                         | E03, E15; DPIA inputs in `internal/architecture/dpia-inputs.md`; register in `internal/architecture/data-classification.md` | DPIA audit                       | Pre-P0            | LINKED |
| NFR-501–502 | String externalization; EN/Twi (Pidgin per founder decision §24.7); WCAG 2.2 AA, TalkBack/VoiceOver, 48 dp targets                                  | E01-S05, E01-S06                                                                                                            | Axe + manual SR passes           | P0 / Sprints 1–3  | LINKED |
| NFR-601–603 | Crash-free ≥99.6%; golden-path monitor every 5 min; every FR has ≥1 automated test; chaos/fuzz for alternation, allowance, Sika Shield, age gate    | E02-S07, E06-S12, E17-S06                                                                                                   | Synthetic monitors + this matrix | P0 / Sprint 10    | LINKED |

NFR-603 note: this matrix is itself the NFR-603 control — the release gate
checks that every FR row above has existing, passing test cases before P0.

## 4. End-to-end traced sample (verification artifact for S0-013)

Requirement **FR-202** traced through every link:

1. **Requirement (SRS):** "Sow MUST require ≥20 s cumulative playback of the
   target's Voice of Introduction, verified server-side by playback
   telemetry."
2. **PRD module:** M4 (Seed economy) — server authority over the sow
   gesture; "voice precedes face" product law (`agent_plan.md` §4).
3. **Screens (Doc 06):** S-21 person page — voice player progress ring feeds
   the 20 s sow-arm rule; S-22 sow composer — Sow stays disabled until armed,
   tooltip explains why.
4. **Epic/story (agent_plan §20):** E06-S03 "Server-verified 20-second
   listening eligibility"; supporting E06-S04 (sow composer deliberate send).
5. **Consent row:** `—` (playback telemetry is service-provision data; it
   must not leak into analytics events per the Doc 08 §3 taxonomy and must
   never be exposed to the sower — see FR-205).
6. **Test cases:**
   - `TI-E06-S03-01` (Testcontainers): playback telemetry below 20 s → sow
     rejected with stable machine code; ≥20 s cumulative across sessions →
     sow armed. Asserted against real MongoDB state, not client-reported
     flags.
   - `TP-E06-S03-01` (property/fuzz): randomized telemetry sequences
     (duplicates, out-of-order, replayed heartbeats, multi-device) never arm
     sow below 20 s of unique cumulative playback and never double-count.
   - `TC-E06-S03-02` (contract): API error uses a stable machine code with a
     safe localized presentation mapping (`agent_plan.md` §11).
7. **Test class:** API contract + fuzz (SRS §4 row FR-200).
8. **Release/sprint:** P0, Sprint 5 (agent_plan §21); metric gate affected:
   seed→sprout ≥25% (`agent_plan.md` §22).
9. **Status:** LINKED — implementation not yet claimed.

This row satisfies the S0-013 verification criterion "requirement sample
traced end-to-end": every link from requirement to test case and release is
present and checkable.

## 5. Maintenance rules

- A story may not move to `IN REVIEW` with its requirement rows missing test
  case IDs (Definition of Done, `agent_plan.md` §19).
- New requirement IDs are appended; IDs are never recycled after removal —
  deprecate with a note instead.
- Any change to a consent-map row, safety control, or product law requires
  the recorded decision process in `agent_plan.md` §1 before this matrix is
  edited to match.
- Synthetic actors used by the golden-path monitor and integration tests are
  defined in `packages/test-fixtures/` and must satisfy the fixture policy
  there.
