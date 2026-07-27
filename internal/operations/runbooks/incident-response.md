# Incident Response Runbook v0 — Obiara

Status: Pre-P0 baseline (E12-S08)
Prepared: 2026-07-27
Owner: trust, safety and care workstream
Sources: `agent_plan.md` §15, §16, §22; Obiara Doc 09 §3, §7;
`internal/architecture/threat-model-v0.md`

This runbook covers incident response, evidence handling, and the
regulatory 72-hour reporting clock for the Obiara platform. It is a
living document: every post-incident review updates it (blameless, with
actions tracked as backlog work).

## 1. Severity scheme

| Severity | Definition                                                                                 | First response        | Examples                                                                             |
| -------- | ------------------------------------------------------------------------------------------ | --------------------- | ------------------------------------------------------------------------------------ |
| SEV-1    | Member physical safety at risk; Tier-A abuse in progress; data breach of C4 data           | 15 minutes, 24/7 page | Scam-arc case with imminent meeting; biometric/ID leak; active account-takeover wave |
| SEV-2    | Product-law violation live; privileged-action misuse; provider compromise affecting safety | 1 hour                | Allowance/alternation bypass; admin evidence abuse; verification provider breach     |
| SEV-3    | Degraded core journey without data exposure                                                | 4 hours               | Fire-window capacity loss, OTP provider outage, media pipeline stall                 |
| SEV-4    | Operational anomaly without member impact                                                  | next business day     | Dead-letter accumulation, retry storms, metric drift                                 |

Priority order in any conflict (Doc 09 §1): member physical safety, member
dignity and privacy, platform integrity, growth.

## 2. Response flow

1. **Detect.** Signals: safety queues (SLA dashboards), Sentinel ladders,
   golden-path monitors, provider error budgets, member reports, staff.
2. **Declare.** The on-call engineer declares severity and opens the
   incident channel. SEV-1/2 pages the safety lead and the DPO
   (once appointed) immediately.
3. **Contain.** Kill switches first: Sow, Fires, AI, Payments, Gate flags
   (E02-S08). Suspend affected provider adapters, not the whole service.
4. **Protect evidence.** Place legal holds before any cleanup (see §3).
   Never let containment destroy evidence — no log rotation, no cache
   flushes on affected systems.
5. **Mitigate.** Restore the golden path: register → verify → sow →
   sprout → room. Verify with the synthetic monitor before un-flagging.
6. **Communicate.** Member-facing copy follows the voice rules: what
   happened, what to do, no blame, no pressure language (banned-copy
   lint applies to incident comms too).
7. **Review.** Blameless post-incident review within 5 business days for
   SEV-1/2; actions enter the backlog as claimed tasks.

## 3. Evidence handling

- **Preserve first.** Place a legal hold (privacy service: `PlaceHold`)
  for any incident involving member data, abuse, or potential regulation.
  Holds block deletion and retention automation until lifted — no
  exceptions for convenience.
- **Least exposure.** T&S evidence is viewed only through the redacted
  evidence service (E12-S03): subject visible, third parties masked,
  purpose declared (triage/appeal/legal), every access immutably audited.
- **Chain of custody.** Evidence exports are append-only and
  reference-keyed (case id + export id). Fraud-victim exports follow the
  victim path (E12-S11) with no-shame copy.
- **Voice evidence.** Transcribed for review; original retained per the
  binding retention table (T&S evidence: case life + 2 years or legal
  hold).
- **Privileged actions.** Every contain/hold/export action is itself an
  audited privileged action (action log + evidence access log).

## 4. The 72-hour reporting clock (Cybersecurity Act 2020 + Act 843)

When a breach of personal data is confirmed or reasonably suspected:

| Time  | Action                                                                                            |
| ----- | ------------------------------------------------------------------------------------------------- |
| T+0   | Incident declared; DPO (once appointed) and safety lead paged; containment begins                 |
| T+2h  | Breach scope draft: data classes affected, member count estimate, systems involved                |
| T+24h | Evidence preservation complete; CERT-GH liaison informed per Doc 09 §7                            |
| T+48h | Member impact assessment; notification decision recorded with reasons                             |
| T+72h | Report filed to the Data Protection Commission and CERT-GH; DPO sign-off; clock evidence archived |

If scope is not confirmed by T+72h, file an interim report with what is
known and mark it interim. The clock runs from reasonable suspicion, not
certainty. Every hour of the timeline is recorded in the incident record.

## 5. Provider outage playbooks

- **SMS/OTP down:** OTP falls to WhatsApp (routing ladder already built);
  if both fail, suspend registration, keep sessions alive, banner on
  clients. Watch resend counters for abuse during the brownout.
- **Verification provider down:** Ghana Card submissions route to the
  manual fallback queue (already built); notify members with honest wait
  copy; watch queue age metrics.
- **LiveKit down:** fires fall to listen-only degradation ladder (E09-S04
  client path); host notified; resumption re-checks room state.
- **MongoDB down:** API reports dependency-degraded `/ready` (already
  built); worker jobs fail into backoff/dead letters; do not accept
  writes on partial connectivity — better to fail visibly than corrupt.
- **MoMo aggregator down (P1+):** payments suspend via flag; no own
  float fallback (Doc 09 §7); reconciliation backlog reviewed on resume.

## 6. Safety incidents (care)

- Self-harm signals go to the care queue (E12-S05), never to enforcement.
  Trained humans respond with resource-first scripts only.
- Tier-A safety cases have p50 <2h / p95 <8h SLAs (Doc 09 §3); breaches
  are themselves a SEV-2 signal.
- Moderation welfare: any incident involving graphic evidence triggers a
  welfare check for the reviewers involved (Doc 09 §7).

## 7. Maintenance

Update after every SEV-1/2 post-incident review, every new kill switch,
every new provider boundary, and before P0 UAT. Rehearse this runbook at
least once before launch (Sprint 11 backup/incident rehearsal gate).
