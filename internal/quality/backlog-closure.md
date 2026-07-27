# Backlog story closure audit

- Auditor: `codex-root`
- Audited: 2026-07-27
- Source: `agent_plan.md` E01–E18 story catalog and completed task ledger
- Production readiness: **blocked**

## Method

The audit extracted all 182 unique `E##-S##` story identifiers and compared
them with task-ledger labels, then inspected code and focused tests for every
apparent miss. Combined labels such as `E15-S01+S02` and
`E18-S06+S07+S08` account for five apparent misses. Eleven stories were
implemented under an earlier or broader task label. Two P2 stories were
genuinely absent, opened as S8-074 and S8-073, and are now implemented and
DONE.

## Implemented stories whose exact ID was absent from a row label

| Story | Implementation evidence | Verification evidence |
| --- | --- | --- |
| E11-S08 liveness adapter/human uncertainty | `services/api/internal/verification/liveness/`; S2-018 | Exact-live-only pass, uncertainty/manual review and temporary-ref removal tests |
| E13-S08 delivery observability/fallback/opt-out | notification preferences S7-010, routing S8-009, WhatsApp S8-006 and email S8-010 | GoMock fallback/suppression, idempotent inbox and Mongo delivery-log Testcontainers proofs |
| E14-S08 finance admin/exports/four-eyes pricing | `apps/admin/app/finance/` | Admin reducer tests prove reason/redaction/export confirmation and distinct second approver; admin test/typecheck green in this audit |
| E14-S10 security/reconciliation/chaos tests | reconciliation S8-044, performance S8-067, DAST S8-068 and DR S8-070 | Signed/replay/concurrency/malformed-row, provider-failure, bounded load, passive DAST and restore-corruption proofs |
| E15-S02 consent-aware event pipeline | combined analytics S8-030 | Producer registry, opt-out emits nothing and pseudonymized Mongo persistence |
| E15-S05 recomputable marks/decay/anti-gaming | suban ledger S8-033 | Decay, suppression, permanent-fraud, cap and append-only Testcontainers proofs |
| E16-S02 verification operations | admin UI S2-020 and audited backend S2-023 | Recent-MFA evidence gate, HMAC refs, transaction/audit and rendered queue QA |
| E16-S04 T&S/care/panel operations | safety S8-013/S8-015, care S8-016 and Mpanyimfo S8-019 | Redaction/access, action-ladder, approved-script and recusal/quorum/appeal tests |
| E16-S05 finance/reconciliation operations | `apps/admin/app/finance/` plus S8-044 | Redacted exception workflow and read-only ledger reconciliation evidence |
| E16-S08 four-eyes pricing/policy | `apps/admin/app/finance/finance-model.ts` | Same-person rejection, bounded price/reason and second-approval reducer tests |
| E16-S09 operational dashboard/case SLA | analytics S8-051, safety S7-026 and observability S8-069 | Exact denominators, oldest-SLA routing, unresolved Tier-A and owned burn-alert dashboards |

## Combined ledger labels

| Stories | Closing task |
| --- | --- |
| E18-S02 and E18-S03 | S8-059 and backend S8-062 |
| E18-S05 | S8-061 |
| E18-S07 and E18-S08 | S8-071 |

## Genuine gaps found

| Story | Task | Required closure |
| --- | --- | --- |
| E13-S07 P2 Gate links and USSD | S8-074 | Consent-safe bounded link/session boundary, no contact exposure/provider call, deterministic and persistence proofs |
| E14-S09 P2 diaspora payment isolation | S8-073 | Isolated provider/payment boundary, idempotency/privacy/ledger invariants and no real-fund/provider call |

## Completion rule

The backlog is complete only when S8-073 and S8-074 are DONE, no ledger row is
`TODO`, `READY`, `IN PROGRESS`, `PENDING` or `BLOCKED`, focused verification is
green, `main` is synchronized, and the production blockers remain truthfully
documented. Completion of engineering backlog does not provision providers,
approve residency/DPIA, release stores, deploy production, recruit members or
declare launch readiness.

## Final result

All 182 unique E01–E18 story identifiers now map to completed implementation
evidence. S8-073 and S8-074 closed the two genuine code gaps found by this
audit. The repository-wide `pnpm run check` gate completed with all 58
TypeScript workspace tasks and the complete Go suite green. The ledger has no
remaining TODO, READY, IN PROGRESS, PENDING or BLOCKED task row.
