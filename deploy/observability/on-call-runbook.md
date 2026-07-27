# Service reliability and on-call runbook

- Owner: Platform engineering
- Safety escalation owner: Trust and Safety lead
- Last reviewed: 2026-07-27
- Production status: **blocked**

## Operating law

Dashboards and alerts are generated from the finite specification in
[`slo.yaml`](slo.yaml). Missing telemetry, an unknown objective, an unowned
alert, or an unresolved release-blocking error-budget breach blocks release;
absence of data is never treated as health. Labels are low-cardinality service
facts only. Member, session, request, correlation, contact, content and secret
values are prohibited.

The primary on-call acknowledges a page within five minutes and appoints an
incident commander. Safety pages also notify the Trust and Safety lead and
must not wait for platform diagnosis before protective routing. Incident notes
use opaque case/change references and the existing redacted incident packet.

| Severity | Meaning                                                                                  | Acknowledge      | Update cadence | Authority                                              |
| -------- | ---------------------------------------------------------------------------------------- | ---------------- | -------------- | ------------------------------------------------------ |
| P0       | Confirmed active safety/data breach or broad critical outage                             | 5 min            | 15 min         | Incident commander plus T&S/privacy lead as applicable |
| P1       | Material degradation, Tier-A routing breach, payment integrity risk or dependency outage | 5 min            | 30 min         | Incident commander                                     |
| P2       | Slow-burn budget loss or bounded non-critical failure                                    | 4 business hours | Daily          | Service owner                                          |

## Availability or latency

1. Confirm the exact objective, long/short windows and whether telemetry is
   complete; never dismiss a page from a single healthy instance.
2. Compare `/live`, dependency-aware `/ready`, request rate, bounded route,
   status and latency signals against the last known-good deploy.
3. Freeze promotion. Roll back the application only when database compatibility
   is preserved; otherwise disable the affected capability through the governed
   kill-switch path.
4. Record detection, acknowledgement, impact window, candidate SHA, mitigation
   and current error budget without member identifiers.

## Dependency readiness

Treat MongoDB readiness loss as unavailable even if `/live` remains healthy.
Check regional provider status, pool pressure, timeout and Atlas topology
evidence. Do not weaken TLS, public-access, identity or residency controls to
restore service. Escalate to the database owner and initiate the approved
restore/DR runbook only when corruption or unrecoverable loss is evidenced.

## Safety routing

Immediately provide the fallback human Tier-A queue to the T&S lead. Preserve
the minimal redacted report envelope and immutable timestamps. Do not copy
evidence into chat, dashboards or incident notes. Platform recovery and member
protection proceed in parallel; no platform operator adjudicates a report.

## Worker dead letters

Identify the finite job name, oldest age and failure class. Pause only the
affected scheduler lane, retain its dead letter and retry through the
idempotent replay path after repair. Never delete a failed job, bypass inbox
deduplication or replay an unbounded batch.

## Closure

Two roles confirm restored SLI health and completed safety/privacy follow-up.
P0/P1 incidents require a blameless review with contributing controls,
evidence, owner and dated backlog actions. Closing an alert does not close an
incident automatically, replenish an error budget, or authorize release.
