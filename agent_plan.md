# Obiara Agent Plan

Status: Active; founder decisions pending
Plan version: 1.0  
Prepared: 2026-07-26  
Execution state: **ACTIVE — Sprint 1 conditional engineering; production no-go**
Planning horizon: P0 proof through P2 launch, with P3 captured as a roadmap  
Delivery cadence: two-week sprints; Sprint 0 is a two-week inception sprint

## 1. Purpose and approval rule

This file is the execution ledger for Obiara. It converts the strategy, PRD,
SRS, release plan, UX specification, architecture, data/AI specification,
trust-and-safety pack, localization pack, launch plan, brand package, and
engineering operating manuals into an implementable backlog.

Execution is authorized from 2026-07-26.

- Sprint 0 foundation work may begin immediately.
- External paid resources, production deployments, legal commitments, and
  provider contracts still require explicit founder approval.
- Any change to a Product Law, P0 cut line, safety control, consent purpose, or
  member-facing commercial rule requires a recorded architecture/product
  decision and founder approval.

## 2. Sources and precedence

Read sources:

1. `Obiara_Founding_Blueprint.docx`
2. `Obiara_Interaction_Model.docx`
3. `Obiara_00_Package_Guide.docx`
4. `Obiara_01_PRD.docx`
5. `Obiara_02_SRS.docx`
6. `Obiara_03_Revenue_Pricing_Model.docx`
7. `Obiara_04_MVP_Release_Plan.docx`
8. `Obiara_05_Brand_Design_Language.docx`
9. `Obiara_06_UX_Flows_Screens.docx`
10. `Obiara_07_Technical_Architecture.docx`
11. `Obiara_08_Data_Model_AI_Spec.docx`
12. `Obiara_09_Trust_Safety_Compliance.docx`
13. `Obiara_10_Content_Localization.docx`
14. `Obiara_11_Launch_Operations.docx`
15. `Obiara_Brand_Guidelines.docx` and the supplied production assets
16. `AI_Development_Workflow_Training_Manual.docx`
17. `AI_Native_Software_Engineering_Operations_Manual.docx`

Precedence when sources conflict:

1. Explicit founder decisions recorded in this plan
2. Trust, safety, law, consent, and data-retention obligations
3. PRD and SRS requirement IDs
4. MVP release cut lines
5. Finished brand guidelines and supplied assets
6. UX screen contract
7. Technical architecture and data/AI specification
8. Strategy and interaction model
9. Interim design-language content not superseded by the finished brand

## 3. Confirmed founder decisions and scope overrides

The following decisions are treated as approved inputs to planning:

- Monorepo.
- GitHub repository: `git@github.com:stanleyHayes/obiara.git`.
- Backend: Go using hexagonal architecture.
- Primary database: MongoDB.
- Member web client: Next.js, TypeScript, Material UI.
- Admin platform: separate Next.js, TypeScript, Material UI application.
- Mobile client: React Native with Expo, Android and iOS from one codebase.
- Deployment: Render Blueprint.
- Transactional email: Resend.
- Go unit testing: GoMock for boundary mocks.
- Go integration testing: Testcontainers against real containerized
  dependencies; mocks are not substitutes for persistence/provider integration
  coverage.
- Graph persistence: do not add a graph database speculatively. Start with
  MongoDB edge collections/materialized trust paths, profile real queries, and
  introduce a graph store behind a port only when measured requirements justify
  it.
- Product font family: Outfit. This founder decision supersedes Plus Jakarta
  Sans and Fraunces in the supplied brand documents for all product, admin and
  marketing UI. Supplied logo artwork remains authoritative.
- Delivery: verified work is rebased onto current `origin/main` and pushed
  directly to `main`; pull requests and feature branches are not required at
  this stage.
- Review: code review may happen asynchronously after direct-main delivery.
  Reviewers may correct issues in a separately claimed task and must leave a
  review note in the coordination ledger.
- All new dependencies use the latest stable, non-prerelease release available
  when the repository is bootstrapped, then are exactly pinned.

Approved scope override:

- PRD v1 originally says “no web member client.” This plan supersedes that
  non-goal: member web is in scope.
- Member web must obey the same product, consent, verification, alternation,
  safety, localization, and accessibility rules as mobile.
- Mobile remains the reference experience for device performance and
  voice/gesture interaction. Web gets accessible keyboard/pointer equivalents
  for every gesture.

Architecture substitutions:

- Go replaces the Kotlin/Spring or NestJS backend decision.
- MongoDB replaces PostgreSQL/Neo4j as the initial system of record and graph
  persistence strategy.
- Render Blueprint replaces Kubernetes/Terraform for the initial platform.
- These substitutions do not weaken server authority, ledger integrity,
  idempotency, auditability, privacy, or recovery requirements.

## 4. Product laws and non-negotiable invariants

Every epic, story, pull request, and release must preserve:

1. Intent over inventory.
2. Introduction over discovery.
3. Voice over text.
4. Character over appearance.
5. Circles over profiles.
6. Ceremony over gamification.
7. Slow is sacred.
8. Reputation is capital.
9. Measure unions, not minutes.
10. The community is the algorithm.

Hard technical invariants:

- Seeds are never sold.
- Conversations strictly alternate.
- No member-to-member money path exists.
- Romantic surfaces require Tier 1; sowing requires Tier 2.
- Age below 18 is a hard block with required data purge.
- Clients render state; clients never authoritatively decide state.
- No infinite feed, visibility boost, public follower graph, ad-tech SDK, or
  engagement-ranked recommendation feed.
- Voice precedes face and text at the specified thresholds.
- Consent is purpose-specific, versioned, revocable where permitted, and
  enforced at data-access boundaries.
- No analytics event contains raw conversation content, voice, or free text.
- All privileged and irreversible actions are auditable.

## 5. Scope map

### P0 — Proof, closed cohort

Target: approximately 300 members, Android-first Expo app plus responsive member
web, English and Twi, two campus circles and one professional circle.

In scope:

- Identity registration, OTP, Ghana Card/provider abstraction, liveness,
  verification fallback, age gate.
- Voice of Introduction, Doorway Question, photo vault, veil.
- Fie shell and first-run experience.
- Circles and minimal host tools.
- Complete seed/pod/sprout/water loop.
- Doorway and room state machines, strict alternation, pause, closure, honesty
  ribbon, first guided theme.
- Cloth v0.
- In-room asynchronous Oware.
- Weekly fires v1 and embers.
- Core notification rituals and caps.
- Sentinel v0, report flow, basic T&S and verification admin desks.
- WhatsApp OTP and pod alerts only.
- Foundational analytics, consent, export/deletion, observability, backup and
  incident controls.

Not in P0:

- Paid products, matchmaker marketplace, Nnoboa, Abusua Gate, tournaments,
  Ampe, Ɛbɛ, Anansesɛm, USSD, production AI matching, broad multi-city tooling.

### P1 — Foundation, waitlisted Accra

- Productized Tier-2 vouching and suban ledger.
- Nnoboa and auntie WhatsApp flows.
- Matchmaker pilot and MoMo escrow/payment rails.
- Paid membership/passes and premium fires without paid seeds or visibility.
- Guided themes 2–4, games wave, nightly fire grid.
- Pidgin/Ga market packs as approved.
- Mpanyimfo workflows, finance admin, iOS production readiness.
- Rules/graph matching first, then consented model-assisted ranking only after
  data and fairness gates are met.

### P2 — Launch, Greater Accra to Ghana

- Abusua Gate.
- Diaspora Tier 3.
- Registry and wedding-suite beta.
- Cloth harvest and weaver workflows.
- Tournaments/campus league.
- USSD.
- Additional language packs.
- Transparency reporting, flagship durbars, sponsorship pilot.

### P3 — Roadmap

- Kumasi/Takoradi density expansion.
- Nigeria market pack and closed-cohort replay.
- Friendship/mentor ring.
- Regulatory evaluation of susu custody.
- Kenya discovery.

## 6. Target monorepo

```text
obiara/
├── apps/
│   ├── web/                 # member-facing Next.js app
│   ├── admin/               # internal Next.js admin platform
│   └── mobile/              # Expo React Native app
├── services/
│   ├── api/                 # Go modular monolith / public API
│   ├── worker/              # Go async consumers and scheduled work
│   └── realtime/            # optional Go websocket/live coordination adapter
├── packages/
│   ├── api-client/          # generated TypeScript client and contracts
│   ├── config-eslint/
│   ├── config-typescript/
│   ├── design-tokens/       # finished brand tokens
│   ├── i18n/                # typed registry and market packs
│   ├── observability/
│   ├── test-fixtures/
│   ├── ui-web/              # shared MUI primitives for web/admin
│   └── validation/          # client-safe generated validators
├── internal/
│   ├── architecture/        # ADRs, diagrams, threat models
│   ├── product/             # traceability matrix and decision register
│   ├── quality/             # test plans, scorecards, release evidence
│   └── operations/          # runbooks, incident and launch material
├── deploy/
│   └── render/              # Blueprint notes and environment matrix
├── scripts/
├── .github/
├── AGENTS.md
├── CLAUDE.md
├── agent_plan.md
├── go.mod
├── go.work
├── package.json
├── pnpm-workspace.yaml
├── render.yaml
└── README.md
```

Tooling proposal:

- pnpm workspaces plus Turborepo for TypeScript/JavaScript task orchestration.
- Go workspace for Go services and internal tooling.
- OpenAPI 3.1 as the external client contract; generated TypeScript SDK.
- REST for client APIs and partner webhooks. WebSocket/SSE only for realtime
  state where justified. GraphQL from the older architecture is not mandatory.
- One command for lint, type-check, unit tests, contract generation checks, and
  builds across the monorepo.

## 7. Backend architecture

### 7.1 Deployment shape

Start as a modular monolith plus worker, not premature microservices:

- `api`: stateless HTTP API, authentication, authorization, commands, queries,
  partner webhook ingress.
- `worker`: durable jobs, outbox consumers, expiry/state transitions,
  notifications, media processing orchestration, reconciliation.
- Managed/external boundaries: MongoDB Atlas, object storage/CDN, LiveKit,
  identity/liveness provider, SMS/WhatsApp, MoMo, Resend, push providers and AI
  gateway.

Split a module into its own service only when load shape, security isolation,
failure isolation, or team ownership provides measured value.

### 7.2 Hexagonal module template

Each domain module follows:

```text
internal/<context>/
├── domain/          # entities, value objects, policies, events, errors
├── application/     # commands, queries, ports, orchestration
├── adapters/
│   ├── inbound/     # HTTP, job, webhook adapters
│   └── outbound/    # MongoDB, media, provider, email adapters
└── module.go        # composition root
```

Dependency rule:

- Domain imports only standard library/domain code.
- Application depends on domain and declared ports.
- Adapters depend inward and implement ports.
- Composition happens at service startup.
- Provider SDK types never cross into domain models.
- GoMock is used for application port tests, not for mocking domain logic.

### 7.3 Bounded contexts

1. Identity and Access
2. Member Profile and Verification
3. Consent and Privacy
4. Circles and Trust Paths
5. Seeds and Doorways
6. Courtship Rooms
7. Cloth and Ceremonies
8. Fires and Realtime
9. Games
10. Matching and Introductions
11. Suban
12. Trust, Safety and Care
13. Notifications and Localization
14. Companions
15. Commerce and Ledger
16. Admin and Configuration
17. Analytics and Audit

### 7.4 State, concurrency, and durability

- MongoDB transactions guard multi-document invariants where required.
- Every write command carries an idempotency key and actor/context metadata.
- Optimistic concurrency/version fields prevent stale state transitions.
- Strict unique indexes prevent duplicate rooms, active accounts per verified
  identity, duplicate allowances, and duplicated provider webhook effects.
- Durable outbox records are committed with domain changes and processed
  asynchronously.
- Inbox/dedup records make consumers at-least-once safe.
- All time-based transitions use server time and explicit IANA time zones.
- Alternation, allowance, expiry, pause, water, closure, and Gate flows are
  explicit state machines with transition tables and property-based tests.
- Redis may be added for ephemeral coordination/rate limits, but never as the
  sole source of durable product truth.

## 8. MongoDB data design

MongoDB is the authoritative operational store for the initial architecture.

Collection families:

- `accounts`, `sessions`, `roles`, `admin_access`
- `identity_verifications`, `identity_bindings`, `device_risk`
- `profiles`, `voice_assets`, `photo_assets`, `doorway_questions`
- `consent_records`, `privacy_requests`, `legal_holds`
- `circles`, `circle_memberships`, `vouches`, `trust_edges`
- `seed_allowance_entries`, `seeds`, `pods`, `doorway_threads`
- `rooms`, `room_events`, `pause_stones`, `theme_progress`
- `cloth_artifacts`, `ceremony_events`, `gate_reviews`
- `fires`, `fire_attendance`, `embers`
- `games`, `game_moves`, `ratings`
- `introductions`, `matching_explanations`
- `suban_events`, `suban_marks`
- `reports`, `safety_cases`, `case_evidence`, `panel_rulings`
- `notification_preferences`, `notification_deliveries`
- `orders`, `ledger_entries`, `payouts`, `reconciliations`
- `audit_events`, `analytics_events`, `outbox`, `inbox`
- `market_packs`, `feature_flags`, `configuration_changes`

Rules:

- Prefer bounded aggregates; do not embed unbounded event/message arrays.
- Store high-volume append-only room/game/audit events separately.
- Use immutable ledger entries rather than mutable balances.
- Keep sensitive identity/biometric records isolated with envelope encryption
  and narrower database credentials.
- Use TTL indexes only where automatic deletion is legally correct; preserve
  legal-hold paths.
- Use application-level field encryption for the most sensitive PII.
- Keep media in object storage; MongoDB stores metadata, ownership, retention,
  checksum and signed-access policy.
- Build trust-path queries initially from scoped edge collections and bounded
  traversal/materialized projections. Revisit a dedicated graph store only
  after profiled queries prove it necessary.
- Define index budgets and query plans for every production endpoint.
- Backups must meet RPO <= 5 minutes and RTO <= 60 minutes or the discrepancy
  must be explicitly approved before launch.

Required design artifacts before implementation:

- Collection catalog and ownership.
- JSON schema validation per collection.
- Index plan and expected cardinality.
- Data classification and retention map.
- Transaction/invariant matrix.
- Erasure/export strategy.
- Migration and backward-compatibility strategy.

## 9. Client architecture

### 9.1 Member web

- Next.js App Router, TypeScript strict mode, Material UI.
- Responsive, installable where useful, optimized for low bandwidth.
- Authenticated product surface, not a public browse/feed.
- Server rendering for public/low-risk surfaces; sensitive member state fetched
  through authenticated APIs with cache controls preventing leaks.
- Accessible pointer/keyboard alternatives for Hold, Sow, Stone and Gather.
- Web media recording and resumable upload with capability fallbacks.

### 9.2 Admin platform

- Separate Next.js app and deployment boundary.
- MUI component system shared at token/primitive level, not member navigation.
- SSO/MFA-ready authentication, RBAC/ABAC, least privilege and session
  hardening.
- Privacy-redacted evidence viewer, four-eyes controls and immutable audit.
- Modules: verification, T&S/care, circles/hosts, fires, matchmakers, panels,
  finance, market packs, feature flags, retention/legal holds and reporting.

### 9.3 Mobile

- Expo managed workflow unless a verified native requirement forces a
  development build/config plugin.
- Expo Router, TypeScript strict mode, React Native.
- Android 8 / low-memory / 3G acceptance remains the reference constraint;
  confirm Expo SDK compatibility during Sprint 0.
- Secure credential storage, device attestation adapters, resumable media,
  offline command queue and deterministic reconciliation.
- EAS Build/Submit/Update decisions documented; over-the-air updates never
  bypass native-store or safety review requirements.

### 9.4 Shared client rules

- Generated typed API client; no hand-maintained duplicate DTOs.
- Shared design tokens and localization keys.
- Platform-native view components; do not force DOM/RN component sharing.
- Server state uses a query cache with explicit privacy-aware invalidation.
- Local state is feature-scoped.
- Forms use schema-based validation aligned to API contracts.
- All screens ship loading, empty, error, offline and permission-denied states.
- Every irreversible action has deliberate confirmation and consequence copy.

## 10. Brand, design system, localization and accessibility

Finished brand source of truth:

- Marigold `#FF9F1C`
- Hibiscus `#FF4D6D`
- Deep Plum `#3A0E2E`
- Blush Cream `#FFF3E6`
- Palm Green `#12A67C`
- Outfit is the single product font family. The repository must obtain and
  self-host the latest stable Outfit variable font from its authoritative
  licensed source during the foundation slice.
- Supplied SVG/PNG marks; no placeholder logo

Tasks:

- Convert brand rules into platform-neutral tokens.
- Build web/MUI and React Native theme adapters.
- Establish semantic colors, type scale, spacing, radius, elevation, motion,
  haptic and sound tokens.
- Implement the Kente accent as a controlled system, not decoration everywhere.
- Create Storybook for web/admin and an Expo component gallery.
- Build light/dark/high-contrast behavior where approved.
- Externalize every member-facing/admin string.
- Seed English/Twi/Pidgin terminology registry; launch language cut line follows
  the approved phase.
- Add localization lint: no hardcoded product copy, missing keys, invalid
  variables, banned pressure copy, mixed-register violations.
- Community review and in-context screenshot review are release requirements;
  machine-only translation cannot ship.
- Meet WCAG 2.2 AA, TalkBack/VoiceOver, captions, reduced motion, 48 dp touch
  targets, visible focus and keyboard access.

## 11. API, authentication and integration standards

- OpenAPI 3.1 is version-controlled and CI-validated.
- API errors use stable machine codes and safe localized presentation mapping.
- Cursor pagination; no unbounded endpoints.
- Request correlation IDs and trace context everywhere.
- Rate limits per actor/device/IP/endpoint class.
- Authentication proposal: phone OTP for members, short-lived access tokens,
  rotated refresh sessions, device/session revocation, step-up verification.
- Admin authentication must support organization SSO and MFA before production.
- Authorization is deny-by-default and tested at route plus use-case boundaries.
- Partner webhooks are signature-verified, timestamp/replay checked, persisted,
  deduplicated and dead-lettered.
- Provider adapters include timeout, retry, circuit-breaker, idempotency and
  graceful degradation rules.

Provider abstractions:

- SMS/OTP with WhatsApp fallback.
- Resend for transactional email and admin/ops notifications.
- Identity/Ghana Card verification.
- Face/voice liveness.
- Media/object storage and CDN.
- LiveKit for fires/calls.
- Push notifications.
- WhatsApp Business.
- MoMo aggregator and later diaspora payment adapter.
- ASR/TTS, AI gateway and moderation/classification.

No provider is selected solely by SDK convenience. Security, Ghana support,
data location, retention, cost, fallback and 30-day replaceability are scored.

## 12. Render Blueprint and environments

Environments:

- Local: containerized dependencies or approved managed dev resources.
- Preview: per-pull-request where economical, synthetic data only.
- Staging: production-like, synthetic personas, integration sandboxes.
- Production: restricted access, encrypted data, audited operations.

Proposed Blueprint resources:

- Member web service.
- Admin web service.
- Go API web service.
- Go background worker.
- Optional realtime gateway if websocket separation is needed.
- Cron jobs only for coarse scheduling; durable product timers remain in the
  worker/state model.
- Health checks, pre-deploy migration job, environment groups and autoscaling
  configuration as supported by the selected Render plan.

External managed resources:

- MongoDB Atlas.
- Object storage/CDN.
- LiveKit Cloud or approved self-hosted region.
- Resend and communications providers.

Deployment gates:

- `render.yaml` validates in CI.
- Every secret is declared by name but never committed.
- `/live`, `/ready` and dependency-degraded health semantics are documented.
- Deployments are backward-compatible: expand, migrate/backfill, contract.
- Rollback does not require reversing destructive data migrations.
- Feature flags/kill switches exist for Sow, Fires, AI, Payments and Gate.
- Production backup restore is rehearsed.

Critical architecture risk:

- The documents require in-region member-content residency and low Ghana
  latency. Render region availability and each managed provider’s data
  location must be verified in Sprint 0. If Render cannot satisfy the approved
  residency/DPIA posture, the founder must approve either a documented legal
  interpretation, hybrid hosting, or a platform change before P0 production.

## 13. Testing strategy

### Go

- Standard `testing` package and table-driven tests.
- GoMock generated from application ports at controlled boundaries.
- Domain policies/state machines tested without mocks.
- Testcontainers for Go integration tests against real ephemeral MongoDB and
  any other dependency whose behavior is part of the contract.
- Testcontainer suites must wait on explicit readiness, use isolated databases,
  run migrations/index setup, assert cleanup, and never silently skip because
  a developer environment variable is absent.
- Docker/container-runtime absence must fail the explicitly requested
  integration target with an actionable message; unit-only results may not be
  reported as integration verification.
- Contract tests for provider adapters and webhooks.
- Property/fuzz tests for allowance, alternation, time windows, idempotency,
  ledger balance, age gate and Sika Shield invariants.
- Race detector and static analysis in CI.
- Mutation testing considered for critical policy packages.

### Web/admin

- Unit/component tests for logic and UI states.
- React Testing Library for accessible behavior.
- Playwright for critical journeys and role/permission matrices.
- Visual regression for brand primitives and signature screens.
- Axe accessibility checks plus manual keyboard/screen-reader passes.

### Mobile

- Unit/component tests.
- Maestro or approved equivalent for device journeys.
- Real-device tests on reference low-memory Android hardware.
- Network shaping for 3G/offline/reconnect.
- Media permission, interruption, background and low-storage cases.

### Platform and release

- API contract and generated-client drift checks.
- Golden path: register → verify → sow → sprout → room every five minutes in
  production using isolated synthetic actors.
- Load tests for fires, media, state transitions and admin queues.
- Chaos tests for duplicate/out-of-order webhooks, retries and reconnect.
- SAST, dependency, secret, container and IaC/config scans.
- DAST on staging.
- External penetration testing before public launch.
- Monthly fraud red-team scripts by P1.

Coverage policy:

- Do not use a single percentage as proof of quality.
- Critical domain/application packages require high branch coverage and mapped
  invariant tests.
- Every FR/NFR maps to at least one test case; critical laws map to multiple
  layers.

## 14. Code quality and engineering standards

Repository gates:

- Go formatting, vet/static analysis, lint, tests and race checks.
- TypeScript strict mode; no unexplained `any`.
- ESLint, formatting, type-check, unit tests and production builds.
- Dependency license and vulnerability policy.
- Generated files must be reproducible and drift-free.
- No secret, raw production data or provider credential in code/logs/fixtures.
- Structured logging with PII redaction.
- Public APIs and non-obvious invariants documented.
- Complexity and duplication monitored; exceptions are explicit.
- Architecture-boundary tests prevent inward dependency violations.
- Conventional commit style prefixed by work item once Jira is connected.

Review gates:

- At least one code review; two reviewers for security, privacy, payments,
  identity and state-machine changes.
- CODEOWNERS for critical areas.
- Consent-map row and requirement IDs included in relevant PRs.
- Threat-model delta required for new data purpose or trust boundary.
- Database/index/migration impact included for persistence changes.
- UI PRs include responsive/accessibility/low-bandwidth evidence.

Dependency policy:

- Resolve latest stable versions only at bootstrap or deliberate upgrade time.
- Pin exact application/tool versions with lockfiles/toolchain declarations.
- No prerelease/RC/beta unless separately approved.
- Automated updates may open grouped PRs; they never auto-merge major versions.
- Every new package needs ownership, license, maintenance and security review.
- Prefer standard library/framework capability over thin low-value dependencies.

## 15. Security, privacy, safety and compliance workstream

Before P0:

- Data inventory, processing register and classification.
- Ghana DPC registration and DPO ownership.
- DPIA covering identity, biometrics, voice, matching, safety and AI.
- Threat models for account takeover, identity fraud, media, state engine,
  admin, webhooks, provider compromise and insider access.
- Encryption in transit and at rest; envelope encryption/per-user key design for
  identity, biometric and voice assets.
- Secrets manager and rotation.
- Admin MFA, least privilege, break-glass and access review.
- OWASP ASVS L2 service baseline; MASVS L2 mobile baseline.
- Incident response, evidence handling and 72-hour reporting clock runbook.
- Data export and deletion with legal-hold behavior.
- Retention automation and proof-of-deletion records.
- Consent receipts and data-use enforcement.
- Care escalation designed with qualified local specialists.
- Women’s-safety review gate on doorway/family-facing features.
- Moderation worker exposure/welfare controls.

AI controls:

- Model/vendor register, purpose, data policy and version.
- AI gateway; no direct model calls from clients.
- Prompt/output safety tests and multilingual red-team set.
- Human review on liveness uncertainty.
- Matching reasons must be grounded; no generated flattery.
- Okyeame cannot initiate romance, impersonate, auto-send, claim therapy, or
  expose another member’s private data.
- Counsel content is excluded from matching.
- Fairness/exposure audits and documented model cards.
- New AI purposes require DPIA/consent review before release.

## 16. Observability, reliability and operations

Signals:

- OpenTelemetry traces, structured logs and RED/USE metrics.
- Product-law metrics and phase exit metrics.
- Queue lag, job age, webhook failures, notification deliverability.
- MongoDB latency, index usage, transaction conflicts and connection pressure.
- Media upload/transcode/playback latency.
- Verification latency and human queue age.
- Safety queue SLAs and unresolved Tier-A cases.
- Error budgets for core and fire-window availability.

Operational artifacts:

- Service catalog and owners.
- SLOs, alerts and escalation matrix.
- Runbooks for provider outage, verification backlog, media failure, fire
  degradation, duplicate payments, data breach and safety incident.
- On-call and incident severity scheme.
- Blameless post-incident reviews with actions tracked as backlog work.
- Capacity and cost dashboards.
- Backup/restore and disaster-recovery drills.

Required NFR targets:

- Core monthly availability: 99.9%.
- Fire-window target: 99.95%.
- RPO <= 5 minutes; RTO <= 60 minutes.
- Crash-free sessions >= 99.6%; ANR < 0.4%.
- Reference-device and 3G budgets from NFR-101 through NFR-106.

## 17. Delivery workflow and governance

Hierarchy:

`Project → Epic → Story/Bug/Spike → Task/Subtask`

Every story contains:

- Requirement and screen IDs.
- User story and business value.
- Acceptance criteria in observable language.
- Technical notes and data/consent impact.
- Dependencies and risks.
- Estimate.
- Test cases.
- Definition of Done.

Suggested Jira key: `OBI`.

Git conventions:

- Publish target: `main`.
- Commit: `OBI-123 concise imperative summary`.
- Each commit contains one claimed task or one tightly related review fix.
- Before editing: fetch/prune, inspect `origin/main`, read the live task ledger,
  and claim the task.
- Before publishing: fetch again, rebase the agent's commit(s) onto current
  `origin/main`, rerun affected verification, then push directly to `main`.
- After publishing: verify `HEAD...origin/main` is `0 0`, refresh
  `agent_plan.md`, and select/claim the next ready task.
- Never stage another agent's uncommitted files. If the shared worktree is
  dirty outside the claimed path, use an isolated Git worktree rooted at
  `origin/main`.
- Feature flags protect incomplete release scope even though delivery is
  direct-to-main.

Workflow:

`Backlog → Ready → In Progress → Code Review → QA → Staging → UAT → Done`

Separate release states:

`Release Candidate → Beta → Production → Hypercare → Closed`

Rituals:

- Daily async update: done, next, blocker, risk.
- Weekly backlog refinement.
- Two-week planning/review/retro.
- Weekly 30-minute Auntie Review; failing work does not advance to release.
- Architecture/security review as-needed, required at trust-boundary changes.
- Monthly dependency/reliability/cost review.
- Phase exit review based on metrics, not dates alone.

## 18. Multi-agent coordination and task ledger

`agent_plan.md` is the authoritative shared coordination surface. Every agent
must read the latest `origin/main` version before claiming work.

Allowed statuses:

- `BACKLOG`: not yet dependency-ready.
- `READY`: may be claimed.
- `IN PROGRESS`: exclusively claimed by one named agent.
- `IN REVIEW`: implementation is published; review/correction is active.
- `BLOCKED`: cannot proceed; the note names the dependency and owner.
- `DONE`: acceptance and verification are complete and published to `main`.

Atomic claim protocol:

1. Fetch `origin/main` and re-read this ledger.
2. Select one `READY` task whose dependencies are complete.
3. Change its status to `IN PROGRESS`, add the exact agent identifier,
   `claimed_at` timestamp, and owned paths.
4. Rebase that claim commit onto current `origin/main` and push it immediately.
   A task is not reserved until the claim is visible on `origin/main`.
5. If two agents race, the first claim on `origin/main` wins. The other agent
   must drop the duplicate claim, sync and choose another task.
6. Do not edit paths owned by another `IN PROGRESS` task without first recording
   a handoff or coordinated shared-path note.
7. A stale claim may be released only after the agent is unreachable or reports
   a handoff; record the reason in Notes.

Completion protocol:

1. Sync/fetch before final verification.
2. Update status to `DONE`, add verification evidence and completion timestamp.
3. Rebase only the claimed slice onto current `origin/main`.
4. Push directly to `main` and verify local/remote parity.
5. Re-read the ledger because another agent may have moved the queue.

Review protocol:

- Any agent may review a `DONE` task.
- Material corrections require a new `READY` review task or an explicit
  handoff; do not silently reopen or overwrite another active task.
- The reviewer records date, agent, findings, corrections and verification in
  the Review note column.
- Cross-review is encouraged: frontend agents review backend API ergonomics and
  states; backend agents review contract/security/invariant handling; Codex or
  ChatGPT performs the primary UI implementation and rendered UI review.

Agent allocation:

- `codex-ui`: member web, admin, mobile presentation, design tokens,
  accessibility, responsive behavior and rendered visual QA.
- `backend-*`: Go domain/application/adapters, MongoDB, workers, provider ports
  and Testcontainers integration suites.
- `platform-*`: monorepo, CI/CD, Render, observability and security automation.
- Agents may review outside their implementation lane but must claim correction
  work before editing.

### Sprint task status

This table contains every executable task promoted from the epic backlog.
Later epic stories remain `BACKLOG` in Section 20 and are promoted into this
table with a concrete sprint, owner and path boundary during refinement.

| Task | Deliverable | Sprint | Agent | Status | Owned paths | Claimed / completed | Verification | Review note |
|---|---|---:|---|---|---|---|---|---|
| S0-001 | Publish plan v1 and coordination ledger | 0 | `codex-root` | DONE | `agent_plan.md` | claimed/completed 2026-07-26 | Source consistency checked; remote parity verified after baseline push | Self-reviewed; no conflicting source rules |
| S0-002 | Sync/initialize local checkout from confirmed GitHub `main` | 0 | `codex-root` | DONE | repository metadata and source handover | completed 2026-07-26 | Empty remote confirmed; `main` initialized; parity verified after baseline push | Handover files preserved |
| S0-003 | ADR set: monorepo, Go hexagon, Mongo, REST/OpenAPI, Expo, Render | 0 | `/root/backend_architecture` | DONE | `internal/architecture/` | claimed/completed 2026-07-26 | Six accepted ADRs indexed and manually checked for source/plan consistency | Cross-reviewed by `codex-root` 2026-07-26; accepted without correction |
| S0-004 | Current stable/compatible toolchain and dependency matrix | 0 | `codex-root` | DONE | `internal/architecture/dependency-matrix.md` | claimed/completed 2026-07-26 | Registry/module resolution and engine/peer compatibility recorded; final build proof delegated to S0-005 | React pinned to Expo-compatible stable; TS/ESLint require scaffold proof |
| S0-005 | Monorepo and application/service skeletons | 0 | `/root/platform_scaffold` | DONE | root manifests, `apps/`, `services/`, `packages/` | claimed/completed 2026-07-26 | pnpm 11.17 frozen install + peer check; lint/typecheck/test/Go test and all client production builds passed | UI intentionally skeletal for S0-008; TS 5.9/ESLint 9 compatibility pins documented |
| S0-006 | CI, security and supply-chain baseline | 0 | `/root/platform_scaffold` | DONE | `.github/`, `internal/quality/ci-security-baseline.md` | claimed/completed 2026-07-26 | actionlint passed; audit gate proved and S0-015 reduced high/critical findings to zero; frozen full checks passed | SHA-pinned Actions, least-privilege permissions, Dependabot and private reporting policy; no advisories waived |
| S0-007 | Go hexagonal module template and Mongo/Testcontainers spike | 0 | `/root/backend_architecture` | DONE | `services/api/`, `services/worker/`, Go module files | claimed/completed 2026-07-26 | Go unit tests with GoMock; Mongo 8 Testcontainers integration verifies index, persistence, lookup and duplicate rejection | Self-reviewed; dependency direction and API/worker composition roots verified; no client paths touched |
| S0-008 | Outfit brand tokens and member web signature UI prototype | 0 | `codex-root` | DONE | `apps/web/`, `packages/design-tokens/`, `packages/ui-web/` | claimed/completed 2026-07-26 | lint, type-check, test and production build pass; desktop and 390px browser QA; zero runtime errors or horizontal overflow | Hydration mismatch found in review and fixed with MUI Next 16 cache provider |
| S0-009 | Admin shell signature UI prototype | 0 | `codex-root` | DONE | `apps/admin/`, shared UI primitives | claimed/completed 2026-07-26 | lint, type-check, test and production build pass; desktop and 390px mobile browser QA pass with zero runtime errors and no horizontal overflow | Operator command centre and responsive navigation verified; mobile header visibility defect found during cross-review and corrected |
| S0-010 | Expo/Android floor, media and 3G feasibility spike | 0 | `codex-root` | DONE | `apps/mobile/`, `internal/architecture/mobile-feasibility.md` | claimed/completed 2026-07-26 after S0-005 completion | Expo Doctor 20/20; lint, type-check, unit tests and web export pass; native release build completed 531 Gradle tasks with minSdk 24/targetSdk 36; 390px rendered interactions and no overflow; 97.3 MB universal APK measured | Mobile UI and interaction cross-review passed; booted API 36 install stalled in Package Manager, and API 26/2 GB physical-device plus Play split-size evidence remain explicit release gates |
| S0-011 | Render/residency/provider feasibility matrix | 0 | `/root/render_residency` | DONE | `deploy/render/provider-residency-feasibility.md` | claimed/completed 2026-07-26 | Primary-source provider matrix, four topology options, production decision gates, and follow-up evidence; Render has no African region, so production remains founder/privacy/legal-blocked | No external resource creation or deployment; self-reviewed 2026-07-26 |
| S0-012 | Initial threat model, data classification and DPIA inputs | 0 | `/root/backend_security` | DONE | `internal/architecture/threat-model-v0.md`, `internal/architecture/data-classification.md`, `internal/architecture/dpia-inputs.md` | claimed/completed 2026-07-26 | Content checked against plan §§4/7/8/11/15/23 and extracted Doc 08 (consent map, AI systems, suban) and Doc 09 (action ladder, SLAs, binding retention table, compliance pack); no code paths touched | Self-reviewed; E2E-vs-safety-processing and residency recorded as the two open pre-production decisions |
| S0-013 | Traceability matrix, synthetic personas and fixture policy | 0 | `/root/backend_traceability` | DONE | `internal/product/`, `packages/test-fixtures/` | claimed/completed 2026-07-26 | All P0 FR/NFR rows linked req→module→screen→story→consent→tests→release; FR-202 sample traced end-to-end (matrix §4); tsc strict pass on personas registry; prettier clean | Self-reviewed; FR-105 mapped to P1 productization with P0 assisted vouch noted |
| S0-014 | Sprint 0 integrated verification and founder checkpoint | 0 | `codex-root` | DONE | `agent_plan.md`, `internal/quality/sprint-0-founder-checkpoint.md` | claimed/completed 2026-07-26 after S0-003–S0-013 completion | frozen install, peers, 9/9 checks, 3/3 builds, Go vet, actionlint, Expo Doctor 20/20 and JS high/critical audit gate pass; checkpoint records the red Go security workflow and Docker integration timeout | Conditional Sprint 1 engineering proceed; production no-go and founder approval pending; active S1-002 paths were reviewed read-only |
| S0-015 | Remediate high transitive dependency advisories | 0 | `/root/platform_scaffold` | DONE | `pnpm-workspace.yaml`, `pnpm-lock.yaml` | claimed/completed 2026-07-26 | audit: 0 high/0 critical; frozen install, peers, 9/9 Turbo checks, Go tests/vet, 3/3 client builds passed | Exact patched transitive overrides; no waiver; one moderate advisory remains below the approved gate |
| S1-001 | E02-S01 Go composition root and hexagonal module wiring | 1 | `/root/backend_kernel` | DONE | `services/api/main.go`, `services/api/internal/platform/` | claimed/completed 2026-07-26 | gofmt/vet clean; go build ok; config+health unit tests pass; startup failure path exits 1 with actionable message; /live vs dependency-degraded /ready unit-verified; pre-existing S0-007 Testcontainers suite attempted but Docker Desktop saturated (58 containers, daemon calls >2 min) — environment issue recorded, not a code failure | Self-reviewed; integration rerun pending Docker recovery (asked founder); member module built but unrouted until S1-003 HTTP envelope |
| S1-002 | E02-S05 MongoDB client, transaction, repository and migration conventions | 1 | `/root/backend_kernel` | DONE | `services/api/internal/platform/mongo/`, Go module files | claimed/completed 2026-07-26 | gofmt/vet/build clean; unit tests green; Testcontainers integration green after founder-approved Docker recovery (133 stale containers removed): migrations apply once + unique index enforced, transaction commit/abort verified against single-node replica set via directConnection | Self-reviewed; replica-set + directConnection pattern documented for all future integration suites |
| S1-003 | E02-S02 API envelope, OpenAPI, validation and generated TS client | 1 | `codex-root` + `/root/api_contract` | DONE | `services/api/internal/platform/http/`, `services/api/openapi/`, `packages/api-client/` | claimed/completed 2026-07-26 after S1-001 | Redocly strict contract lint; generated-client drift check; 3/3 client tests; TypeScript typecheck/build; Go HTTP + OpenAPI contract tests; API vet clean | Completion independently verified by codex-root after S1-011 conflict mapping; stable envelopes, bounded strict JSON, safe correlation IDs, generated typed client and internal-error redaction are present. Attribution corrected 2026-07-26 by `/root/backend_kernel`: `/root/backend_http` was a raced duplicate claim dropped per protocol §18.5; implementation and first claim were codex-root's |
| S1-004 | E02-S06 Idempotency, optimistic concurrency, outbox and inbox | 1 | `/root/backend_kernel` | DONE | `services/api/internal/platform/`, `services/worker/` | claimed/completed 2026-07-26 | gofmt/vet clean; unit tests green; Testcontainers integration green: outbox commit/abort atomicity with domain change, publish flow, inbox consumer-scoped dedup under redelivery, idempotency claim/complete/replay | Self-reviewed; async outbox relay lands with worker scheduling (S1-009) |
| S1-005 | E02-S03 Authentication/session/device model | 1 | `/root/backend_identity` | DONE | `services/api/internal/identity/` | claimed/completed 2026-07-26 | gofmt/vet clean; GoMock application tests + domain policy tests pass; Testcontainers integration green: session round-trip, stale-version rejection, device/member revoke scopes; wired into composition root | Self-reviewed; opaque hashed tokens (no JWT dependency); reuse of rotated refresh revokes session; OTP endpoints land with E03-S01 |
| S1-006 | E02-S04 Authorization/RBAC/ABAC kernel | 1 | `/root/backend_authz` | DONE | `services/api/internal/authz/` | claimed/completed 2026-07-26 | gofmt/vet clean; policy matrix tests green: deny-by-default (anonymous, unknown action/resource), owner rules, FR-101 tier gates (romantic Tier 1, sowing Tier 2), least-privilege desk roles, host circle scoping | Self-reviewed; explicit grant table — new capabilities require deliberate rule additions; role assignment persistence lands with E16 |
| S1-007 | E02-S07 Structured logging, traces, metrics, health and correlation | 1 | `codex-root` + `/root/telemetry` | DONE | `services/api/internal/platform/telemetry/`, `packages/observability/` | claimed/completed 2026-07-26 after S1-001 completion | 15/15 workspace lint/type/test tasks; 5/5 builds; Go telemetry race tests; 4 client privacy tests; full Go tests/vet; govulncheck clean; JS audit 0 high/critical; Prettier and actionlint pass | Isolated-agent implementation rebased twice before integration; cross-review added self-contained package tooling and default redaction for raw errors, payloads and member identifiers; production exporter/composition binding tracked as S1-013 |
| S1-008 | E02-S08 Feature flags, configuration audit and kill switches | 1 | codex-root + /root/feature_flags | DONE | `services/api/internal/platform/flags/` | claimed/completed 2026-07-26 after S1-002 | race-tested immutable snapshots; strict environment parsing; atomic batch/rejection tests; precedence and privacy-safe audit tests; full Go tests/vet; govulncheck found 0 reachable vulnerabilities | Cross-reviewed after agent rebase; Sow/Fires/AI/Payments/Gate default disabled; runtime kill > environment kill > runtime/environment enablement; persistent admin/API composition remains E16 |
| S1-009 | E02-S09 Worker scheduling, retries, dead letters and job observability | 1 | `/root/backend_worker` | DONE | `services/worker/`, relocated shared packages `internal/platform/` (mongo, outbox, inbox, idempotency) + import updates in `services/api/` | claimed/completed 2026-07-26 | gofmt/vet clean; 18/18 Go test packages pass incl. Testcontainers: relay publishes pending records and marks them published against real MongoDB, dead-letter round trip, scheduler backoff/timeout/dead-letter unit tests, relay failure-path tests | Self-reviewed; shared packages moved to `internal/platform/` because `services/api/internal` is not importable from the worker; telemetry is codex's lane (S1-013) and intentionally not moved |
| S1-010 | E02-S10 Media abstraction, signed access and retention metadata | 1 | codex-root + /root/media_kernel | DONE | `services/api/internal/media/` | claimed/completed 2026-07-26 after S1-003 | media race tests; full Go tests/vet; signed TTL, ownership/purpose, checksum, retention/legal-hold, provider-redaction and nil-readiness tests | Cross-review corrected zero/pre-creation deletion timestamps and nil dependency panic paths; vendor persistence/HTTP composition remains a later integration lane |
| S1-011 | Review fix (S1-003 note): transport-neutral duplicate-conflict sentinel | 1 | `/root/backend_review` | DONE | `services/api/internal/member/`, `services/api/internal/platform/http/member.go`, `services/api/openapi/openapi.yaml` | claimed/completed 2026-07-26 | gofmt/vet clean; Go unit tests incl. new 409 `member_conflict` mapping pass; openapi contract test + Redocly strict lint pass; client regenerated with drift check green; prettier clean | Self-reviewed; `domain.ErrDuplicateMember` sentinel keeps mongo types out of the HTTP layer per S1-003 review note |
| S1-012 | Review fix (Sprint 0 CP-001): remediate reachable Go text vulnerability | 1 | `codex-root` | DONE | Go module files, `internal/quality/sprint-0-founder-checkpoint.md` | claimed/completed 2026-07-26 after S1-002 released Go-module ownership | `golang.org/x/text` upgraded from `v0.37.0` to fixed `v0.39.0`; full Go tests/vet pass; govulncheck clean | Closes `GO-2026-5970` without touching active S1-003 HTTP/client work; `golang.org/x/sync v0.21.0` selected by tidy as the compatible transitive floor |
| S1-013 | S1-007 follow-up: bind telemetry kernel and approved exporter in API/worker composition roots | 1 | Unassigned | BACKLOG | API/worker composition roots, HTTP correlation bridge, deployment telemetry config | Depends S1-003, S1-004 and provider/topology approval | runtime smoke proves correlated log, trace and RED/USE metric without PII | Kernel is complete; deferred to avoid active S1-003/S1-004 paths and premature vendor/exporter commitment |
| S2-001 | E01-S01 Complete platform-neutral brand tokens and asset manifest | 2 | `codex-root` | DONE | `packages/design-tokens/` | claimed/completed 2026-07-26 after E01 audit | 5/5 token tests; strict typecheck; frozen install; 19/19 workspace checks; 6/6 builds; full Go suite | Cross-platform semantic/state/type/space/elevation/motion/feedback/a11y/zone/Kente tokens; supplied-asset manifest; Outfit-only and AA contrast enforced |
| S2-002 | E01-S05 Typed EN/Twi string registry and terminology policy | 2 | `/root/i18n_foundation` | DONE | `packages/i18n/` | claimed/completed 2026-07-26 after E01 audit | 8/8 registry tests; strict typecheck; lint/build; Prettier and diff checks | Typed keys/params, parity and placeholder validation, Accept-Language resolution, English fallback, terminology metadata and human-review readiness; `tw` remains provisional and unreviewed |
| S2-003 | E01-S02 Member MUI theme and shared web primitives | 2 | `codex-root` | DONE | `packages/ui-web/` | claimed/completed 2026-07-26 after S2-001 | 3/3 theme tests; strict package/member/admin typechecks; member and admin production builds | Preference-aware light/dark/high-contrast themes; reduced motion, 48px targets, focus visibility and stable Button/Card/Status/Focus exports; Outfit preserved |
| S2-004 | E01-S03 Distinct admin theme and operator primitives | 2 | `codex-root` | DONE | `packages/ui-web/src/admin*`, `apps/admin/app/layout.tsx` | claimed/completed 2026-07-26 after S2-003 | 3/3 admin-theme tests; package/admin typechecks; admin production build; formatting | Dedicated operator provider with denser geometry, finite status colors, high contrast, reduced motion, 48px controls and visible focus; Outfit/shared tokens preserved |
| S2-005 | E01-S04 Expo theme and base interaction primitives | 2 | `/root/mobile_ui` | DONE | `packages/ui-mobile/`, `apps/mobile/app/_layout.tsx`, mobile workspace dependency + lock importer | claimed/completed 2026-07-26 after S2-001 | 4/4 theme tests; UI/mobile typechecks; mobile lint; Expo web export (822 modules); frozen install | Light/dark provider, reduced-motion contract, opt-in bounded haptics, accessible 48dp Button/Pressable/Card; root corrected relative import to workspace dependency |
| S2-006 | E01-S06 Localization, accessibility and copy-policy lint | 2 | `/root/client_quality` | DONE | `packages/config-eslint/`, `packages/quality-client/`, relevant CI configuration | claimed/completed 2026-07-26 after S2-002 | 3 ESLint policy tests, 4 client-quality tests, frozen install, 34/34 workspace checks, full Go suite and client builds | Hardcoded copy, missing names/keys, placeholder drift, pressure copy and mixed-register checks; narrow reasoned non-product escape only |
| S2-007 | E01-S08 Cross-platform loading, empty, error and offline patterns | 2 | `codex-root` | DONE | `packages/ui-web/src/states/`, `packages/ui-mobile/src/states/`, state catalogs in `packages/i18n/` | claimed/completed 2026-07-26 after S2-002, S2-003, S2-005 | 8/8 i18n, 8/8 web UI and 6/6 mobile UI tests; strict package/member/mobile typechecks; formatting | Loading, warm empty, recoverable error, offline, queued, low-bandwidth and permission-denied components with deterministic live-region/busy/action semantics |
| S2-008 | E01-S09a Web Hold, Sow, Stone and Gather prototypes | 2 | `codex-root` | DONE | `apps/web/app/design-lab/`, web gesture primitives | claimed/completed 2026-07-26 after S2-007 | 4/4 reducer tests; typecheck, lint and production build; live 1280px/390px browser QA with no overflow, 48px controls and no app console errors | Pointer Hold/Stone cancellation, explicit alternatives, Sow consequence confirmation, Gather discrete choices, reduced motion and semantic live state |
| S2-009 | E01-S09b Expo Hold, Sow, Stone and Gather prototypes | 2 | `/root/mobile_gestures` + `codex-root` | DONE | `apps/mobile/app/design-lab/`, mobile gesture primitives | claimed/completed 2026-07-26 after S2-007 | 10/10 UI-mobile tests; package/mobile typechecks; mobile lint; Expo web export (835 modules) | Agent delivered accessible gestures; root cross-review fixed stale Expo typed-route compatibility before publish; physical Android gesture smoke remains a release-device gate |
| S2-010 | E01-S07 Component galleries and visual-regression gate | 2 | `codex-root` | DONE | Storybook/config, `apps/mobile/app/gallery/`, visual-test configuration and baselines | claimed/completed 2026-07-26 after S2-003 through S2-009 completion | 34/34 workspace checks, full Go suite, Storybook 10.5.4 production build, 2/2 axe plus desktop/390 screenshot comparisons, Expo web export (836 modules) | Reviewed deterministic baselines committed separately by viewport; CI installs Chromium, rejects axe violations and pixel drift, and retains the report artifact |
| S2-011 | E03-S01 Phone registration, OTP, rate limits and session issuance | 2 | `/root/backend_identity` | DONE | `services/api/internal/identity/`, auth HTTP routes + OpenAPI | claimed/completed 2026-07-26 | 19/19 Go packages pass incl. Testcontainers end-to-end (request→simulator code→verify→persisted session, resend throttle, attempt accounting, consumed-code rejection); auth HTTP error-mapping tests; Redocly strict lint + client regen + drift check green | Self-reviewed; OTP codes hashed at rest, never logged or returned; simulator sender only until scored SMS/WhatsApp provider (S2-015 lane); contract test surface list extended for auth paths |
| S2-012 | E03-S02 Promise/terms/age consent record | 2 | `/root/consent_kernel` | DONE | `services/api/internal/consent/` | claimed/completed 2026-07-26 after S1-002 | consent race tests, full Go tests and vet; immutable version, grant/withdraw, replay and optimistic-revision cases | Deny-by-default exact-version evaluation and PII-resistant evidence; persistence adapter must later atomically enforce revision plus globally unique command IDs |
| S2-013 | E03-S06 Tier/account state machines and transition audit | 2 | `/root/backend_identity` | DONE | `services/api/internal/identity/` | claimed/completed 2026-07-26 | 21/21 Go packages pass; transition matrix unit tests (one-step promotion, reasoned demotion, same-tier/skip/out-of-range rejection, blocked accounts); GoMock audit-persistence tests; Testcontainers integration: 0→1→2 with 2 audit rows, invalid transition writes nothing, optimistic-concurrency versions | Self-reviewed; tier numerics mirror authz kernel Tier; audit records commit in the same transaction as the account update |
| S2-014 | E03-S05 Under-18 hard block and 24-hour purge proof | 2 | `/root/consent_kernel` | DONE | `services/api/internal/safeguarding/` | claimed/completed 2026-07-26 after S2-013; isolated from concurrent S2-015 | safeguarding race suite, full Go tests/vet, Testcontainers outage/retry proof deletes Mongo verification/account/session/consent/media artifacts before 24 hours | Persistent HMAC-keyed block remains without raw PII; worker scheduling, object-store purger and S2-015 composition are explicit follow-ups |
| S2-015 | E03-S03 Ghana Card provider adapter and fallback queue skeleton | 2 | `/root/backend_verification` | DONE | `services/api/internal/verification/`, verification HTTP route + OpenAPI | claimed/completed 2026-07-26 | 24/24 Go packages pass incl. Testcontainers end-to-end: match→approve→Tier 1, outage/uncertain→manual queue (never silent pass), queue listing, manual approval promotes; GoMock branch tests; Redocly lint + client regen + drift check green | Self-reviewed; simulator scripts match/mismatch/uncertain/outage until scored vendor; tier bridge only at composition root; cases hold minimal proof set, no card media |
| S2-016 | E03-S10 Export, deletion, pause and legal-hold workflows | 2 | `/root/backend_privacy` | DONE | `services/api/internal/privacy/`, privacy HTTP routes + OpenAPI | claimed/completed 2026-07-26 | 28/28 Go packages pass incl. Testcontainers end-to-end: export/deletion open with statutory due times (72h/30d), duplicate rejection, hold blocks new + open deletions, processor skips held, lift unblocks and erasure+session-revocation executes; GoMock processor/service tests; Redocly + client regen + drift green | Self-reviewed; export assembler and erasure runner ports stubbed in tests — real cross-context assembly lands with worker wiring; legal-hold paths preserved per Doc 09 |
| S2-017 | E03-S07 Core profile and privacy controls | 2 | `/root/client_quality` | DONE | `services/api/internal/profile/` | claimed/completed 2026-07-26 after S1-002 | profile race suite, full Go tests/vet, Mongo Testcontainers stale-write/replay/cross-profile-command proof | Privacy-safe display name/introduction only; per-field visibility and fail-closed consent; HTTP/authz audience composition remains a follow-up |
| S2-018 | E03-S04 Voice+face liveness orchestration and uncertainty handling | 2 | `/root/consent_kernel` | DONE | `services/api/internal/verification/liveness/` | claimed/completed 2026-07-26 after S2-014 | liveness race suite, full Go tests/vet, simulator contract and in-memory end-to-end retry/replay/manual-review tests | Exact live is sole automatic pass; temporary biometric refs removed after review; durable adapters/composition remain follow-ups |
| S2-019 | E03-S12 Shared-device/collision and known-name review cases | 2 | `/root/mobile_gestures` | DONE | `services/api/internal/identity/collision/` | claimed/completed 2026-07-26 after S2-011 | collision race suite, full Go tests/vet, Mongo Testcontainers privacy/stale-write proof | HMAC-pseudonymous signals, fail-closed shared-device review and one-way audited transitions; composition remains a follow-up |
| S2-020 | E03-S11 Admin verification queue UI prototype | 2 | `codex-root` | DONE | `apps/admin/app/verification/`, shared admin shell styling | claimed/completed 2026-07-26 after S2-015 | 3/3 interaction-model tests; typecheck/lint/build; live 1280px/390px QA with no overflow, 48px controls and no app runtime errors | Redacted proof, evidence-access warning and reasoned confirmation work; backend queue/evidence audit contract promoted separately as S2-023 |
| S2-021 | E03-S08 Voice of Introduction capture/upload/transcription consent kernel | 2 | `/root/consent_kernel` | DONE | `services/api/internal/introduction/` | claimed/completed 2026-07-26 after S1-010 and S2-012 | introduction race suite, full Go tests/vet, consent/media/transcription/revocation/retention lifecycle tests | Exact consent checked before upload and transcription; only media/transcript refs retained; durable store, media/consent binding and production transcriber remain composition follow-ups |
| S2-022 | E03-S09 Doorway Question and photo vault/veil | 2 | `/root/backend_profile` | IN PROGRESS | `services/api/internal/profile/` (doorway question + vault extension), HTTP routes + OpenAPI | claimed 2026-07-26; S2-017 profile context complete | unit + integration + contract tests pending | — |
| S2-023 | E03-S11 Admin verification queue and evidence access audit contract | 2 | `/root/mobile_gestures` | DONE | `services/api/internal/verification/admin/`, admin HTTP handler, OpenAPI and generated client | claimed/completed 2026-07-26 after S2-015 and S2-020 | admin/HTTP race tests, full Go tests/vet, Mongo Testcontainers, Redocly/client drift/typecheck/tests | Independent scopes, recent-MFA evidence gate, HMAC refs and transactional decision/audit; trusted principal/HMAC composition and tier reconciliation remain follow-ups |
| S2-024 | E03 client identity and verification onboarding flows | 2 | `codex-root` | DONE | `apps/web/app/onboarding/`, member shell hydration boundary | claimed/completed 2026-07-26 after S2-011 through S2-020 | 8/8 web interaction tests; typecheck/lint/build; full phone→OTP→consent→card→liveness browser flow; live 1280px/390px QA with no overflow, minimum 52px controls and no app runtime errors | Raw card input clears before the next stage and reducer stores only an opaque reference; manual-review and liveness-consent gates are deterministic; transport binding remains a composition follow-up |
| S3-001 | E04-S01 Fie information architecture and route map | 3 | `codex-root` | DONE | `internal/product/fie-route-map.md` | claimed/completed 2026-07-26 after E03 client foundation | route registry, tier/consent/session guard order, deep-link/offline outcome and web/mobile parity review | Canonical Fie plus four-zone routes, 48px/dp shell laws and follow-on implementation slices accepted; admin/API boundaries remain separate |
| S3-002 | E04-S02 First-run interactive walk | 3 | `codex-root` | DONE | `apps/web/app/fie/welcome/` route and client state model | claimed/completed 2026-07-26 after S3-001 | 12/12 web tests; typecheck/lint/build; completion and skip browser flows; live 1280px/390px QA with no overflow, 48px controls and no app errors | Five-zone keyboard/screen-reader walk, reduced-motion safe, explicit versioned finish/skip preference and no account penalty |
| S3-003 | E04-S03 Compound home and zone indicators | 3 | `codex-root` | DONE | `apps/web/app/fie/` home shell, state model and responsive zone navigation | claimed/completed 2026-07-26 after S3-001/S3-002 | 15/15 web tests; typecheck/lint/build; connection saver to synced browser flow; live 1280px/390px QA with zero horizontal overflow, 48px controls and no app errors | Four zones, connection/queued state and tonight's fire are visible without a feed; isolated the fire card from global CSS collision before publish |
| S3-004 | E04-S04 Abɔnten public community shell without romantic initiation | 3 | `codex-root` | DONE | `apps/web/app/fie/abonten/` route plus shared compound navigation | claimed/completed 2026-07-26 after S3-003 | 18/18 web tests; typecheck/lint/build; filter/save browser flow; live 1280px/390px QA with zero overflow, 48px controls and no app errors | Community fires, learning and notices only; prohibited romantic action vocabulary is regression-tested and no romantic controls are rendered |
| S3-005 | E04-S05 Adiwo circle courtyard shell | 3 | `codex-root` | DONE | `apps/web/app/fie/adiwo/` route and membership-safe interaction model | claimed/completed 2026-07-26 after S3-004 and S3-010/S3-012 | 21/21 web tests; typecheck/lint/build; discovery/request review browser flow; live 1280px/390px QA with zero overflow, 48px controls and no app errors | Invite-only circles stay non-requestable; request review discloses only display name/request and explicitly excludes other memberships |
| S3-006 | E04-S06 Ɛpono ano deliberate introduction shell | 3 | `codex-root` | IN PROGRESS | new `apps/web/app/fie/epono-ano/` route and review interaction model using shared compound navigation | claimed 2026-07-26 after S3-005 and E03 introduction/doorway foundations | tier/consent/review-state tests, typecheck/lint/build and desktop/390 rendered QA pending | Tier 1 kind gate, bounded introductions and explicit accept/pass review without swipe mechanics; rebase before every commit |
| S3-010 | E05-S01 Circle types, privacy defaults and membership kernel | 3 | `/root/client_quality` | DONE | `services/api/internal/circle/` | claimed/completed 2026-07-26 after E03 profile/auth foundation | circle race/property suite, full Go tests/vet, Mongo Testcontainers concurrent-write/privacy/replay proof | Finite taxonomy, private default, one-way audited membership and deny-by-default capabilities; owner transfer/demotion/rejoin require future explicit policy |
| S3-011 | E05-S02 Host application and institutional verification kernel | 3 | `/root/consent_kernel` | DONE | `services/api/internal/host/` bounded context, privacy keyer and Mongo adapter | claimed/completed 2026-07-26 after S3-010 | race, GoMock and lifecycle tests; Mongo Testcontainers manual-review/idempotency/raw-identifier proof; full Go tests and vet | Opaque evidence refs, fail-closed uncertainty, bounded approval/recheck/expiry and privacy-keyed audit; production provider and worker composition remain follow-ups |
| S3-012 | E05-S03 Circle invites, request/approval and expulsion workflows | 3 | `/root/mobile_gestures` | DONE | `services/api/internal/circle/workflow/` bounded workflow package and Mongo/token adapters | claimed/completed 2026-07-26 after S3-010 | race, GoMock and 300-case property tests; Mongo Testcontainers concurrent redemption/expulsion/raw-token proof; full Go tests and vet | 256-bit opaque single-use invites, authorization port, replay fingerprints, optimistic revisions and append-only audit; kernel composition remains a follow-up |
| S3-014 | E05-S04 Voice-first circle room, events and noticeboard kernel | 3 | `/root/consent_kernel` | IN PROGRESS | isolated `services/api/internal/circle/room/` bounded context without editing circle membership/workflow files | claimed 2026-07-26 after S3-010 through S3-012 | voice-first post, event, notice, authorization, retention and Mongo Testcontainers tests pending | Membership-scoped reads/writes, voice-first content refs only, no media bytes/free-text analytics, replay-safe mutations; rebase before every commit and hand off without pushing |
| S3-013 | E05-S05 Trust edge model and bounded trust-path projection | 3 | `/root/client_quality` | DONE | `services/api/internal/trust/` bounded context and Mongo adapter | claimed/completed 2026-07-26 after S3-010 | race, property and fuzz tests; Mongo Testcontainers bounded-read/revoke/provenance/replay proof; full Go tests and vet | Directed immutable provenance, owner authorization before reads, per-edge consent and visibility, cycle-safe depth 4/node 100 bounds, no global or reverse graph browser |
| S3-015 | E05-S06 Member trust-path visibility and privacy-safe explanations | 3 | `/root/client_quality` | DONE | `services/api/internal/trust/visibility/` revalidation projection and isolated HTTP contract | claimed/completed 2026-07-26 after S3-013 | race, GoMock, HTTP and 200-order property tests; Mongo Testcontainers consent-withdrawal revalidation; full Go tests and vet | Revalidates edge identity, both endpoints, consent, visibility, expiry and revocation immediately before disclosure; generic reason codes only and privacy-safe 404; shared route/OpenAPI registration remains composition work |
| S3-016 | E05-S07 P0 assisted vouch workflow | 3 | `/root/mobile_gestures` | IN PROGRESS | isolated `services/api/internal/vouch/assisted/` bounded context without editing trust/circle packages | claimed 2026-07-26 after S3-013 | assisted request, operator decision, authorization, replay, audit and Mongo Testcontainers tests pending | Manual-assisted only, explicit voucher consent, immutable outcome provenance, no stake/payment/productized behavior; rebase before every commit and hand off without pushing |

## 19. Estimation and capacity rules

- Story points use Fibonacci 1, 2, 3, 5, 8, 13.
- A 13-point item must be split or treated as a time-boxed epic/spike.
- Estimates include implementation, tests, docs, review and rollout evidence.
- Reserve 15% sprint capacity for defects, security, reliability and discovery.
- Do not count external provider/legal waiting time as engineering effort; track
  it as a dependency with owner and due date.
- Sprint commitments begin only when acceptance criteria and dependencies meet
  Definition of Ready.

Definition of Ready:

- Outcome and actor are clear.
- Requirement/screen IDs are linked.
- Acceptance criteria are testable.
- UX/content dependency is available or explicitly part of the story.
- Data classification/consent row is identified.
- External dependency has owner/fallback.
- Estimate is <= 8 points.

Definition of Done:

- Acceptance criteria met.
- Unit/integration/e2e tests appropriate to risk pass.
- Requirement/test traceability updated.
- Security/privacy/accessibility/localization checks pass.
- Observability and operational behavior included.
- Documentation and decision records updated.
- Reviewed and merged.
- Deployed to required environment behind the correct flag.
- QA/UAT evidence attached where required.
- Jira and dashboard status synchronized once integrations exist.

## 20. Epic backlog

The IDs below are stable local planning IDs until Jira keys are assigned.

### E00 — Inception, governance and monorepo foundation

Outcome: a reproducible, governed workspace ready for product delivery.

Stories:

- E00-S01 Approve architecture and scope decisions.
  - Record Go, MongoDB, Next.js/MUI, Expo, Render, Resend and web-scope ADRs.
  - Close residency/hosting posture or record blocking decision.
  - Approve API style and managed-provider evaluation criteria.
- E00-S02 Initialize repository governance.
  - Git repository, protected default branch, issue/PR templates, CODEOWNERS.
  - `AGENTS.md`, `CLAUDE.md`, README, contribution and security policies.
- E00-S03 Scaffold monorepo.
  - pnpm/Turborepo and Go workspace.
  - Web, admin, mobile, API and worker buildable skeletons.
  - Shared configs and deterministic local commands.
- E00-S04 Establish CI and supply-chain baseline.
  - Lint/type/test/build/security workflows and caches.
  - Dependency pinning, update policy, SBOM and license scan.
- E00-S05 Establish delivery traceability.
  - Requirement-screen-test-release matrix.
  - Jira/GitHub/dashboard mapping format without premature external mutation.
- E00-S06 Developer environment.
  - Example environment files, secret names, local dependency strategy.
  - Synthetic personas, fixtures and onboarding verification.

Exit: clean checkout passes all baseline commands; no product feature is claimed.

### E01 — Brand system, UX foundation and localization

Requirements: M2, M11, NFR-500; screens S-01–S-90.

Stories:

- E01-S01 Encode authoritative brand tokens and fonts/assets.
- E01-S02 Build MUI member theme and shared web primitives.
- E01-S03 Build distinct admin theme on shared tokens.
- E01-S04 Build Expo theme and signature interaction primitives.
- E01-S05 Implement typed string registry and EN/Twi foundations.
- E01-S06 Implement accessibility and banned-copy lint rules.
- E01-S07 Build component documentation/gallery and visual regression.
- E01-S08 Define loading/empty/error/offline patterns.
- E01-S09 Prototype Hold, Sow, Stone, Gather and keyboard alternatives.

Exit: approved signature primitives and automated brand/accessibility baselines.

### E02 — Platform kernel, API contracts and observability

Requirements: NFR-200, NFR-300, NFR-600.

Stories:

- E02-S01 Go composition root and hexagonal module template.
- E02-S02 API envelope, OpenAPI, validation and generated TS client.
- E02-S03 Authentication/session/device model.
- E02-S04 Authorization/RBAC/ABAC kernel.
- E02-S05 MongoDB client, transaction, repository and migration conventions.
- E02-S06 Idempotency, optimistic concurrency, outbox and inbox.
- E02-S07 Structured logging, traces, metrics, health and correlation.
- E02-S08 Feature flags, configuration audit and kill switches.
- E02-S09 Worker scheduling, retries, dead letters and job observability.
- E02-S10 Media abstraction, signed access and retention metadata.

Exit: a synthetic write/read/event path works end-to-end in all clients.

### E03 — Identity, account, verification and consent

Requirements: M1-01–M1-10, FR-101–FR-106; screens S-01–S-09.

Stories:

- E03-S01 Phone registration, OTP, rate limits and session issuance.
- E03-S02 Promise/terms/age consent record.
- E03-S03 Ghana Card capture/provider adapter and fallback queue.
- E03-S04 Voice+face liveness orchestration and uncertainty handling.
- E03-S05 Under-18 hard block and 24-hour purge proof.
- E03-S06 Tier/account state machines and transition audit.
- E03-S07 Core profile and privacy controls.
- E03-S08 Voice of Introduction capture/upload/transcription consent.
- E03-S09 Doorway Question and photo vault/veil.
- E03-S10 Export, deletion, pause and legal-hold workflows.
- E03-S11 Admin verification queue and evidence access audit.
- E03-S12 Shared-device/collision and known-name review cases.

Exit: no Tier-0 romantic access; verification timing and purge tests pass.

### E04 — Fie shell, navigation and member home

Requirements: M2-01–M2-07; screens S-10–S-15.

Stories:

- E04-S01 Fie information architecture and route map.
- E04-S02 First-run interactive walk.
- E04-S03 Compound home and zone indicators.
- E04-S04 Abɔnten surface without romantic initiation.
- E04-S05 Adiwo shell.
- E04-S06 Ɛpono ano shell.
- E04-S07 Dan mu shell.
- E04-S08 Okyeame presence point placeholder/capability boundary.
- E04-S09 Mobile/web navigation accessibility and deep-link guards.

Exit: usability test finds House Front, circle and tonight’s fire within 60 s.

### E05 — Circles, hosts, vouches and trust paths

Requirements: M3, M1-04, P0 manual/P1 productized vouching.

Stories:

- E05-S01 Circle types, privacy defaults and membership.
- E05-S02 Host application and institutional verification.
- E05-S03 Circle invites, request/approval and expulsion.
- E05-S04 Voice-first room, events and noticeboard.
- E05-S05 Trust edge model and bounded trust-path projection.
- E05-S06 Member-controlled trust-path visibility.
- E05-S07 P0 assisted vouch workflow.
- E05-S08 P1 immutable vouch attestation and stake behavior.
- E05-S09 Admin circle legitimacy and vouch audit.

Exit: privacy-scoped paths are derived, immutable in provenance and explainable.

### E06 — Seed economy, pods and doorway

Requirements: M4-01–M4-09, FR-201–FR-206; screens S-20–S-33.

Stories:

- E06-S01 Immutable weekly allowance ledger and renewal.
- E06-S02 Introduction-source abstraction without global browsing.
- E06-S03 Server-verified 20-second listening eligibility.
- E06-S04 Sow composer, media screening and deliberate send.
- E06-S05 Pod delivery, cap and privacy-safe playback.
- E06-S06 Decline, 90-day exclusion and kind notification.
- E06-S07 Sprout and three-exchange alternating doorway.
- E06-S08 Mutual water and single room creation under races.
- E06-S09 Garden states, expiries and dawn summary.
- E06-S10 Ember issuance/redemption.
- E06-S11 Abuse controls, demand throttling and care review signals.
- E06-S12 Property/fuzz/chaos suite for all hard invariants.

Exit: seed laws are impossible to bypass by direct API, retries or concurrency.

### E07 — Courtship rooms and guided arc

Requirements: M5-01–M5-09, FR-301–FR-304; screens S-40–S-46.

Stories:

- E07-S01 Room event/state model and projection.
- E07-S02 Strict drum alternation and voice-first stage opening.
- E07-S03 Offline queue/retry/multi-device idempotency.
- E07-S04 Response windows, resting, re-light and archival.
- E07-S05 Pause stone.
- E07-S06 Honesty ribbon.
- E07-S07 Guided theme one and simultaneous reveal.
- E07-S08 Call/meeting/exclusivity proposal objects.
- E07-S09 Kind closure and ghost-pattern behavior.
- E07-S10 Safety sheet, block/report and watermark.
- E07-S11 Themes 2–4 for P1.

Exit: consecutive send is impossible and reconciliation never duplicates events.

### E08 — Cloth and ceremonies

Requirements: M7 and later M6; screens S-50–S-54.

Stories:

- E08-S01 Deterministic cloth grammar and render seed.
- E08-S02 P0 first-thread/theme-one band.
- E08-S03 Pair-owned archival/export/deletion.
- E08-S04 Abusua Gate consent configuration.
- E08-S05 Reviewer links, OTP, watermark and expiry.
- E08-S06 Per-question consent relay.
- E08-S07 Gate ceremony and optional circle announcement.
- E08-S08 Harvest/weaver specification and order handoff.

Exit: cloth is deterministic; Gate never exposes unconsented material.

### E09 — Fires, calls and realtime

Requirements: M9, FR-401–FR-403, NFR-104; screens S-60–S-65.

Stories:

- E09-S01 Fire scheduling, capacity, RSVP and waitlist.
- E09-S02 LiveKit room/token adapter.
- E09-S03 Host stage/mute/eject/co-host controls.
- E09-S04 Low-bandwidth degradation ladder and captions.
- E09-S05 Runsheet/timer/game-segment mount points.
- E09-S06 Fire consent/recording policy.
- E09-S07 Ember-close transaction.
- E09-S08 Incident hotkey and T&S live routing.
- E09-S09 In-app call without phone-number exposure.
- E09-S10 Load/device/3G acceptance testing.

Exit: 150-seat acceptance target and safety controls pass before scale rollout.

### E10 — Games of character

Requirements: M8, FR-501–FR-502; screens S-70–S-73.

Stories:

- E10-S01 Oware domain rules and legality engine.
- E10-S02 Async game/move timers and room embedding.
- E10-S03 Glicko-2 ratings and notation.
- E10-S04 Conduct-only suban integration.
- E10-S05 Ɛbɛ duels and reviewed proverb content.
- E10-S06 Anansesɛm relay and publish consent.
- E10-S07 Ampe realtime pulse.
- E10-S08 Tournaments, ladders and anti-cheat.

Exit: game skill cannot influence matching visibility.

### E11 — Matching, Okyeame and Sentinel

Requirements: Doc 08 AI systems and consent map.

Stories:

- E11-S01 AI gateway, vendor policy and audit metadata.
- E11-S02 Rules+trust-path cold-start introductions.
- E11-S03 Grounded resonance explanations.
- E11-S04 Consent-controlled matching features.
- E11-S05 Model/ranker readiness, offline evaluation and fairness gates.
- E11-S06 Okyeame whitelist, persona, disclosure and refusal controls.
- E11-S07 Counsel isolation from matching.
- E11-S08 Liveness adapter/human uncertainty route.
- E11-S09 Sow pre-delivery multilingual screening.
- E11-S10 Sika Shield text/voice patterns and precision gate.
- E11-S11 Scam-arc sequence signals and action ladder.
- E11-S12 Syndicate/vouch-ring/device anomaly detection.
- E11-S13 AI red-team, model cards and appeal metrics.

Exit: no autonomous romance or silent uncertain liveness; consent boundaries pass.

### E12 — Trust, safety, care and compliance admin

Requirements: M13-01, Doc 09.

Stories:

- E12-S01 Universal report/block sheet.
- E12-S02 Tiered queues and SLA routing.
- E12-S03 Privacy-redacted evidence viewer and legal holds.
- E12-S04 Action ladder and propagated account/device controls.
- E12-S05 Care queue with approved resource-first scripts.
- E12-S06 Mpanyimfo docket, recusal, ruling and appeals.
- E12-S07 Women’s-safety review evidence.
- E12-S08 Incident response and regulatory runbooks.
- E12-S09 Retention, erasure and transparency-report data.
- E12-S10 Moderation workforce safeguards.
- E12-S11 Fraud victim evidence-export path.

Exit: Tier-A and care paths meet SLA/review requirements in rehearsals.

### E13 — Notifications, companions and communications

Requirements: M11, M12, FR-701.

Stories:

- E13-S01 Notification preference, quiet hours and six/day cap.
- E13-S02 Dawn, Monday, fire and Sunday rituals.
- E13-S03 Push, in-app and SMS routing.
- E13-S04 Resend domain/authentication/templates/webhooks.
- E13-S05 WhatsApp OTP and pod alerts.
- E13-S06 P1 Nnoboa/auntie flows.
- E13-S07 P2 Gate links and USSD.
- E13-S08 Delivery observability, provider fallback and opt-out.
- E13-S09 Banned engagement-pattern test suite.

Exit: caps and preferences are server-enforced across all channels.

### E14 — Commerce, payments and marketplace

Requirements: M10, FR-601–FR-603; P1+ only.

Stories:

- E14-S01 Product/SKU rules proving seeds/visibility are unsellable.
- E14-S02 MoMo provider adapter and USSD-push flow.
- E14-S03 Immutable double-entry ledger in MongoDB.
- E14-S04 Idempotent webhooks and daily reconciliation.
- E14-S05 Membership/pass grace, cancellation, receipt and refund.
- E14-S06 Matchmaker profiles/licensing/booking.
- E14-S07 Escrow, settlement, payout statement and dispute.
- E14-S08 Finance admin, exports and four-eyes pricing.
- E14-S09 Diaspora payment isolation for P2.
- E14-S10 Security/reconciliation/chaos test suite.

Exit: ledger balances, provider statements reconcile and no member transfer exists.

### E15 — Analytics, suban and product measurement

Requirements: Doc 08 event taxonomy, NFR-402.

Stories:

- E15-S01 Producer-enforced analytics schema registry.
- E15-S02 Consent-aware event pipeline without content/free text.
- E15-S03 P0 funnel and phase-exit dashboard.
- E15-S04 Append-only suban event ledger.
- E15-S05 Recomputable marks, decay and anti-gaming.
- E15-S06 Member explanation/appeal view.
- E15-S07 Fairness, regret and safety dashboards.
- E15-S08 Retention/pseudonymization/aggregation jobs.

Exit: P0 gates can be measured truthfully without privacy violations.

### E16 — Admin operations and configuration

Requirements: M13-02–M13-05.

Stories:

- E16-S01 Admin shell, roles, MFA/SSO and audited access.
- E16-S02 Verification operations.
- E16-S03 Community/circle/host/fire operations.
- E16-S04 T&S/care/panel operations.
- E16-S05 Finance/reconciliation operations.
- E16-S06 Market-pack and terminology governance.
- E16-S07 Feature flags and kill switches.
- E16-S08 Four-eyes pricing/policy configuration.
- E16-S09 Operational dashboard and case SLA reporting.

Exit: least-privilege role matrix and four-eyes actions pass adversarial tests.

### E17 — Infrastructure, release and reliability

Requirements: NFR-200/300/600, Render Blueprint.

Stories:

- E17-S01 Render Blueprint and environment matrix.
- E17-S02 MongoDB Atlas/network/security/backup configuration.
- E17-S03 Secret management and rotation.
- E17-S04 CI/CD, preview/staging/prod promotions.
- E17-S05 Mobile EAS build/release channels.
- E17-S06 SLOs, alerts, dashboards and on-call.
- E17-S07 Backup restore and disaster-recovery rehearsal.
- E17-S08 Performance, load and cost tests.
- E17-S09 Security scans, DAST and penetration-test closure.
- E17-S10 Release notes, UAT, rollback and hypercare.

Exit: production-readiness review passes with restore and rollback evidence.

### E18 — Launch operations enablement

Requirements: Doc 11; delivery is cross-functional, not software-only.

Stories:

- E18-S01 First Hundred Families CRM/ops workflow.
- E18-S02 Host School materials/certification tracking.
- E18-S03 Matchmaker Agyina licensing workflow.
- E18-S04 Verification/support staffing dashboards and scripts.
- E18-S05 Launch calendar, density gates and waitlist throttle.
- E18-S06 Campus ambassador quality-gated attribution.
- E18-S07 UAT cohort, training and feedback triage.
- E18-S08 Hypercare command center and daily launch review.

Exit: P0 does not open without populated circles, trained hosts and support cover.

## 21. Sprint roadmap

Dates begin only after plan approval and goal creation.

### Sprint 0 — Inception and risk retirement

- E00-S01 through E00-S06.
- Confirm current stable toolchain/package versions and pin them.
- Provider/residency spike: Render, MongoDB Atlas, storage, LiveKit.
- Identity/liveness/SMS/WhatsApp provider shortlist.
- Mongo transaction/index/state-machine spike.
- Mobile Android 8 compatibility and 3G media spike.
- Initial threat model, DPIA inputs, data classification.
- Brand tokens and one signature cross-platform prototype.

Gate: founder/engineering approval of ADRs, runnable baseline and dependency map.

### Sprint 1 — Platform walking skeleton

- E01 token/theme/localization foundations.
- E02 API/client generation, Mongo, auth skeleton, observability.
- Member web/admin/mobile shells and CI.
- Synthetic end-to-end authenticated read.

### Sprint 2 — Registration and verification foundation

- OTP/session/security.
- Promise, consent and age gate.
- Verification adapter and admin fallback queue skeleton.
- S-01–S-05 across mobile/web.

### Sprint 3 — Profile, voice and Fie

- Voice of Introduction/media pipeline.
- Doorway Question, photos/veil.
- Fie navigation and first-run.
- Profile/export/deletion foundations.

### Sprint 4 — Circles and trust paths

- Circle membership/host basics.
- Trust path projection.
- P0 manual-assisted vouching.
- Adiwo surfaces and admin verification.

### Sprint 5 — Seeds: allowance to sow

- Allowance ledger.
- Intro source/rules baseline.
- Playback eligibility.
- Sow recording/screening/delivery.

### Sprint 6 — Pods, sprout and doorway

- House Front, hold-to-listen, decline.
- Three-exchange doorway.
- Mutual water and race-safe room creation.
- Garden/expiry jobs and invariant tests.

### Sprint 7 — Courtship room core

- Room projection and strict alternation.
- Offline queue/reconciliation.
- Windows, rest/re-light, pause and honesty ribbon.

### Sprint 8 — Guided relationship and safety

- Theme one.
- Kind closure.
- Report/block/safety sheet.
- Cloth v0.
- T&S queue/action foundations.

### Sprint 9 — Fires, embers and Oware

- LiveKit integration and host controls.
- Weekly fire flow and degradation.
- Ember transaction.
- In-room Oware.

### Sprint 10 — P0 hardening and operational readiness

- Notification rituals/caps, WhatsApp OTP/pod alerts, Resend.
- Sentinel v0 and Sika Shield rule set.
- Admin verification/T&S completion.
- Analytics/exit-metric dashboards.
- Load, device, accessibility, security and retention testing.

### Sprint 11 — P0 UAT and closed-cohort release

- Full golden-path UAT.
- Auntie Review remediation.
- Backup restore, incident and safety rehearsals.
- Staged rollout, cohort support, hypercare.
- P0 exit metrics begin; P1 is not unlocked until gates pass.

### Sprints 12–22 — P1 foundation

- Productized vouching/suban.
- Nnoboa/auntie flows.
- Games/themes/fires expansion.
- Matchmaker/commerce/finance.
- Pidgin/Ga packs as approved.
- Mpanyimfo and fraud/safety maturity.
- Matching model only after consent, data and fairness readiness.

### Sprints 23–40 — P2 launch

- Gate, diaspora, ceremony/registry/harvest.
- Tournaments, USSD, additional languages.
- Transparency report, Ghana rollout, production maturity.

Sprint numbering beyond P0 is forecast-only and must be re-planned from measured
P0 velocity, provider outcomes and phase-exit metrics.

## 22. P0 traceability and release gates

Functional gates:

- Identity: FR-101–FR-106.
- Seeds: FR-201–FR-206.
- Rooms: FR-301–FR-304.
- Fires: FR-401–FR-403.
- Oware: FR-501–FR-502.
- Companions: relevant FR-701 subset.
- Admin: FR-801.

Metric gates:

- Pods heard >= 65%.
- Seed to sprout >= 25%.
- Sprout to room >= 35%.
- Weekly fire attendance >= 40% of cohort.
- Day-30 retention >= 45%.
- Safety regret trending down.
- Zero unresolved Tier-A incidents.

Quality/release gates:

- No critical/high unaccepted security finding.
- No open data-loss, duplicate-send, allowance, age-gate, payment-path, consent
  or authorization defect.
- All P0 FRs have test evidence and screen mapping.
- Accessibility critical journeys pass.
- 3G/reference-device budgets pass or carry explicit approved exceptions.
- Backup restore and rollback rehearsed.
- DPO/security/women’s-safety/engineering/product approvals recorded.
- Support, verification and T&S staffing ready.
- Two campus circles, one professional circle and trained fire hosts ready.

## 23. Risks and mitigations

| Risk | Severity | Mitigation / decision gate |
|---|---:|---|
| Render/provider data residency may not meet Ghana/in-region policy | Critical | Verify in Sprint 0; DPIA/legal decision; hybrid/platform change before production |
| MongoDB transactions/graph traversal may complicate invariant-heavy domains | High | State-machine/ledger spike, strict indexes, transaction tests, materialized trust paths |
| Expo support for Android 8/native identity/liveness requirements | High | Sprint 0 device/provider PoC; development build/config plugin or approved min-version decision |
| P0 scope expands due to three clients | High | Shared contracts/tokens, mobile reference, identical domain API, strict P0 cut line |
| Identity/MoMo/WhatsApp provider procurement blocks delivery | High | Port-first adapters, simulators, dual shortlist, explicit fallback/manual ops |
| E2E encryption conflicts with consented safety processing | Critical | Dedicated architecture/legal threat-model decision before room implementation |
| Voice/media costs and 3G performance miss budgets | High | Opus/resumable/progressive PoC, lifecycle policies, device/network benchmarks |
| AI safety/quality insufficient in Twi/Pidgin | High | Rules/human fallback, local paid annotation/review, threshold gates |
| Cold-start matching lacks data | Medium | Rules, circles, vouches, fires, Nnoboa; no premature model |
| Admin evidence access creates insider/privacy risk | Critical | Redaction, purpose scopes, just-in-time access, MFA, immutable audit and reviews |
| Operational community is not ready when software is | High | E18 runs from Sprint 0 with release gate |
| “Latest version” upgrades destabilize delivery | Medium | Pin exact stable versions, controlled upgrade PRs, no prerelease by default |

## 24. Decision register and founder questions

These answers improve Sprint 0 but do not prevent review of this plan:

1. **Member web P0 depth:** should member web have full P0 feature parity,
   including recording, Sow, rooms and fires, or should it be responsive
   read/limited-action access while mobile remains the only complete P0 client?
   Current plan assumes full parity except device-only identity capabilities,
   which use a secure mobile handoff.
2. **Public marketing site:** should it live inside `apps/web` as public routes
   or be a fourth independently deployed application? Current plan assumes
   public routes in `apps/web`.
3. **API style:** approve REST/OpenAPI as proposed, or retain GraphQL from the
   original technical architecture? Current plan assumes REST/OpenAPI.
4. **Project ownership:** GitHub repository is confirmed as
   `git@github.com:stanleyHayes/obiara.git`. Jira project, Render team, MongoDB
   Atlas organization and environment owners are not yet named.
5. **Provider commercial choices:** Ghana Card verification, liveness,
   SMS/WhatsApp, object storage, LiveKit hosting, MoMo aggregation, push and AI
   vendors remain selection tasks.
6. **Residency interpretation:** must all member content physically reside in
   Ghana, or is an approved African region acceptable? Legal/DPO confirmation is
   required before production infrastructure is selected.
7. **P0 language:** release plan says English+Twi; PRD mentions Pidgin at
   arrival. Current plan treats Pidgin as P1 unless founder moves it into P0.
8. **Age/device floor:** confirm whether Android 8 remains binding even if the
   latest stable Expo SDK cannot support it safely.
9. **AI timing:** current plan uses deterministic/rules-first matching and
   human-backed safety at P0, with model-assisted matching gated later. Confirm.
10. **Delivery team/capacity:** number and disciplines of available people are
    needed before sprint point commitments or calendar dates can be credible.

## 25. Approval and execution checkpoint

Confirmed:

- [x] Scope override: member web is included.
- [x] Go modular monolith with hexagonal modules.
- [x] MongoDB as initial system of record.
- [x] Expo mobile, Next.js/MUI web and admin, Render, Resend.
- [x] Outfit supersedes the document typography.
- [x] Testcontainers is the integration-testing standard.
- [x] Add a graph store only when measured need justifies it.
- [x] Direct verified delivery to `main` after rebase/sync.
- [x] Multi-agent claims, status ledger, cross-review and review notes.
- [x] Sprint 0 is the first active execution milestone.
- [x] Bounded Sprint 0 execution goal created on 2026-07-26.
- [x] Sprint 0 integrated engineering checkpoint recorded on 2026-07-26.
- [ ] Founder approval of the conditional Sprint 1 proceed decision.

Open decisions remain in Section 24 and the checkpoint corrective-action table.
They must be closed by the task that first depends on them. Sprint 1 engineering
may continue on independent stories; production provisioning and real-member
data remain prohibited.
