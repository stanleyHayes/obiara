# Launch decision evidence inventory

- Owner: Release manager
- Last reconciled: 2026-07-27
- Engineering backlog: **complete**
- Synthetic staging qualification: **complete**
- Production: **blocked**

This inventory is the handoff from completed engineering delivery to launch
decision readiness. It separates evidence the repository can prove from
decisions, contracts, credentials, people and production actions that require
named external owners. A committed document, synthetic fixture or green CI run
can never satisfy an external gate.

## Gate inventory

| Gate | Authority | Owner role | Current evidence | State | What closes it |
| --- | --- | --- | --- | --- | --- |
| Exact candidate and full engineering checks | Repository | Engineering lead | `internal/quality/backlog-closure.md`, CI for exact SHA | Ready for synthetic staging | Exact-SHA release evidence remains green and current |
| Security policy and synthetic DAST | Repository | Security engineer | `deploy/security/`, security closure tests | Ready for synthetic staging | Exact-SHA checks and an unexpired synthetic evidence record |
| Backup/restore orchestration | Repository | Data platform | `deploy/atlas/rehearsal/`, DR tests | Ready for synthetic staging | Synthetic rehearsal record passes RPO/RTO and integrity validation |
| Rollback and hypercare contract | Repository | Release manager | `deploy/release/closeout-runbook.md`, release-bundle tests | Ready for synthetic staging | Exact known-good SHA, current rehearsal ref and named synthetic rehearsal owners |
| Ghana-only or Africa-region interpretation | External decision | Founder and DPO/legal | `deploy/render/provider-residency-feasibility.md` | Blocked | Recorded interpretation and signed Act 843 transfer/DPIA decision |
| Production topology | Mixed: external decision then repository ADR | Architecture owner | Candidate options only; no approved production Blueprint | Blocked | Residency decision, provider evidence and an accepted follow-up ADR |
| Atlas production tier, region, backup and support | Provider/procurement | Procurement and data platform | `deploy/atlas/production.yaml` is deliberately blocked | Blocked | Written provider evidence, DPA/terms, purchased tier and production restore rehearsal |
| Object storage, CDN, keys, replicas and deletion | Provider/procurement | Procurement and security | No selected production vendor | Blocked | Approved vendor and complete location/retention/key/deletion map |
| LiveKit media, egress, analytics and support boundary | Provider/procurement | Procurement and realtime owner | Replaceable adapter and feasibility evidence only | Blocked | Written provider evidence, approved contract and measured Ghana path |
| Resend transfer posture or replacement | Provider/procurement | DPO/legal and communications owner | Sensitive content prohibited; US processing documented | Blocked | Approved minimized-email posture or approved replacement |
| Ghana representative latency and device budgets | External execution | QA and platform | Local/synthetic performance evidence only | Blocked | Named Ghana 3G/fixed-network and reference-device run with immutable evidence |
| App-store accounts, signing and release credentials | Credential/store | Mobile release owner | No production credential in the repository | Blocked | Controlled accounts, secret-store binding and reviewed store submission |
| UAT cohort consent, training and completion | People/cohort | UAT lead | UI and schema use synthetic aggregates only | Blocked | Real consented cohort, training evidence, zero critical findings and distinct review |
| Circles, hosts, verification, support and T&S cover | People/operations | Launch operations lead | Operator workflow implemented; sample values are synthetic | Blocked | Required populated circles, certified hosts and named staffed coverage |
| Founder go/no-go | External decision | Founder | Not recorded | Blocked | All prerequisite gates pass and a distinct, time-bound decision is recorded |
| Production resource creation and traffic activation | Production action | Platform and release managers | Prohibited and absent | Blocked | Approved topology, credentials, production rehearsal and explicit release authority |

## Repository-controlled Launch 1 scope

Launch 1 can close only these items:

1. A strict, executable registry for the gates above and their evidence.
2. Adversarial tests proving synthetic, stale, self-approved or wrong-scope
   evidence cannot satisfy production.
3. An operator view that shows owner, authority, evidence freshness,
   dependencies and blocker state without offering a launch action.
4. An exact-SHA synthetic staging qualification package that can qualify
   staging only.
5. Full repository verification and synchronized `main`.

Launch 1 is complete. The remaining external gates are promoted into the final
repository handoff milestone in
[`external-gate-handoff.md`](external-gate-handoff.md); their real-world acts
remain blocked.

## Non-claims

Launch 1 does not provision Render, Atlas, LiveKit, storage, Resend or store
resources; purchase a service; create or rotate a real secret; recruit or
contact a cohort; submit an application; accept legal risk; sign a DPIA; run
against production; or approve launch. These gates remain blocked until their
named human authority supplies independently reviewable evidence.
