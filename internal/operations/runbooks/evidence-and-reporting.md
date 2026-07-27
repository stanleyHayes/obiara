# Evidence Handling and Regulatory Reporting Runbook v0 — Obiara

Status: Pre-P0 baseline (E12-S08, companion to incident-response.md)
Prepared: 2026-07-27
Owner: trust, safety and care workstream → DPO (when appointed)
Sources: Obiara Doc 09 §3, §7; `internal/architecture/data-classification.md`;
`internal/architecture/dpia-inputs.md`

## 1. Evidence classes and who may see what

| Class                | Content                             | Access rule                                                                                   |
| -------------------- | ----------------------------------- | --------------------------------------------------------------------------------------------- |
| Case metadata        | Tier, category, surface, timestamps | T&S desk (triage purpose)                                                                     |
| Member text evidence | Report reasons, flagged content     | Redacted viewer only; third parties masked; purpose declared                                  |
| Voice evidence       | Room/pod voice                      | Transcribed for review; original retained per retention table; legal hold required for export |
| Identity/biometric   | Ghana Card artifacts, liveness      | Break-glass only, HSM-held keys, full audit                                                   |
| Financial            | Ledgers, receipts                   | Finance desk + four-eyes for any adjustment                                                   |

Every access is declared with a purpose (triage, appeal, legal) and
immutably audited before the first byte is read (E12-S03). There is no
browse mode.

## 2. Preservation procedure

1. Trigger: any of — Tier-A report, regulator contact, litigation risk,
   safety incident, fraud pattern confirmed.
2. Place the legal hold immediately via the privacy service. The hold:
   - blocks new deletion requests for the account,
   - blocks retention automation for the held data,
   - is itself recorded with actor, reason, timestamp.
3. Confirm hold coverage: identity artifacts, room voice, T&S evidence,
   ledger entries, logs with member identifiers.
4. Review holds weekly; lift with the same audit ceremony when the matter
   closes. Stale holds are a compliance finding, not a convenience.

## 3. Export procedure

- **Member data export (FR-106):** privacy export request; the assembler
  collects the member's own records with secrets and other members' data
  stripped; delivered within 72 hours.
- **Fraud-victim export:** specialist flow (E12-S11): evidence bundle
  with reporting guidance and no-shame copy; regulator/police packaging
  follows counsel (EOCO / Police Cybercrime Unit referrals, Doc 09 §6).
- **Regulator export:** DPO approves, purpose recorded, least-exposure
  principle applies even to authorities; the export itself is audited.

## 4. Regulatory reporting contacts (Doc 09 §7)

- **Data Protection Commission:** registration + breach reports
  (Act 843); DPO owns the relationship once appointed (pre-P0).
- **CERT-GH:** Cybersecurity Act 2020 liaison; 72-hour incident clock per
  the incident runbook.
- **EOCO / Police Cybercrime:** fraud syndicate referrals with
  counsel-approved evidence packages (Doc 09 §6).
- **GBV-response partner NGO:** warm referrals for offline member safety
  (Doc 09 §4); contact list held by the care workstream, not in-repo.

## 5. Records to keep for every report

- Incident declaration time and declarer.
- Scope assessment and the data classes touched.
- Containment actions and their timestamps (all privileged, all audited).
- Clock milestones: T+2h draft, T+24h liaison, T+48h decision, T+72h filing.
- Post-incident review link and tracked remediation tasks.
