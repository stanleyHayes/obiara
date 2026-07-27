# External gate handoff

- Owner: Release manager
- Last reconciled: 2026-07-27
- Product and engineering backlog: **complete**
- Repository launch-decision tooling: **complete**
- External-gate handoff tooling: **complete**
- Production: **blocked**

This is the final boundary between agent-buildable work and actions that require
real authority, accounts, contracts, people, devices, networks or production
control. Each remaining gate must have an executable evidence contract, a
blocked synthetic example, an operator handoff and a named external owner.
Passing repository tests never substitutes for the external act.

## Handoff map

| External gate                              | Named authority                                 | Repository handoff                                                                 | Launch 2 task  | External act that remains                                              |
| ------------------------------------------ | ----------------------------------------------- | ---------------------------------------------------------------------------------- | -------------- | ---------------------------------------------------------------------- |
| Ghana-only or Africa-region interpretation | Founder, DPO/legal                              | Residency/DPIA decision packet and strict record validator                         | L2-002         | Interpret Act 843 posture and record the signed decision               |
| Production topology                        | Architecture owner after founder/legal decision | Existing topology options, launch-gate dependency graph and follow-up ADR template | L2-002         | Select a compliant option after provider evidence and accept the ADR   |
| Atlas production service                   | Procurement and data platform                   | Provider diligence contract and Atlas evidence checklist                           | L2-003         | Obtain written provider evidence, approve terms and procure the tier   |
| Object storage/CDN/KMS                     | Procurement and security                        | Provider diligence contract and storage location map template                      | L2-003         | Select the service, approve terms and verify replicas, caches and keys |
| Live media and egress                      | Procurement and realtime owner                  | Provider diligence contract and media boundary template                            | L2-003         | Obtain written evidence, approve the service and measure the real path |
| Transactional communications               | DPO/legal and communications owner              | Provider diligence contract and minimized-content decision template                | L2-003         | Approve Resend posture or procure a replacement                        |
| Ghana device/network evidence              | QA and platform                                 | Field manifest, deterministic validator, budgets and execution runbook             | L2-004         | Run named Ghana devices on representative 3G and fixed networks        |
| Production credentials                     | Credential custodians                           | Secret-free custody and signing-ceremony evidence template                         | L2-005         | Create controlled accounts and bind secrets outside Git                |
| Mobile store release                       | Mobile release owner and reviewer               | Store-readiness evidence template                                                  | L2-005         | Control store accounts, review the build and submit it                 |
| Real UAT cohort                            | UAT lead and safety reviewer                    | Consent/training/completion aggregate evidence template                            | L2-005         | Invite, consent, train and support the real cohort                     |
| Circles, hosts and operational cover       | Launch operations lead                          | Density, certification and staffed-coverage evidence template                      | L2-005         | Populate circles, certify hosts and staff verification/support/T&S     |
| Founder go/no-go                           | Founder with distinct release reviewer          | Executable 17-gate registry and admin decision desk                                | L1-002, L2-006 | Review all current evidence and record the time-bound decision         |
| Production activation                      | Platform and release change authorities         | Fail-closed activation evidence category with no mutation port                     | L1-002         | Execute an approved change only after every dependency passes          |

## Completion rule

The repository backlog is complete only when every row above has:

1. a strict machine-validated metadata contract;
2. an intentionally blocked synthetic fixture;
3. a human-readable execution or decision packet;
4. operator visibility without an approval or activation shortcut;
5. adversarial tests for synthetic, stale, self-approved, incomplete,
   wrong-scope and dependency-bypass evidence; and
6. a ledger entry with reproducible verification.

At that point, no agent-buildable task remains. The project is not production
ready: the final column still requires real human and provider action.

This completion rule passed on 2026-07-27. The integrated evidence is recorded
in [`repository-backlog-closure.md`](repository-backlog-closure.md).
