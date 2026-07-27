# Sprint 0 founder checkpoint

- Task: `S0-014`
- Review date: 2026-07-26
- Reviewer: `codex-root`
- Baseline commit: `47165a4`
- Scope: Sprint 0 evidence and integrated repository health
- Engineering verdict: **conditional proceed to Sprint 1**
- Production verdict: **no-go**
- Founder approval: **pending**

## Executive decision

Sprint 0 produced a runnable monorepo, accepted architecture decisions, a
working Go/Mongo hexagonal spike, three buildable clients, initial security and
privacy controls, provider/residency evidence, traceability fixtures, and
representative Outfit UI work. Sprint 1 engineering may continue on stories
that do not depend on the open decisions below.

Sprint 0 is not unconditionally approved. The current `main` security workflow
fails on a reachable Go vulnerability, the local Docker environment cannot
currently reproduce the Mongo Testcontainers proof, and founder/legal/provider
decisions remain open. No production infrastructure, real member data, or
external provider commitment is authorized by this checkpoint.

## Integrated gate record

| Gate                                                   | Result      | Evidence                                                                | Disposition                                                                                               |
| ------------------------------------------------------ | ----------- | ----------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| All Sprint 0 implementation tasks claimed and reviewed | PASS        | `S0-003` through `S0-013` and `S0-015` are `DONE` in `agent_plan.md`    | S0-014 closes the evidence pass, not founder approval                                                     |
| Architecture baseline                                  | PASS        | ADR-0001 through ADR-0006 are accepted and indexed                      | REST/OpenAPI, Go hexagon, Mongo, Expo and Render Blueprint remain the implementation defaults             |
| Exact dependency installation                          | PASS        | `pnpm install --frozen-lockfile`                                        | 13 workspace projects, lockfile unchanged                                                                 |
| Peer dependency compatibility                          | PASS        | `pnpm peers check`                                                      | No peer issues                                                                                            |
| Lint, type-check and unit tests                        | PASS        | `pnpm run check`                                                        | 9/9 Turbo tasks and all non-integration Go tests pass                                                     |
| Client production builds                               | PASS        | `pnpm run build`                                                        | web, admin and Expo web export pass                                                                       |
| Go static analysis                                     | PASS        | `go vet ./...`                                                          | No findings                                                                                               |
| Workflow syntax                                        | PASS        | `actionlint`                                                            | CI and security workflow YAML passes                                                                      |
| Mobile ecosystem validation                            | PASS        | `pnpm dlx expo-doctor@latest apps/mobile`                               | 20/20 checks pass                                                                                         |
| JavaScript high/critical advisory threshold            | PASS        | `pnpm audit --json`                                                     | 0 high, 0 critical; 1 moderate transitive `uuid` advisory remains                                         |
| Go vulnerability threshold                             | **FAIL**    | `govulncheck@v1.1.4 ./...`; Security run `30210233021`                  | Reachable `GO-2026-5970` in `golang.org/x/text v0.37.0`; fixed in `v0.39.0`                               |
| Mongo integration reproducibility                      | **BLOCKED** | tagged repository test timed out after 150 seconds in Docker API `Info` | Previous S0-007 proof remains evidence, but this environment must be recovered and rerun                  |
| Repository formatting baseline                         | **FAIL**    | `pnpm run format`                                                       | The root check scans the active `.worktrees/` lane and numerous tracked files are also not Prettier-clean |
| CI on checkpoint claim                                 | PASS        | GitHub Actions CI run for `47165a4`                                     | Required build/test workflow is green                                                                     |
| Security on checkpoint claim                           | **FAIL**    | GitHub Actions Security run for `47165a4`                               | Failure is the reachable Go vulnerability above                                                           |
| Source diff hygiene                                    | PASS        | `git diff --check`                                                      | No whitespace errors                                                                                      |

## Sprint 0 outcome review

| Outcome                                       | Status                     | Evidence / limitation                                                                                                                        |
| --------------------------------------------- | -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Governed monorepo and delivery ledger         | PARTIAL                    | Shared task claims, CODEOWNERS, security policy, CI and Dependabot exist; root README, contributor guide and agent guide are not yet present |
| Reproducible polyglot workspace               | PASS                       | pnpm/Turborepo and Go workspace install, check and build successfully                                                                        |
| Architecture and dependency map               | PASS                       | Six ADRs and the pinned dependency matrix are present                                                                                        |
| Go hexagonal/Mongo feasibility                | PASS WITH ENVIRONMENT HOLD | Unit/static gates pass and S0-007 recorded a real Mongo 8 Testcontainers pass; current Docker runtime is unhealthy                           |
| Web/admin/mobile signature UI                 | PASS                       | All three clients build; rendered desktop/mobile reviews are recorded in the ledger                                                          |
| Android 8 and constrained-network feasibility | PARTIAL                    | Native release build proves `minSdk 24`; API 26/2 GB physical-device, Play split-size and field-network evidence remain release gates        |
| Provider/residency feasibility                | DECISION BLOCKED           | Render has no African region; production topology requires founder, DPO and legal approval                                                   |
| Initial threat model, classification and DPIA | PASS AS V0                 | Artifacts exist; they are explicitly pre-P0 and not signed assessments                                                                       |
| P0 traceability and synthetic fixtures        | PASS                       | All P0 FR/NFR rows are linked and the synthetic-persona policy compiles                                                                      |
| Supply-chain baseline                         | PARTIAL                    | JS high/critical gate passes; current Go vulnerability makes the overall security gate red                                                   |

## Blocking decisions and assigned follow-up

| ID     | Decision or corrective action                                                                                                                                     | Owner                                                         | Required before                                                 | Status                                                                                                    |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| CP-001 | Upgrade or otherwise remediate reachable `golang.org/x/text` `GO-2026-5970`, then obtain a green Security workflow                                                | Backend/platform owner; coordinate with active `S1-002` owner | Unconditional Sprint 0 engineering approval                     | CLOSED 2026-07-26 by S1-012; `x/text v0.39.0`, local govulncheck clean; Security run `30212221435` passed |
| CP-002 | Recover Docker Desktop, remove resource saturation and rerun Mongo Testcontainers integration                                                                     | Engineering environment owner                                 | Merging work that changes Mongo transaction/repository behavior | OPEN                                                                                                      |
| CP-003 | Decide Ghana-only versus approved-African-region residency and approve a production topology                                                                      | Founder + DPO/legal                                           | Provisioning production data services                           | OPEN                                                                                                      |
| CP-004 | Decide E2E room encryption versus consented safety processing                                                                                                     | Founder + architecture + legal/safety                         | E07 room implementation                                         | OPEN                                                                                                      |
| CP-005 | Confirm member-web P0 depth, P0 Pidgin timing and Android 8 binding floor                                                                                         | Founder/product                                               | First dependent client stories                                  | OPEN                                                                                                      |
| CP-006 | Name Jira, Render and MongoDB Atlas organization/environment owners                                                                                               | Founder/engineering lead                                      | External project/resource creation                              | OPEN                                                                                                      |
| CP-007 | Select identity/liveness, messaging, storage, realtime, MoMo, push and AI provider owners/shortlists                                                              | Product + engineering + privacy                               | Provider adapter production acceptance                          | OPEN                                                                                                      |
| CP-008 | Add root onboarding/governance documents and example environment contract; scope Prettier away from agent worktrees and establish a clean tracked-source baseline | Platform owner                                                | Sprint 1 foundation exit                                        | OPEN                                                                                                      |
| CP-009 | Produce API 26/2 GB physical-device evidence and AAB/bundletool split-size report                                                                                 | Mobile owner                                                  | Mobile release approval                                         | OPEN                                                                                                      |

## Founder checkpoint

The requested approval is:

1. allow Sprint 1 implementation to continue under the accepted ADRs;
2. keep production provisioning and real-member data prohibited;
3. record `CP-001` as satisfied while retaining the remaining founder gates;
4. answer the product, residency, encryption and ownership questions before
   their dependent stories begin.

Until that approval is recorded, the project state is **Sprint 1 engineering
allowed, production no-go, founder decisions pending**.

## Reproduction commands

```sh
pnpm install --frozen-lockfile
pnpm peers check
actionlint
pnpm run check
pnpm run build
go vet ./...
pnpm dlx expo-doctor@latest apps/mobile
pnpm audit --json
go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
go test -tags=integration \
  ./services/api/internal/member/adapters/outbound/mongodb \
  -count=1 -timeout=150s -v
pnpm run format
git diff --check
```
