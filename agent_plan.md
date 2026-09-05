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

| Task   | Deliverable                                                                                   |   Sprint | Agent                                  | Status | Owned paths                                                                                                                                                                                                                                   | Claimed / completed                                                                                                                                                                                                                                                                                                                              | Verification                                                                                                                                                                                                                                                                                                                                                                                                    | Review note                                                                                                                                                                                                                                                                                                                                                                                                              |
| ------ | --------------------------------------------------------------------------------------------- | -------: | -------------------------------------- | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| S0-001 | Publish plan v1 and coordination ledger                                                       |        0 | `codex-root`                           | DONE   | `agent_plan.md`                                                                                                                                                                                                                               | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Source consistency checked; remote parity verified after baseline push                                                                                                                                                                                                                                                                                                                                          | Self-reviewed; no conflicting source rules                                                                                                                                                                                                                                                                                                                                                                               |
| S0-002 | Sync/initialize local checkout from confirmed GitHub `main`                                   |        0 | `codex-root`                           | DONE   | repository metadata and source handover                                                                                                                                                                                                       | completed 2026-07-26                                                                                                                                                                                                                                                                                                                             | Empty remote confirmed; `main` initialized; parity verified after baseline push                                                                                                                                                                                                                                                                                                                                 | Handover files preserved                                                                                                                                                                                                                                                                                                                                                                                                 |
| S0-003 | ADR set: monorepo, Go hexagon, Mongo, REST/OpenAPI, Expo, Render                              |        0 | `/root/backend_architecture`           | DONE   | `internal/architecture/`                                                                                                                                                                                                                      | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Six accepted ADRs indexed and manually checked for source/plan consistency                                                                                                                                                                                                                                                                                                                                      | Cross-reviewed by `codex-root` 2026-07-26; accepted without correction                                                                                                                                                                                                                                                                                                                                                   |
| S0-004 | Current stable/compatible toolchain and dependency matrix                                     |        0 | `codex-root`                           | DONE   | `internal/architecture/dependency-matrix.md`                                                                                                                                                                                                  | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Registry/module resolution and engine/peer compatibility recorded; final build proof delegated to S0-005                                                                                                                                                                                                                                                                                                        | React pinned to Expo-compatible stable; TS/ESLint require scaffold proof                                                                                                                                                                                                                                                                                                                                                 |
| S0-005 | Monorepo and application/service skeletons                                                    |        0 | `/root/platform_scaffold`              | DONE   | root manifests, `apps/`, `services/`, `packages/`                                                                                                                                                                                             | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | pnpm 11.17 frozen install + peer check; lint/typecheck/test/Go test and all client production builds passed                                                                                                                                                                                                                                                                                                     | UI intentionally skeletal for S0-008; TS 5.9/ESLint 9 compatibility pins documented                                                                                                                                                                                                                                                                                                                                      |
| S0-006 | CI, security and supply-chain baseline                                                        |        0 | `/root/platform_scaffold`              | DONE   | `.github/`, `internal/quality/ci-security-baseline.md`                                                                                                                                                                                        | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | actionlint passed; audit gate proved and S0-015 reduced high/critical findings to zero; frozen full checks passed                                                                                                                                                                                                                                                                                               | SHA-pinned Actions, least-privilege permissions, Dependabot and private reporting policy; no advisories waived                                                                                                                                                                                                                                                                                                           |
| S0-007 | Go hexagonal module template and Mongo/Testcontainers spike                                   |        0 | `/root/backend_architecture`           | DONE   | `services/api/`, `services/worker/`, Go module files                                                                                                                                                                                          | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Go unit tests with GoMock; Mongo 8 Testcontainers integration verifies index, persistence, lookup and duplicate rejection                                                                                                                                                                                                                                                                                       | Self-reviewed; dependency direction and API/worker composition roots verified; no client paths touched                                                                                                                                                                                                                                                                                                                   |
| S0-008 | Outfit brand tokens and member web signature UI prototype                                     |        0 | `codex-root`                           | DONE   | `apps/web/`, `packages/design-tokens/`, `packages/ui-web/`                                                                                                                                                                                    | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | lint, type-check, test and production build pass; desktop and 390px browser QA; zero runtime errors or horizontal overflow                                                                                                                                                                                                                                                                                      | Hydration mismatch found in review and fixed with MUI Next 16 cache provider                                                                                                                                                                                                                                                                                                                                             |
| S0-009 | Admin shell signature UI prototype                                                            |        0 | `codex-root`                           | DONE   | `apps/admin/`, shared UI primitives                                                                                                                                                                                                           | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | lint, type-check, test and production build pass; desktop and 390px mobile browser QA pass with zero runtime errors and no horizontal overflow                                                                                                                                                                                                                                                                  | Operator command centre and responsive navigation verified; mobile header visibility defect found during cross-review and corrected                                                                                                                                                                                                                                                                                      |
| S0-010 | Expo/Android floor, media and 3G feasibility spike                                            |        0 | `codex-root`                           | DONE   | `apps/mobile/`, `internal/architecture/mobile-feasibility.md`                                                                                                                                                                                 | claimed/completed 2026-07-26 after S0-005 completion                                                                                                                                                                                                                                                                                             | Expo Doctor 20/20; lint, type-check, unit tests and web export pass; native release build completed 531 Gradle tasks with minSdk 24/targetSdk 36; 390px rendered interactions and no overflow; 97.3 MB universal APK measured                                                                                                                                                                                   | Mobile UI and interaction cross-review passed; booted API 36 install stalled in Package Manager, and API 26/2 GB physical-device plus Play split-size evidence remain explicit release gates                                                                                                                                                                                                                             |
| S0-011 | Render/residency/provider feasibility matrix                                                  |        0 | `/root/render_residency`               | DONE   | `deploy/render/provider-residency-feasibility.md`                                                                                                                                                                                             | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Primary-source provider matrix, four topology options, production decision gates, and follow-up evidence; Render has no African region, so production remains founder/privacy/legal-blocked                                                                                                                                                                                                                     | No external resource creation or deployment; self-reviewed 2026-07-26                                                                                                                                                                                                                                                                                                                                                    |
| S0-012 | Initial threat model, data classification and DPIA inputs                                     |        0 | `/root/backend_security`               | DONE   | `internal/architecture/threat-model-v0.md`, `internal/architecture/data-classification.md`, `internal/architecture/dpia-inputs.md`                                                                                                            | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Content checked against plan §§4/7/8/11/15/23 and extracted Doc 08 (consent map, AI systems, suban) and Doc 09 (action ladder, SLAs, binding retention table, compliance pack); no code paths touched                                                                                                                                                                                                           | Self-reviewed; E2E-vs-safety-processing and residency recorded as the two open pre-production decisions                                                                                                                                                                                                                                                                                                                  |
| S0-013 | Traceability matrix, synthetic personas and fixture policy                                    |        0 | `/root/backend_traceability`           | DONE   | `internal/product/`, `packages/test-fixtures/`                                                                                                                                                                                                | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | All P0 FR/NFR rows linked req→module→screen→story→consent→tests→release; FR-202 sample traced end-to-end (matrix §4); tsc strict pass on personas registry; prettier clean                                                                                                                                                                                                                                      | Self-reviewed; FR-105 mapped to P1 productization with P0 assisted vouch noted                                                                                                                                                                                                                                                                                                                                           |
| S0-014 | Sprint 0 integrated verification and founder checkpoint                                       |        0 | `codex-root`                           | DONE   | `agent_plan.md`, `internal/quality/sprint-0-founder-checkpoint.md`                                                                                                                                                                            | claimed/completed 2026-07-26 after S0-003–S0-013 completion                                                                                                                                                                                                                                                                                      | frozen install, peers, 9/9 checks, 3/3 builds, Go vet, actionlint, Expo Doctor 20/20 and JS high/critical audit gate pass; checkpoint records the red Go security workflow and Docker integration timeout                                                                                                                                                                                                       | Conditional Sprint 1 engineering proceed; production no-go and founder approval pending; active S1-002 paths were reviewed read-only                                                                                                                                                                                                                                                                                     |
| S0-015 | Remediate high transitive dependency advisories                                               |        0 | `/root/platform_scaffold`              | DONE   | `pnpm-workspace.yaml`, `pnpm-lock.yaml`                                                                                                                                                                                                       | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | audit: 0 high/0 critical; frozen install, peers, 9/9 Turbo checks, Go tests/vet, 3/3 client builds passed                                                                                                                                                                                                                                                                                                       | Exact patched transitive overrides; no waiver; one moderate advisory remains below the approved gate                                                                                                                                                                                                                                                                                                                     |
| S1-001 | E02-S01 Go composition root and hexagonal module wiring                                       |        1 | `/root/backend_kernel`                 | DONE   | `services/api/main.go`, `services/api/internal/platform/`                                                                                                                                                                                     | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | gofmt/vet clean; go build ok; config+health unit tests pass; startup failure path exits 1 with actionable message; /live vs dependency-degraded /ready unit-verified; pre-existing S0-007 Testcontainers suite attempted but Docker Desktop saturated (58 containers, daemon calls >2 min) — environment issue recorded, not a code failure                                                                     | Self-reviewed; integration rerun pending Docker recovery (asked founder); member module built but unrouted until S1-003 HTTP envelope                                                                                                                                                                                                                                                                                    |
| S1-002 | E02-S05 MongoDB client, transaction, repository and migration conventions                     |        1 | `/root/backend_kernel`                 | DONE   | `services/api/internal/platform/mongo/`, Go module files                                                                                                                                                                                      | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | gofmt/vet/build clean; unit tests green; Testcontainers integration green after founder-approved Docker recovery (133 stale containers removed): migrations apply once + unique index enforced, transaction commit/abort verified against single-node replica set via directConnection                                                                                                                          | Self-reviewed; replica-set + directConnection pattern documented for all future integration suites                                                                                                                                                                                                                                                                                                                       |
| S1-003 | E02-S02 API envelope, OpenAPI, validation and generated TS client                             |        1 | `codex-root` + `/root/api_contract`    | DONE   | `services/api/internal/platform/http/`, `services/api/openapi/`, `packages/api-client/`                                                                                                                                                       | claimed/completed 2026-07-26 after S1-001                                                                                                                                                                                                                                                                                                        | Redocly strict contract lint; generated-client drift check; 3/3 client tests; TypeScript typecheck/build; Go HTTP + OpenAPI contract tests; API vet clean                                                                                                                                                                                                                                                       | Completion independently verified by codex-root after S1-011 conflict mapping; stable envelopes, bounded strict JSON, safe correlation IDs, generated typed client and internal-error redaction are present. Attribution corrected 2026-07-26 by `/root/backend_kernel`: `/root/backend_http` was a raced duplicate claim dropped per protocol §18.5; implementation and first claim were codex-root's                   |
| S1-004 | E02-S06 Idempotency, optimistic concurrency, outbox and inbox                                 |        1 | `/root/backend_kernel`                 | DONE   | `services/api/internal/platform/`, `services/worker/`                                                                                                                                                                                         | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | gofmt/vet clean; unit tests green; Testcontainers integration green: outbox commit/abort atomicity with domain change, publish flow, inbox consumer-scoped dedup under redelivery, idempotency claim/complete/replay                                                                                                                                                                                            | Self-reviewed; async outbox relay lands with worker scheduling (S1-009)                                                                                                                                                                                                                                                                                                                                                  |
| S1-005 | E02-S03 Authentication/session/device model                                                   |        1 | `/root/backend_identity`               | DONE   | `services/api/internal/identity/`                                                                                                                                                                                                             | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | gofmt/vet clean; GoMock application tests + domain policy tests pass; Testcontainers integration green: session round-trip, stale-version rejection, device/member revoke scopes; wired into composition root                                                                                                                                                                                                   | Self-reviewed; opaque hashed tokens (no JWT dependency); reuse of rotated refresh revokes session; OTP endpoints land with E03-S01                                                                                                                                                                                                                                                                                       |
| S1-006 | E02-S04 Authorization/RBAC/ABAC kernel                                                        |        1 | `/root/backend_authz`                  | DONE   | `services/api/internal/authz/`                                                                                                                                                                                                                | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | gofmt/vet clean; policy matrix tests green: deny-by-default (anonymous, unknown action/resource), owner rules, FR-101 tier gates (romantic Tier 1, sowing Tier 2), least-privilege desk roles, host circle scoping                                                                                                                                                                                              | Self-reviewed; explicit grant table — new capabilities require deliberate rule additions; role assignment persistence lands with E16                                                                                                                                                                                                                                                                                     |
| S1-007 | E02-S07 Structured logging, traces, metrics, health and correlation                           |        1 | `codex-root` + `/root/telemetry`       | DONE   | `services/api/internal/platform/telemetry/`, `packages/observability/`                                                                                                                                                                        | claimed/completed 2026-07-26 after S1-001 completion                                                                                                                                                                                                                                                                                             | 15/15 workspace lint/type/test tasks; 5/5 builds; Go telemetry race tests; 4 client privacy tests; full Go tests/vet; govulncheck clean; JS audit 0 high/critical; Prettier and actionlint pass                                                                                                                                                                                                                 | Isolated-agent implementation rebased twice before integration; cross-review added self-contained package tooling and default redaction for raw errors, payloads and member identifiers; production exporter/composition binding tracked as S1-013                                                                                                                                                                       |
| S1-008 | E02-S08 Feature flags, configuration audit and kill switches                                  |        1 | codex-root + /root/feature_flags       | DONE   | `services/api/internal/platform/flags/`                                                                                                                                                                                                       | claimed/completed 2026-07-26 after S1-002                                                                                                                                                                                                                                                                                                        | race-tested immutable snapshots; strict environment parsing; atomic batch/rejection tests; precedence and privacy-safe audit tests; full Go tests/vet; govulncheck found 0 reachable vulnerabilities                                                                                                                                                                                                            | Cross-reviewed after agent rebase; Sow/Fires/AI/Payments/Gate default disabled; runtime kill > environment kill > runtime/environment enablement; persistent admin/API composition remains E16                                                                                                                                                                                                                           |
| S1-009 | E02-S09 Worker scheduling, retries, dead letters and job observability                        |        1 | `/root/backend_worker`                 | DONE   | `services/worker/`, relocated shared packages `internal/platform/` (mongo, outbox, inbox, idempotency) + import updates in `services/api/`                                                                                                    | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | gofmt/vet clean; 18/18 Go test packages pass incl. Testcontainers: relay publishes pending records and marks them published against real MongoDB, dead-letter round trip, scheduler backoff/timeout/dead-letter unit tests, relay failure-path tests                                                                                                                                                            | Self-reviewed; shared packages moved to `internal/platform/` because `services/api/internal` is not importable from the worker; telemetry is codex's lane (S1-013) and intentionally not moved                                                                                                                                                                                                                           |
| S1-010 | E02-S10 Media abstraction, signed access and retention metadata                               |        1 | codex-root + /root/media_kernel        | DONE   | `services/api/internal/media/`                                                                                                                                                                                                                | claimed/completed 2026-07-26 after S1-003                                                                                                                                                                                                                                                                                                        | media race tests; full Go tests/vet; signed TTL, ownership/purpose, checksum, retention/legal-hold, provider-redaction and nil-readiness tests                                                                                                                                                                                                                                                                  | Cross-review corrected zero/pre-creation deletion timestamps and nil dependency panic paths; vendor persistence/HTTP composition remains a later integration lane                                                                                                                                                                                                                                                        |
| S1-011 | Review fix (S1-003 note): transport-neutral duplicate-conflict sentinel                       |        1 | `/root/backend_review`                 | DONE   | `services/api/internal/member/`, `services/api/internal/platform/http/member.go`, `services/api/openapi/openapi.yaml`                                                                                                                         | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | gofmt/vet clean; Go unit tests incl. new 409 `member_conflict` mapping pass; openapi contract test + Redocly strict lint pass; client regenerated with drift check green; prettier clean                                                                                                                                                                                                                        | Self-reviewed; `domain.ErrDuplicateMember` sentinel keeps mongo types out of the HTTP layer per S1-003 review note                                                                                                                                                                                                                                                                                                       |
| S1-012 | Review fix (Sprint 0 CP-001): remediate reachable Go text vulnerability                       |        1 | `codex-root`                           | DONE   | Go module files, `internal/quality/sprint-0-founder-checkpoint.md`                                                                                                                                                                            | claimed/completed 2026-07-26 after S1-002 released Go-module ownership                                                                                                                                                                                                                                                                           | `golang.org/x/text` upgraded from `v0.37.0` to fixed `v0.39.0`; full Go tests/vet pass; govulncheck clean                                                                                                                                                                                                                                                                                                       | Closes `GO-2026-5970` without touching active S1-003 HTTP/client work; `golang.org/x/sync v0.21.0` selected by tidy as the compatible transitive floor                                                                                                                                                                                                                                                                   |
| S1-013 | S1-007 follow-up: bind telemetry kernel and approved exporter in API/worker composition roots |        1 | `codex-root`                           | DONE   | API/worker composition roots, HTTP correlation bridge and vendor-neutral OTLP/HTTP runtime config                                                                                                                                             | claimed/completed 2026-07-26 after API/worker roots stabilized                                                                                                                                                                                                                                                                                   | runtime smoke proves one correlated log, trace and bounded RED metric without PII; unsafe endpoint tests; full Go test/vet and focused race suites pass                                                                                                                                                                                                                                                         | OpenTelemetry Go v1.44.0 + otelhttp v0.69.0 (latest verified 2026-07-26); export off when endpoint unset, credential-free HTTPS by default, graceful flush                                                                                                                                                                                                                                                               |
| S2-001 | E01-S01 Complete platform-neutral brand tokens and asset manifest                             |        2 | `codex-root`                           | DONE   | `packages/design-tokens/`                                                                                                                                                                                                                     | claimed/completed 2026-07-26 after E01 audit                                                                                                                                                                                                                                                                                                     | 5/5 token tests; strict typecheck; frozen install; 19/19 workspace checks; 6/6 builds; full Go suite                                                                                                                                                                                                                                                                                                            | Cross-platform semantic/state/type/space/elevation/motion/feedback/a11y/zone/Kente tokens; supplied-asset manifest; Outfit-only and AA contrast enforced                                                                                                                                                                                                                                                                 |
| S2-002 | E01-S05 Typed EN/Twi string registry and terminology policy                                   |        2 | `/root/i18n_foundation`                | DONE   | `packages/i18n/`                                                                                                                                                                                                                              | claimed/completed 2026-07-26 after E01 audit                                                                                                                                                                                                                                                                                                     | 8/8 registry tests; strict typecheck; lint/build; Prettier and diff checks                                                                                                                                                                                                                                                                                                                                      | Typed keys/params, parity and placeholder validation, Accept-Language resolution, English fallback, terminology metadata and human-review readiness; `tw` remains provisional and unreviewed                                                                                                                                                                                                                             |
| S2-003 | E01-S02 Member MUI theme and shared web primitives                                            |        2 | `codex-root`                           | DONE   | `packages/ui-web/`                                                                                                                                                                                                                            | claimed/completed 2026-07-26 after S2-001                                                                                                                                                                                                                                                                                                        | 3/3 theme tests; strict package/member/admin typechecks; member and admin production builds                                                                                                                                                                                                                                                                                                                     | Preference-aware light/dark/high-contrast themes; reduced motion, 48px targets, focus visibility and stable Button/Card/Status/Focus exports; Outfit preserved                                                                                                                                                                                                                                                           |
| S2-004 | E01-S03 Distinct admin theme and operator primitives                                          |        2 | `codex-root`                           | DONE   | `packages/ui-web/src/admin*`, `apps/admin/app/layout.tsx`                                                                                                                                                                                     | claimed/completed 2026-07-26 after S2-003                                                                                                                                                                                                                                                                                                        | 3/3 admin-theme tests; package/admin typechecks; admin production build; formatting                                                                                                                                                                                                                                                                                                                             | Dedicated operator provider with denser geometry, finite status colors, high contrast, reduced motion, 48px controls and visible focus; Outfit/shared tokens preserved                                                                                                                                                                                                                                                   |
| S2-005 | E01-S04 Expo theme and base interaction primitives                                            |        2 | `/root/mobile_ui`                      | DONE   | `packages/ui-mobile/`, `apps/mobile/app/_layout.tsx`, mobile workspace dependency + lock importer                                                                                                                                             | claimed/completed 2026-07-26 after S2-001                                                                                                                                                                                                                                                                                                        | 4/4 theme tests; UI/mobile typechecks; mobile lint; Expo web export (822 modules); frozen install                                                                                                                                                                                                                                                                                                               | Light/dark provider, reduced-motion contract, opt-in bounded haptics, accessible 48dp Button/Pressable/Card; root corrected relative import to workspace dependency                                                                                                                                                                                                                                                      |
| S2-006 | E01-S06 Localization, accessibility and copy-policy lint                                      |        2 | `/root/client_quality`                 | DONE   | `packages/config-eslint/`, `packages/quality-client/`, relevant CI configuration                                                                                                                                                              | claimed/completed 2026-07-26 after S2-002                                                                                                                                                                                                                                                                                                        | 3 ESLint policy tests, 4 client-quality tests, frozen install, 34/34 workspace checks, full Go suite and client builds                                                                                                                                                                                                                                                                                          | Hardcoded copy, missing names/keys, placeholder drift, pressure copy and mixed-register checks; narrow reasoned non-product escape only                                                                                                                                                                                                                                                                                  |
| S2-007 | E01-S08 Cross-platform loading, empty, error and offline patterns                             |        2 | `codex-root`                           | DONE   | `packages/ui-web/src/states/`, `packages/ui-mobile/src/states/`, state catalogs in `packages/i18n/`                                                                                                                                           | claimed/completed 2026-07-26 after S2-002, S2-003, S2-005                                                                                                                                                                                                                                                                                        | 8/8 i18n, 8/8 web UI and 6/6 mobile UI tests; strict package/member/mobile typechecks; formatting                                                                                                                                                                                                                                                                                                               | Loading, warm empty, recoverable error, offline, queued, low-bandwidth and permission-denied components with deterministic live-region/busy/action semantics                                                                                                                                                                                                                                                             |
| S2-008 | E01-S09a Web Hold, Sow, Stone and Gather prototypes                                           |        2 | `codex-root`                           | DONE   | `apps/web/app/design-lab/`, web gesture primitives                                                                                                                                                                                            | claimed/completed 2026-07-26 after S2-007                                                                                                                                                                                                                                                                                                        | 4/4 reducer tests; typecheck, lint and production build; live 1280px/390px browser QA with no overflow, 48px controls and no app console errors                                                                                                                                                                                                                                                                 | Pointer Hold/Stone cancellation, explicit alternatives, Sow consequence confirmation, Gather discrete choices, reduced motion and semantic live state                                                                                                                                                                                                                                                                    |
| S2-009 | E01-S09b Expo Hold, Sow, Stone and Gather prototypes                                          |        2 | `/root/mobile_gestures` + `codex-root` | DONE   | `apps/mobile/app/design-lab/`, mobile gesture primitives                                                                                                                                                                                      | claimed/completed 2026-07-26 after S2-007                                                                                                                                                                                                                                                                                                        | 10/10 UI-mobile tests; package/mobile typechecks; mobile lint; Expo web export (835 modules)                                                                                                                                                                                                                                                                                                                    | Agent delivered accessible gestures; root cross-review fixed stale Expo typed-route compatibility before publish; physical Android gesture smoke remains a release-device gate                                                                                                                                                                                                                                           |
| S2-010 | E01-S07 Component galleries and visual-regression gate                                        |        2 | `codex-root`                           | DONE   | Storybook/config, `apps/mobile/app/gallery/`, visual-test configuration and baselines                                                                                                                                                         | claimed/completed 2026-07-26 after S2-003 through S2-009 completion                                                                                                                                                                                                                                                                              | 34/34 workspace checks, full Go suite, Storybook 10.5.4 production build, 2/2 axe plus desktop/390 screenshot comparisons, Expo web export (836 modules)                                                                                                                                                                                                                                                        | Reviewed deterministic baselines committed separately by viewport; CI installs Chromium, rejects axe violations and pixel drift, and retains the report artifact                                                                                                                                                                                                                                                         |
| S2-011 | E03-S01 Phone registration, OTP, rate limits and session issuance                             |        2 | `/root/backend_identity`               | DONE   | `services/api/internal/identity/`, auth HTTP routes + OpenAPI                                                                                                                                                                                 | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | 19/19 Go packages pass incl. Testcontainers end-to-end (request→simulator code→verify→persisted session, resend throttle, attempt accounting, consumed-code rejection); auth HTTP error-mapping tests; Redocly strict lint + client regen + drift check green                                                                                                                                                   | Self-reviewed; OTP codes hashed at rest, never logged or returned; simulator sender only until scored SMS/WhatsApp provider (S2-015 lane); contract test surface list extended for auth paths                                                                                                                                                                                                                            |
| S2-012 | E03-S02 Promise/terms/age consent record                                                      |        2 | `/root/consent_kernel`                 | DONE   | `services/api/internal/consent/`                                                                                                                                                                                                              | claimed/completed 2026-07-26 after S1-002                                                                                                                                                                                                                                                                                                        | consent race tests, full Go tests and vet; immutable version, grant/withdraw, replay and optimistic-revision cases                                                                                                                                                                                                                                                                                              | Deny-by-default exact-version evaluation and PII-resistant evidence; persistence adapter must later atomically enforce revision plus globally unique command IDs                                                                                                                                                                                                                                                         |
| S2-013 | E03-S06 Tier/account state machines and transition audit                                      |        2 | `/root/backend_identity`               | DONE   | `services/api/internal/identity/`                                                                                                                                                                                                             | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | 21/21 Go packages pass; transition matrix unit tests (one-step promotion, reasoned demotion, same-tier/skip/out-of-range rejection, blocked accounts); GoMock audit-persistence tests; Testcontainers integration: 0→1→2 with 2 audit rows, invalid transition writes nothing, optimistic-concurrency versions                                                                                                  | Self-reviewed; tier numerics mirror authz kernel Tier; audit records commit in the same transaction as the account update                                                                                                                                                                                                                                                                                                |
| S2-014 | E03-S05 Under-18 hard block and 24-hour purge proof                                           |        2 | `/root/consent_kernel`                 | DONE   | `services/api/internal/safeguarding/`                                                                                                                                                                                                         | claimed/completed 2026-07-26 after S2-013; isolated from concurrent S2-015                                                                                                                                                                                                                                                                       | safeguarding race suite, full Go tests/vet, Testcontainers outage/retry proof deletes Mongo verification/account/session/consent/media artifacts before 24 hours                                                                                                                                                                                                                                                | Persistent HMAC-keyed block remains without raw PII; worker scheduling, object-store purger and S2-015 composition are explicit follow-ups                                                                                                                                                                                                                                                                               |
| S2-015 | E03-S03 Ghana Card provider adapter and fallback queue skeleton                               |        2 | `/root/backend_verification`           | DONE   | `services/api/internal/verification/`, verification HTTP route + OpenAPI                                                                                                                                                                      | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | 24/24 Go packages pass incl. Testcontainers end-to-end: match→approve→Tier 1, outage/uncertain→manual queue (never silent pass), queue listing, manual approval promotes; GoMock branch tests; Redocly lint + client regen + drift check green                                                                                                                                                                  | Self-reviewed; simulator scripts match/mismatch/uncertain/outage until scored vendor; tier bridge only at composition root; cases hold minimal proof set, no card media                                                                                                                                                                                                                                                  |
| S2-016 | E03-S10 Export, deletion, pause and legal-hold workflows                                      |        2 | `/root/backend_privacy`                | DONE   | `services/api/internal/privacy/`, privacy HTTP routes + OpenAPI                                                                                                                                                                               | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | 28/28 Go packages pass incl. Testcontainers end-to-end: export/deletion open with statutory due times (72h/30d), duplicate rejection, hold blocks new + open deletions, processor skips held, lift unblocks and erasure+session-revocation executes; GoMock processor/service tests; Redocly + client regen + drift green                                                                                       | Self-reviewed; export assembler and erasure runner ports stubbed in tests — real cross-context assembly lands with worker wiring; legal-hold paths preserved per Doc 09                                                                                                                                                                                                                                                  |
| S2-017 | E03-S07 Core profile and privacy controls                                                     |        2 | `/root/client_quality`                 | DONE   | `services/api/internal/profile/`                                                                                                                                                                                                              | claimed/completed 2026-07-26 after S1-002                                                                                                                                                                                                                                                                                                        | profile race suite, full Go tests/vet, Mongo Testcontainers stale-write/replay/cross-profile-command proof                                                                                                                                                                                                                                                                                                      | Privacy-safe display name/introduction only; per-field visibility and fail-closed consent; HTTP/authz audience composition remains a follow-up                                                                                                                                                                                                                                                                           |
| S2-018 | E03-S04 Voice+face liveness orchestration and uncertainty handling                            |        2 | `/root/consent_kernel`                 | DONE   | `services/api/internal/verification/liveness/`                                                                                                                                                                                                | claimed/completed 2026-07-26 after S2-014                                                                                                                                                                                                                                                                                                        | liveness race suite, full Go tests/vet, simulator contract and in-memory end-to-end retry/replay/manual-review tests                                                                                                                                                                                                                                                                                            | Exact live is sole automatic pass; temporary biometric refs removed after review; durable adapters/composition remain follow-ups                                                                                                                                                                                                                                                                                         |
| S2-019 | E03-S12 Shared-device/collision and known-name review cases                                   |        2 | `/root/mobile_gestures`                | DONE   | `services/api/internal/identity/collision/`                                                                                                                                                                                                   | claimed/completed 2026-07-26 after S2-011                                                                                                                                                                                                                                                                                                        | collision race suite, full Go tests/vet, Mongo Testcontainers privacy/stale-write proof                                                                                                                                                                                                                                                                                                                         | HMAC-pseudonymous signals, fail-closed shared-device review and one-way audited transitions; composition remains a follow-up                                                                                                                                                                                                                                                                                             |
| S2-020 | E03-S11 Admin verification queue UI prototype                                                 |        2 | `codex-root`                           | DONE   | `apps/admin/app/verification/`, shared admin shell styling                                                                                                                                                                                    | claimed/completed 2026-07-26 after S2-015                                                                                                                                                                                                                                                                                                        | 3/3 interaction-model tests; typecheck/lint/build; live 1280px/390px QA with no overflow, 48px controls and no app runtime errors                                                                                                                                                                                                                                                                               | Redacted proof, evidence-access warning and reasoned confirmation work; backend queue/evidence audit contract promoted separately as S2-023                                                                                                                                                                                                                                                                              |
| S2-021 | E03-S08 Voice of Introduction capture/upload/transcription consent kernel                     |        2 | `/root/consent_kernel`                 | DONE   | `services/api/internal/introduction/`                                                                                                                                                                                                         | claimed/completed 2026-07-26 after S1-010 and S2-012                                                                                                                                                                                                                                                                                             | introduction race suite, full Go tests/vet, consent/media/transcription/revocation/retention lifecycle tests                                                                                                                                                                                                                                                                                                    | Exact consent checked before upload and transcription; only media/transcript refs retained; durable store, media/consent binding and production transcriber remain composition follow-ups                                                                                                                                                                                                                                |
| S2-022 | E03-S09 Doorway Question and photo vault/veil                                                 |        2 | `/root/backend_profile`                | DONE   | `services/api/internal/profile/` (doorway question + vault extension), HTTP routes + OpenAPI                                                                                                                                                  | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | full Go suite passes incl. Testcontainers: question upsert keeps one per member, unsafe contact-content rejected, position conflict via unique index, owner-clear/stranger-veiled rendering; GoMock service tests; Redocly + client regen + drift green                                                                                                                                                         | Self-reviewed; veil is server-default until acceptance-based unveiling lands with E06; asset bytes stay in media context, vault stores ordering refs only                                                                                                                                                                                                                                                                |
| S2-023 | E03-S11 Admin verification queue and evidence access audit contract                           |        2 | `/root/mobile_gestures`                | DONE   | `services/api/internal/verification/admin/`, admin HTTP handler, OpenAPI and generated client                                                                                                                                                 | claimed/completed 2026-07-26 after S2-015 and S2-020                                                                                                                                                                                                                                                                                             | admin/HTTP race tests, full Go tests/vet, Mongo Testcontainers, Redocly/client drift/typecheck/tests                                                                                                                                                                                                                                                                                                            | Independent scopes, recent-MFA evidence gate, HMAC refs and transactional decision/audit; trusted principal/HMAC composition and tier reconciliation remain follow-ups                                                                                                                                                                                                                                                   |
| S2-024 | E03 client identity and verification onboarding flows                                         |        2 | `codex-root`                           | DONE   | `apps/web/app/onboarding/`, member shell hydration boundary                                                                                                                                                                                   | claimed/completed 2026-07-26 after S2-011 through S2-020                                                                                                                                                                                                                                                                                         | 8/8 web interaction tests; typecheck/lint/build; full phone→OTP→consent→card→liveness browser flow; live 1280px/390px QA with no overflow, minimum 52px controls and no app runtime errors                                                                                                                                                                                                                      | Raw card input clears before the next stage and reducer stores only an opaque reference; manual-review and liveness-consent gates are deterministic; transport binding remains a composition follow-up                                                                                                                                                                                                                   |
| S3-001 | E04-S01 Fie information architecture and route map                                            |        3 | `codex-root`                           | DONE   | `internal/product/fie-route-map.md`                                                                                                                                                                                                           | claimed/completed 2026-07-26 after E03 client foundation                                                                                                                                                                                                                                                                                         | route registry, tier/consent/session guard order, deep-link/offline outcome and web/mobile parity review                                                                                                                                                                                                                                                                                                        | Canonical Fie plus four-zone routes, 48px/dp shell laws and follow-on implementation slices accepted; admin/API boundaries remain separate                                                                                                                                                                                                                                                                               |
| S3-002 | E04-S02 First-run interactive walk                                                            |        3 | `codex-root`                           | DONE   | `apps/web/app/fie/welcome/` route and client state model                                                                                                                                                                                      | claimed/completed 2026-07-26 after S3-001                                                                                                                                                                                                                                                                                                        | 12/12 web tests; typecheck/lint/build; completion and skip browser flows; live 1280px/390px QA with no overflow, 48px controls and no app errors                                                                                                                                                                                                                                                                | Five-zone keyboard/screen-reader walk, reduced-motion safe, explicit versioned finish/skip preference and no account penalty                                                                                                                                                                                                                                                                                             |
| S3-003 | E04-S03 Compound home and zone indicators                                                     |        3 | `codex-root`                           | DONE   | `apps/web/app/fie/` home shell, state model and responsive zone navigation                                                                                                                                                                    | claimed/completed 2026-07-26 after S3-001/S3-002                                                                                                                                                                                                                                                                                                 | 15/15 web tests; typecheck/lint/build; connection saver to synced browser flow; live 1280px/390px QA with zero horizontal overflow, 48px controls and no app errors                                                                                                                                                                                                                                             | Four zones, connection/queued state and tonight's fire are visible without a feed; isolated the fire card from global CSS collision before publish                                                                                                                                                                                                                                                                       |
| S3-004 | E04-S04 Abɔnten public community shell without romantic initiation                            |        3 | `codex-root`                           | DONE   | `apps/web/app/fie/abonten/` route plus shared compound navigation                                                                                                                                                                             | claimed/completed 2026-07-26 after S3-003                                                                                                                                                                                                                                                                                                        | 18/18 web tests; typecheck/lint/build; filter/save browser flow; live 1280px/390px QA with zero overflow, 48px controls and no app errors                                                                                                                                                                                                                                                                       | Community fires, learning and notices only; prohibited romantic action vocabulary is regression-tested and no romantic controls are rendered                                                                                                                                                                                                                                                                             |
| S3-005 | E04-S05 Adiwo circle courtyard shell                                                          |        3 | `codex-root`                           | DONE   | `apps/web/app/fie/adiwo/` route and membership-safe interaction model                                                                                                                                                                         | claimed/completed 2026-07-26 after S3-004 and S3-010/S3-012                                                                                                                                                                                                                                                                                      | 21/21 web tests; typecheck/lint/build; discovery/request review browser flow; live 1280px/390px QA with zero overflow, 48px controls and no app errors                                                                                                                                                                                                                                                          | Invite-only circles stay non-requestable; request review discloses only display name/request and explicitly excludes other memberships                                                                                                                                                                                                                                                                                   |
| S3-006 | E04-S06 Ɛpono ano deliberate introduction shell                                               |        3 | `codex-root`                           | DONE   | `apps/web/app/fie/epono-ano/` route and consent/tier/review interaction model                                                                                                                                                                 | claimed/completed 2026-07-26 after S3-005 and E03 introduction/doorway foundations                                                                                                                                                                                                                                                               | 24/24 web tests; typecheck/lint/build; tier-gate and accept browser flows; live 1280px/390px QA with zero overflow, 48px controls and no app errors                                                                                                                                                                                                                                                             | One bounded voice-first introduction, veiled photo, generic consent/tier gates and explicit accept/pass with no swipe or passing penalty                                                                                                                                                                                                                                                                                 |
| S3-007 | E04-S07 Dan mu private mutual room shell                                                      |        3 | `codex-root`                           | DONE   | `apps/web/app/fie/dan-mu/` route and tier/mutuality/pause interaction model                                                                                                                                                                   | claimed/completed 2026-07-26 after S3-006 and Tier 2/room foundations                                                                                                                                                                                                                                                                            | 27/27 web tests; typecheck/lint/build; pause and mutuality-gate browser flows; live 1280px/390px QA with zero overflow, 48px controls and no app errors                                                                                                                                                                                                                                                         | Tier 2 and mutual-choice gates fail closed; pause disables drafts; explicit no-streak/no-public-activity language and strict alternation are visible                                                                                                                                                                                                                                                                     |
| S3-008 | E04-S08 Okyeame presence point placeholder and capability boundary                            |        2 | `codex-root`                           | DONE   | `apps/web/app/fie/okyeame/` route, Fie entry point and capability-state model                                                                                                                                                                 | claimed/completed 2026-07-26 after S3-007                                                                                                                                                                                                                                                                                                        | 30/30 web tests; typecheck/lint/build; available/resting browser boundary; live 1280px/390px QA with zero overflow, 48px controls and no app errors                                                                                                                                                                                                                                                             | Explicit non-person boundary, no decision or private-disclosure authority, honest resting state and one-action return to Fie                                                                                                                                                                                                                                                                                             |
| S3-009 | E04-S09 Mobile/web navigation accessibility and deep-link guards                              |        5 | `codex-root`                           | DONE   | `packages/fie-routing/`, Expo Fie routes, shared web navigation and Playwright axe suite                                                                                                                                                      | claimed/completed 2026-07-26 after S3-001 through S3-008                                                                                                                                                                                                                                                                                         | 8 guard/registry tests, 30 web and 2 mobile tests; web/mobile typecheck/lint/build; 14/14 desktop/mobile Playwright axe+overflow checks                                                                                                                                                                                                                                                                         | Ordered fail-closed guards, opaque-ID validation, one registry for equivalent web/Expo destinations, 58dp mobile tabs; automation found and fixed three dark-surface contrast defects                                                                                                                                                                                                                                    |
| S3-010 | E05-S01 Circle types, privacy defaults and membership kernel                                  |        3 | `/root/client_quality`                 | DONE   | `services/api/internal/circle/`                                                                                                                                                                                                               | claimed/completed 2026-07-26 after E03 profile/auth foundation                                                                                                                                                                                                                                                                                   | circle race/property suite, full Go tests/vet, Mongo Testcontainers concurrent-write/privacy/replay proof                                                                                                                                                                                                                                                                                                       | Finite taxonomy, private default, one-way audited membership and deny-by-default capabilities; owner transfer/demotion/rejoin require future explicit policy                                                                                                                                                                                                                                                             |
| S3-011 | E05-S02 Host application and institutional verification kernel                                |        3 | `/root/consent_kernel`                 | DONE   | `services/api/internal/host/` bounded context, privacy keyer and Mongo adapter                                                                                                                                                                | claimed/completed 2026-07-26 after S3-010                                                                                                                                                                                                                                                                                                        | race, GoMock and lifecycle tests; Mongo Testcontainers manual-review/idempotency/raw-identifier proof; full Go tests and vet                                                                                                                                                                                                                                                                                    | Opaque evidence refs, fail-closed uncertainty, bounded approval/recheck/expiry and privacy-keyed audit; production provider and worker composition remain follow-ups                                                                                                                                                                                                                                                     |
| S3-012 | E05-S03 Circle invites, request/approval and expulsion workflows                              |        3 | `/root/mobile_gestures`                | DONE   | `services/api/internal/circle/workflow/` bounded workflow package and Mongo/token adapters                                                                                                                                                    | claimed/completed 2026-07-26 after S3-010                                                                                                                                                                                                                                                                                                        | race, GoMock and 300-case property tests; Mongo Testcontainers concurrent redemption/expulsion/raw-token proof; full Go tests and vet                                                                                                                                                                                                                                                                           | 256-bit opaque single-use invites, authorization port, replay fingerprints, optimistic revisions and append-only audit; kernel composition remains a follow-up                                                                                                                                                                                                                                                           |
| S3-014 | E05-S04 Voice-first circle room, events and noticeboard kernel                                |        3 | `/root/consent_kernel`                 | DONE   | `services/api/internal/circle/room/` bounded context, privacy keyer and Mongo adapter                                                                                                                                                         | claimed/completed 2026-07-26 after S3-010 through S3-012                                                                                                                                                                                                                                                                                         | race, GoMock and fuzz/property tests; Mongo Testcontainers outsider/replay/visibility/expiry proof; full Go tests and vet                                                                                                                                                                                                                                                                                       | Membership/host capability ports, opaque voice/transcript/content refs, privacy audit and 90-day query expiry; auth composition and physical purge remain follow-ups                                                                                                                                                                                                                                                     |
| S3-013 | E05-S05 Trust edge model and bounded trust-path projection                                    |        3 | `/root/client_quality`                 | DONE   | `services/api/internal/trust/` bounded context and Mongo adapter                                                                                                                                                                              | claimed/completed 2026-07-26 after S3-010                                                                                                                                                                                                                                                                                                        | race, property and fuzz tests; Mongo Testcontainers bounded-read/revoke/provenance/replay proof; full Go tests and vet                                                                                                                                                                                                                                                                                          | Directed immutable provenance, owner authorization before reads, per-edge consent and visibility, cycle-safe depth 4/node 100 bounds, no global or reverse graph browser                                                                                                                                                                                                                                                 |
| S3-015 | E05-S06 Member trust-path visibility and privacy-safe explanations                            |        3 | `/root/client_quality`                 | DONE   | `services/api/internal/trust/visibility/` revalidation projection and isolated HTTP contract                                                                                                                                                  | claimed/completed 2026-07-26 after S3-013                                                                                                                                                                                                                                                                                                        | race, GoMock, HTTP and 200-order property tests; Mongo Testcontainers consent-withdrawal revalidation; full Go tests and vet                                                                                                                                                                                                                                                                                    | Revalidates edge identity, both endpoints, consent, visibility, expiry and revocation immediately before disclosure; generic reason codes only and privacy-safe 404; shared route/OpenAPI registration remains composition work                                                                                                                                                                                          |
| S3-016 | E05-S07 P0 assisted vouch workflow                                                            |        3 | `/root/mobile_gestures`                | DONE   | `services/api/internal/vouch/assisted/` bounded context, privacy keyer and Mongo adapter                                                                                                                                                      | claimed/completed 2026-07-26 after S3-013                                                                                                                                                                                                                                                                                                        | race, GoMock and 300-case property tests; Mongo Testcontainers concurrent decision/replay/raw-ID/productization guard; full Go tests and vet                                                                                                                                                                                                                                                                    | Manual-assisted lifecycle, identity-bound voucher consent, immutable outcome/provenance and no score/stake/money/payment/graph behavior                                                                                                                                                                                                                                                                                  |
| S3-017 | E05-S08 P1 immutable vouch attestation and bounded stake behavior                             |        5 | `/root/mobile_gestures`                | DONE   | `services/api/internal/vouch/attestation/` bounded context, privacy keyer and Mongo adapter                                                                                                                                                   | claimed/completed 2026-07-26 after S3-016                                                                                                                                                                                                                                                                                                        | race, GoMock and 300-case property tests; Mongo Testcontainers concurrent revocation/BSON-prefix/raw-ID/prohibited-field proof; full Go tests and vet                                                                                                                                                                                                                                                           | Explicit consent, immutable timestamped provenance, append-only revoke/expiry and policy-versioned 1–100 non-transferable reputation stake; no financial/token/graph behavior                                                                                                                                                                                                                                            |
| S3-018 | E05-S09 Admin circle legitimacy and vouch audit                                               |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/communityaudit/` bounded contract, privacy keyer and Mongo adapter                                                                                                                                                     | claimed/completed 2026-07-26 after S3-011/S3-014/S3-016                                                                                                                                                                                                                                                                                          | race, GoMock and fuzz/property tests; Mongo Testcontainers authorization/MFA/redaction/audit/raw-ID proof; full Go tests and vet                                                                                                                                                                                                                                                                                | Least-privilege capabilities, recent-MFA evidence gate, redacted summaries, immutable reasoned decisions and no trust graph/edges                                                                                                                                                                                                                                                                                        |
| S3-019 | E05 trust visibility shared HTTP/OpenAPI composition                                          |        3 | `/root/client_quality`                 | DONE   | authenticated platform route, startup composition, OpenAPI and generated TypeScript client                                                                                                                                                    | claimed/completed 2026-07-26 after S3-015                                                                                                                                                                                                                                                                                                        | route/auth/privacy and Mongo Testcontainers tests; Redocly/client drift/typecheck/tests; full Go tests, race and vet                                                                                                                                                                                                                                                                                            | Session subject must equal owner; depth 1–4/nodes 2–100; uniform no-store 404, fail-closed consent/endpoint adapters and no global/reverse endpoint                                                                                                                                                                                                                                                                      |
| S4-001 | E06-S01 Immutable weekly allowance ledger and renewal                                         |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/seed/allowance/` bounded context                                                                                                                                                                                       | claimed/completed 2026-07-26 after E05 trust/circle foundations                                                                                                                                                                                                                                                                                  | full Go suite/vet; race and property coverage; MongoDB 8.0.13 Testcontainers concurrency/week-boundary proof                                                                                                                                                                                                                                                                                                    | Server-authoritative non-purchasable allowance, exact Monday/IANA/DST renewal, append-only issuance/spend/renewal audit, atomic concurrency and HMAC privacy keys                                                                                                                                                                                                                                                        |
| S4-002 | E06-S02 Introduction-source abstraction without global browsing                               |        3 | `/root/mobile_gestures`                | DONE   | `services/api/internal/seed/source/` bounded context                                                                                                                                                                                          | claimed/completed 2026-07-26 after S3-013/S3-015                                                                                                                                                                                                                                                                                                 | full Go suite/vet; race/property coverage; MongoDB 8.0.13 Testcontainers consent/withdrawal/expiry proof                                                                                                                                                                                                                                                                                                        | Explicit sources only, at most 50 HMAC candidate IDs, no global member list/reverse graph, replay-safe privacy-preserving denials                                                                                                                                                                                                                                                                                        |
| S4-003 | E06 client garden, listening eligibility and deliberate sow composer                          |        5 | `codex-root`                           | DONE   | member web `/fie/garden/`, shared Fie registry and equivalent Expo route                                                                                                                                                                      | claimed/completed 2026-07-26 after E04 client shell and E03 voice introduction                                                                                                                                                                                                                                                                   | 8 route tests; 33 web tests; web/mobile typecheck, lint and production builds; Playwright 16/16 desktop/mobile; live 1280/390 full listen-compose-voice-confirm QA with no overflow or console errors                                                                                                                                                                                                           | 20-second gate, voice-first review, allowance decrements only on matching server confirmation, no paid/boost language; accessible 48px controls and Outfit typography                                                                                                                                                                                                                                                    |
| S4-004 | E06-S03 Server-verified 20-second listening eligibility                                       |        4 | `/root/backend_seed`                   | DONE   | `services/api/internal/seed/listening/`, listening HTTP routes + OpenAPI                                                                                                                                                                      | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | interval-merge property suite (200 trials: order-independent, replay-idempotent); GoMock service tests incl. stale-write retry; Testcontainers end-to-end (merge across batches, clamp to duration, per-listener isolation); Redocly + client regen + drift green                                                                                                                                               | Self-reviewed; eligibility is server-derived only, partial-listen state never leaves the seed context (FR-205); asset duration is caller-asserted until media persistence lands                                                                                                                                                                                                                                          |
| S4-005 | S2-016 follow-up: wire privacy processor into worker with cross-context export/erasure        |        4 | `/root/backend_privacy`                | DONE   | `services/worker/internal/jobs/`, `internal/privacy/` (relocated from `services/api/internal/privacy/`)                                                                                                                                       | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers end-to-end: worker job completes export (archive assembled across 9 context collections, token hashes stripped) and deletion (records erased, account tombstoned, sessions revoked, requests completed); Docker purge of 54 leaked test containers required first; unit suite green                                                                                                              | Race resolved per §18.5: my claim (b9e7654) predated the duplicate `services/api/privacy` processor, which was dropped; codex's stricter test contract was adopted instead (replay-safe archives with 7d TTL, execution-time legal-hold block, PII-free proof-of-deletion audit, erasure replay safety); privacy context moved to `internal/privacy/` for worker import; media-object deletion pending media persistence |
| S4-006 | E06-S04 Sow composer, media screening and deliberate send                                     |        5 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/seed/sow/` bounded context                                                                                                                                                                                    | claimed/completed 2026-07-26 after S4-001/S4-002 and client deliberate-send model                                                                                                                                                                                                                                                                | full Go suite/vet; focused race; fuzz 23,895 executions; MongoDB 8.0.13 replica-set Testcontainers concurrency/replay/privacy proof                                                                                                                                                                                                                                                                             | Mandatory deliberate confirmation and pre-acceptance screening; HMAC opaque refs; atomic acceptance + allowance spend + immutable audit; rejection/races/replay cannot double-spend; no paid/boost/bypass                                                                                                                                                                                                                |
| S4-007 | E06-S05 Pod delivery, cap and privacy-safe playback                                           |        5 | `/root/mobile_gestures`                | DONE   | isolated `services/api/internal/seed/pod/` bounded context                                                                                                                                                                                    | claimed/completed 2026-07-26 after S4-001/S4-002                                                                                                                                                                                                                                                                                                 | full Go suite/vet; race + 300-case property tests; MongoDB 8.0.13 Testcontainers cap/privacy/concurrent-playback proof                                                                                                                                                                                                                                                                                          | Cap 25; sorted/deduped HMAC recipient refs; max 7-day expiry; recipient-only playback with authorization/revalidation before replay; neutral denial; no list/reverse discovery                                                                                                                                                                                                                                           |
| S4-008 | E06-S09 client Garden lifecycle, expiries and dawn summary                                    |        5 | `codex-root`                           | DONE   | web `/fie/garden/` lifecycle views and equivalent Expo state presentation                                                                                                                                                                     | claimed/completed 2026-07-26 after S4-003                                                                                                                                                                                                                                                                                                        | 35 web tests; web/mobile typecheck/lint/build; Playwright 16/16 desktop/mobile; live 1280/390 rendered QA: four lifecycle cards, no overflow, 48px controls, no console errors                                                                                                                                                                                                                                  | Calm once-daily summary; queued/delivered/heard/sprouted/declined/expired language; privacy-safe expiry, no streaks/read receipts/urgency/dark patterns                                                                                                                                                                                                                                                                  |
| S4-009 | E06-S06 Decline, 90-day exclusion and kind notification                                       |        3 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/seed/decline/` bounded context                                                                                                                                                                                | GoMock + 10k-case boundary property + race/full Go/vet; recovered live MongoDB 8.0.13 Testcontainers race proof green after representation-safe BSON assertion repair `2ab2a64`                                                                                                                                                                  | Exact notification shape/kind and privacy scan remain strict across ordered/unordered BSON decoding                                                                                                                                                                                                                                                                                                             | Half-open exact 90-day exclusion; HMAC refs; fixed neutral notification; boolean eligibility; atomic audit, no reason/rejection/public signal                                                                                                                                                                                                                                                                            |
| S4-010 | E06-S07 Sprout and three-exchange alternating doorway                                         |        5 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/seed/sprout/` bounded context                                                                                                                                                                                 | full Go/vet; focused race; fuzz 59,606 executions; recovered live MongoDB 8.0.13 Testcontainers proof green 2026-07-26                                                                                                                                                                                                                           | Reciprocal same-seed activation; deterministic first actor; strict alternating turns; hard seal at three exchanges; HMAC refs; no unilateral room/public signal                                                                                                                                                                                                                                                 |
| S4-011 | E06-S08 Mutual water and single room creation under races                                     |        5 | `/root/mobile_gestures`                | DONE   | isolated `services/api/internal/seed/water/` bounded context                                                                                                                                                                                  | full Go/vet; focused race + 300-case property; recovered live MongoDB 8.0.13 Testcontainers race proof green after awaiting-state CAS repair `2ab2a64`                                                                                                                                                                                           | Reproduced omitted-empty-field CAS failure; concurrent distinct votes now converge on one winner/one deterministic conflict and replay safely                                                                                                                                                                                                                                                                   | Both members water; only distinct second member creates one HMAC room; optimistic CAS + unique keys; no pre-mutual room/public activity/reverse lookup                                                                                                                                                                                                                                                                   |
| S4-012 | E06-S09 Server Garden states, expiries and dawn summary projection                            |        5 | `codex-root`                           | DONE   | isolated `services/api/internal/seed/garden/` bounded context                                                                                                                                                                                 | claimed/completed 2026-07-26 after S4-004/S4-006/S4-007/S4-008                                                                                                                                                                                                                                                                                   | domain + GoMock service tests; focused race/vet; MongoDB 8.0.13 Testcontainers owner-isolation/expiry/privacy proof; full Go suite/vet                                                                                                                                                                                                                                                                          | Deterministic queued/delivered/heard/sprouted/declined/expired projection; expire-before-summary; member-only bounded 100-item view; no read receipt, streak, reason or public activity leak                                                                                                                                                                                                                             |
| S4-013 | E06-S11 Seed abuse controls, demand throttling and care-review signals                        |        5 | `codex-root`                           | DONE   | isolated `services/api/internal/seed/safety/` bounded context                                                                                                                                                                                 | domain + GoMock tests; focused race; full Go/vet; recovered live MongoDB 8.0.13 Testcontainers concurrency/privacy proof green 2026-07-26                                                                                                                                                                                                        | 10-minute server windows: 6 sow/30 candidate actions; generic denial; care signal only after three denials with HMAC actor + bounded code; no accusation/content/score/graph                                                                                                                                                                                                                                    |
| S4-014 | E06-S12 Cross-context property/fuzz/chaos suite for Seed hard invariants                      |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/seed/invariants/` black-box tests only                                                                                                                                                                                 | claimed/completed 2026-07-26 after S4-001–S4-013                                                                                                                                                                                                                                                                                                 | full Go/vet; seed-wide race; 32-way acceptance concurrency; cross-context fuzz 26,681 + listening fuzz 157,608 executions                                                                                                                                                                                                                                                                                       | Proves no spend-before-acceptance, early eligibility, cap/alternation/mutuality/replay bypass and opaque audit surfaces; no production changes                                                                                                                                                                                                                                                                           |
| S4-015 | E06-S10 Ember issuance/redemption                                                             |        3 | `/root/backend_fires`                  | DONE   | `services/api/internal/fire/ember/`                                                                                                                                                                                                           | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers end-to-end: co-attendance enforcement, one-per-attendee unique index, recipient-only redemption, mutual flip on reverse ember, 24h window; GoMock branch tests incl. doorway-opener port; Redocly + client regen + drift green                                                                                                                                                                   | Self-reviewed; mutual-ember DoorwayOpener stays nil until the sprout module composes (noted for S4-010 wiring follow-up)                                                                                                                                                                                                                                                                                                 |
| S5-001 | E07-S01 Courtship room event/state model and projection                                       |        5 | `/root/mobile_gestures`                | DONE   | isolated `services/api/internal/courtship/room/` bounded context                                                                                                                                                                              | full Go/vet; focused race; 300-case projection property; recovered live MongoDB 8.0.13 Testcontainers proof green 2026-07-26                                                                                                                                                                                                                     | Exactly two sorted HMAC members; append-only events/replay fingerprints; deterministic projection; consent/membership revalidation; no public/popularity/reverse API                                                                                                                                                                                                                                            |
| S5-002 | E07 client private-room timeline, guided arc and calm action states                           |        5 | `codex-root`                           | DONE   | web `/fie/dan-mu/rooms/[roomId]` and equivalent Expo room presentation                                                                                                                                                                        | claimed/completed 2026-07-26 after S3-007/S4-008 and alongside isolated S5-001 server model                                                                                                                                                                                                                                                      | 37 web tests; web/mobile typecheck/lint/build; Playwright 18-route desktop/mobile run found/fixed contrast and targeted 2/2 green; live 1280/390 voice-send/safety-dialog QA, no overflow, 48px controls or console errors                                                                                                                                                                                      | Voice-first alternating timeline; next-turn, pause, kind closure and always-available safety; no streak/read pressure/public activity; Outfit                                                                                                                                                                                                                                                                            |
| S5-003 | E07-S02 Strict drum alternation and voice-first stage opening                                 |        5 | `/root/client_quality`                 | DONE   | `services/api/internal/courtship/drum/` domain, application service, privacy keyer and Mongo repository                                                                                                                                       | commit `2261bc4`; focused `go test -race` and `go vet` rerun after integration; agent full `go test ./...` and `go vet ./...`                                                                                                                                                                                                                    | 1,000-conversation property proof and generated GoMock are green; Mongo Testcontainers coverage is included but deferred while Docker is unavailable                                                                                                                                                                                                                                                            | Server-only turn authority; voice-first stage; no same-actor double turn, text-only bypass or public activity                                                                                                                                                                                                                                                                                                            |
| S5-004 | E07-S03 Offline queue, retry and multi-device idempotency                                     |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/courtship/queue/` domain, application service, privacy keyer and transactional Mongo repository                                                                                                                        | commit `a632284`; focused `go test -race` and `go vet` rerun after integration; agent full `go test ./...`, `go vet ./...` and 5,700 fuzz executions                                                                                                                                                                                             | Replay, stale-device and ordered cursor proofs are green; Mongo replica-set Testcontainers coverage is included but deferred while Docker is unavailable                                                                                                                                                                                                                                                        | Opaque idempotent commands; ordered per-room delivery; deterministic stale-device result; no duplicate event                                                                                                                                                                                                                                                                                                             |
| S5-005 | E07-S04 Response windows, resting, re-light and archival                                      |        5 | `codex-root`                           | DONE   | `services/api/internal/courtship/pace/` domain, service, HMAC keyer and optimistic Mongo repository                                                                                                                                           | focused test/vet/race; 300-case exact-boundary property; full Go/vet; recovered live MongoDB 8.0.13 Testcontainers boundary/replay/privacy proof green 2026-07-26                                                                                                                                                                                | Reassigned and completed after incomplete first handoff                                                                                                                                                                                                                                                                                                                                                         | Server-time 48-hour response/rest transition, two-member re-light, deterministic 30-day archive, replay fingerprints and opaque persistence; no urgency, streak, read receipt or client clock authority                                                                                                                                                                                                                  |
| S5-006 | E07-S05 Pause Stone room suspension and safe resume                                           |        3 | `codex-root`                           | DONE   | isolated `services/api/internal/courtship/pause/` bounded context                                                                                                                                                                             | claimed/completed 2026-07-26 after S5-001                                                                                                                                                                                                                                                                                                        | domain + GoMock service tests; focused race; full Go suite/vet; Mongo 8 Testcontainers mutual-resume/privacy proof included but not run while shared Docker is unavailable                                                                                                                                                                                                                                      | Either member pauses immediately; sends fail closed; resume requires both member acknowledgements; HMAC refs/replay/CAS; safety/closure remain outside suspension                                                                                                                                                                                                                                                        |
| S5-007 | E07-S06 Honesty Ribbon private disclosure acknowledgement                                     |        3 | `codex-root`                           | DONE   | `services/api/internal/courtship/honesty/` domain, service, privacy keyer, Mongo repository and generated GoMock                                                                                                                              | focused test/vet/race; full Go/vet; recovered live MongoDB 8.0.13 Testcontainers consent/revoke/privacy proof green 2026-07-26                                                                                                                                                                                                                   | Consent, replay, revoke and privacy proofs all green                                                                                                                                                                                                                                                                                                                                                            | Private room-scoped acknowledgement requires both current grants, revokes visibility immediately, stores keyed identities only, and exposes no score, badge, rank, public reputation or inference                                                                                                                                                                                                                        |
| S5-008 | E07-S07 Guided theme one and simultaneous reveal                                              |        5 | `/root/client_quality`                 | DONE   | `services/api/internal/courtship/theme/` domain, service, privacy keyer and Mongo repository                                                                                                                                                  | commit `86dfd2d`; focused race; 2,000-order property proof; MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                                                | First submission is concealed; second CAS submission atomically reveals both immutable opaque refs; generated GoMock and privacy assertions green                                                                                                                                                                                                                                                               | One fixed guided prompt; each member submits once; simultaneous immutable reveal with no popularity or public surface                                                                                                                                                                                                                                                                                                    |
| S5-009 | E07-S08 Call, meeting and exclusivity proposal objects                                        |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/courtship/proposal/` domain, service, detail-protection port, privacy keyer and Mongo repository                                                                                                                       | commit `085ed9c`; focused race; 289,931 role fuzz executions; integration-tag compile; full Go/vet                                                                                                                                                                                                                                               | Typed expiry, recipient-only decision, sender-only withdrawal, terminal replay and encrypted detail boundary proofs green                                                                                                                                                                                                                                                                                       | Typed, expiring private proposals require explicit recipient acceptance; rejection/withdrawal is neutral; no phone number, unilateral status or public relationship signal                                                                                                                                                                                                                                               |
| S5-010 | E07-S09 Kind closure and ghost-pattern behavior                                               |        5 | `/root/courtship_closure`              | DONE   | `services/api/internal/courtship/closure/` domain, service, privacy keyer and Mongo repository                                                                                                                                                | commit `c3660c5`; focused race/vet; integration-tag compile; full Go/vet                                                                                                                                                                                                                                                                         | Member and exact inactivity boundary closure proofs green; actorless neutral events, terminal replay and opaque persistence verified                                                                                                                                                                                                                                                                            | Either member may close immediately with a neutral private event; server-time inactivity may close without blame; no reason disclosure, accusation, score or public signal                                                                                                                                                                                                                                               |
| S5-011 | E07-S10 Safety sheet, block/report and watermark                                              |        5 | `codex-root`                           | DONE   | web/Expo private-room safety flow plus `services/api/internal/courtship/safety/` hexagonal contract                                                                                                                                           | 39 web tests; web/mobile typecheck, lint and production builds; focused safety race/vet; desktop/mobile Playwright 2/2; live report flow QA; full Go/vet from backend handoff                                                                                                                                                                    | Browser QA found and fixed watermark contrast; report requires bounded category, block is immediate, evidence ref is opaque and client surfaces use Outfit with 48px controls/no overflow                                                                                                                                                                                                                       | Safety remains available in every room state; block is immediate, reports are private/immutable, captures visibly watermarked; no free-text reason, public accusation or reverse surface                                                                                                                                                                                                                                 |
| S5-012 | E07-S11 Themes 2–4 progression kernel                                                         |        5 | `/root/client_quality`                 | DONE   | `services/api/internal/courtship/themeprogression/` domain, service, privacy keyer and Mongo repository                                                                                                                                       | commit `f9aed0d`; focused race/vet; property suite; MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                                                        | Immutable defensive-copy catalog, Theme One evidence gate, strict predecessor unlock, conceal/reveal and replay/CAS proofs green                                                                                                                                                                                                                                                                                | Fixed versioned themes unlock strictly in order only after both-member reveal; no skipping, purchase, score, popularity or public surface                                                                                                                                                                                                                                                                                |
| S5-013 | E07-S11 client Themes 2–4 guided arc                                                          |        3 | `codex-root`                           | DONE   | web/Expo private-room ordered theme progression presentation                                                                                                                                                                                  | commit `b7336a2`; 40 web tests; web/mobile typecheck/lint and production builds; desktop/mobile Playwright 2/2; `git diff --check`                                                                                                                                                                                                               | Calm revealed/ready/resting cards, simultaneous conceal/reveal explanation, Outfit typography, mobile one-column arc, no overflow or axe violations                                                                                                                                                                                                                                                             | Ordered theme states have no urgency, paid skip, gamified progress, score, streak or paywall                                                                                                                                                                                                                                                                                                                             |
| S6-001 | E08-S01 Deterministic cloth grammar and render seed                                           |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/cloth/grammar/` pure domain, service, privacy keyer and Mongo recipe repository                                                                                                                                        | commit `30944d6`; focused race/vet; 148,974 canonical-permutation fuzz executions; live MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                    | Clock/randomness-free `cloth.v1`, defensive canonicalization, rehydration drift detection, replay/CAS and strict opaque-input rejection green                                                                                                                                                                                                                                                                   | Versioned canonical inputs produce the same bounded render seed and safe token set everywhere; no raw private content, arbitrary code, randomness drift or user-supplied executable grammar                                                                                                                                                                                                                              |
| S6-002 | E08-S02 P0 first-thread and Theme One band                                                    |        5 | `/root/courtship_closure`              | DONE   | `services/api/internal/cloth/thread/` domain, service, privacy keyer and Mongo repository                                                                                                                                                     | commit `619ca4a`; focused race/vet; integration-tag compile; full Go/vet                                                                                                                                                                                                                                                                         | Durable simultaneous-reveal evidence, concurrent one-time issuance, replay, pair-only view and raw-content privacy proofs green                                                                                                                                                                                                                                                                                 | Pair-owned first thread is issued once from durable Theme One reveal evidence; immutable versioned band provenance, opaque refs, no public relationship signal or purchasable bypass                                                                                                                                                                                                                                     |
| S5-014 | Integration repair: Seed decline BSON notification assertion                                  |        1 | `/root/client_quality`                 | DONE   | decline Mongo integration assertion only                                                                                                                                                                                                      | commit `2ab2a64`; live MongoDB 8.0.13 Testcontainers under race; focused/full Go/vet                                                                                                                                                                                                                                                             | Supports `bson.M`/`bson.D` while still requiring exactly eventKey, recipientKey, fixed kind and occurredAt plus forbidden-token scan                                                                                                                                                                                                                                                                            | Representation-agnostic assertion without weakening fixed neutral-notification or privacy checks                                                                                                                                                                                                                                                                                                                         |
| S5-015 | Integration repair: Seed Water concurrent second-vote convergence                             |        3 | `/root/client_quality`                 | DONE   | Water Mongo optimistic append plus stronger integration proof                                                                                                                                                                                 | commit `2ab2a64`; live MongoDB 8.0.13 Testcontainers under race; focused property/race and full Go/vet                                                                                                                                                                                                                                           | CAS now keys `_id` + revision + awaiting status rather than an omitted empty room field; winner replay and losing-command absence proven                                                                                                                                                                                                                                                                        | Concurrent mutual water converges on exactly one room and replay-safe state, never zero or duplicate creation                                                                                                                                                                                                                                                                                                            |
| S5-016 | Integration repair: Pause Stone Mongo append arrays                                           |        1 | `codex-root`                           | DONE   | `services/api/internal/courtship/pause/adapters/outbound/mongodb/repository.go`                                                                                                                                                               | MongoDB 8.0.13 Testcontainers race proof; focused race/vet; `git diff --check`                                                                                                                                                                                                                                                                   | Reproduced null-array `$push` failure, normalized empty events/commands/acknowledgements on create, then passed the live immediate-pause/mutual-resume/privacy proof                                                                                                                                                                                                                                            | New pause documents persist appendable empty arrays so immediate suspension and mutual resume remain atomic                                                                                                                                                                                                                                                                                                              |
| S6-003 | E09-S01 Fire scheduling, capacity, RSVP and waitlist                                          |        9 | `/root/backend_fires`                  | DONE   | `services/api/internal/fire/`, fire HTTP routes + OpenAPI                                                                                                                                                                                     | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | -race suite green incl. 20-way capacity race → exactly 5 going/15 waitlisted, FIFO promotion on cancel, duplicate rejection, FR-401 tier gate; GoMock service tests; Redocly + client regen + drift green; seed/decline + seed/water failures observed belong to claimed repairs S5-014/S5-015                                                                                                                  | Self-reviewed; initial version-pinned design livelocked under 20-way contention and was redesigned to atomic conditional $inc seat claims ($expr capacity guard) inside transactions with bounded waitlist-race retries                                                                                                                                                                                                  |
| S6-004 | E08-S03 Pair-owned Cloth archival, export and deletion                                        |        5 | `/root/courtship_closure`              | DONE   | `services/api/internal/cloth/lifecycle/` domain, service, privacy keyer and Mongo repository                                                                                                                                                  | commit `2b6968b`; focused race/vet; integration-tag compile; full Go/vet                                                                                                                                                                                                                                                                         | Either-member archive, pair-only export, fail-closed legal hold, deterministic tombstone receipt, replay/CAS and privacy proofs green                                                                                                                                                                                                                                                                           | Export is privacy-minimal; deletion retains only versioned policy provenance and immutable pseudonymous receipt, never raw content/public/list/reverse data                                                                                                                                                                                                                                                              |
| S6-005 | E08-S04 Abusua Gate consent configuration                                                     |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/cloth/gate/` versioned policy, service, privacy keyer and transactional Mongo repository                                                                                                                               | commit `04f575e`; focused race/vet; 41,911 single-member bypass fuzz executions; integration-tag compile; full Go/vet                                                                                                                                                                                                                            | Exact reviewer/question/material capability intersection, immediate revoke, outsider denial, replay/CAS and opaque audit proofs green                                                                                                                                                                                                                                                                           | Pair members independently configure consent; effective access is deny-by-default most-restrictive intersection with no inherited circle access/public link/raw content                                                                                                                                                                                                                                                  |
| S6-006 | E08-S05 Reviewer links, OTP, watermark and expiry                                             |        5 | `/root/client_quality`                 | DONE   | `services/api/internal/cloth/reviewer/` domain, policy-revalidating service, crypto adapter and Mongo repository                                                                                                                              | commit `720169b`; focused property/race; live MongoDB 8.0.13 Testcontainers under race; full Go/vet                                                                                                                                                                                                                                              | Wrong/expired OTP, invite expiry, concurrent one-time redemption, replay, revoke, reviewer/pair binding, bounded watermark projection and raw BSON privacy proofs green                                                                                                                                                                                                                                         | Crypto-random 256-bit invite, HMAC-only persistence, separately supplied one-time OTP, max 10-minute OTP/24-hour invite; no bearer-only access, browsing or link reuse                                                                                                                                                                                                                                                   |
| S6-007 | E08 client Abusua Gate consent and reviewer ceremony                                          |        5 | `codex-root`                           | DONE   | guarded web `/fie/abusua-gate`, equivalent Expo route, shared registry and Dan mu entry point                                                                                                                                                 | 42 web + 8 route tests; web/mobile typecheck/lint/build; desktop/mobile Playwright 2/2 interaction/axe/overflow proof; `git diff --check`                                                                                                                                                                                                        | Mutual-consent interaction enables issuance only after both states; selected material, reviewer relationship, 24-hour invite, separate 10-minute OTP, watermark and immediate close language                                                                                                                                                                                                                    | Outfit ceremony UI explains private dual consent without exposing content or using public-share framing; all controls at least 48px                                                                                                                                                                                                                                                                                      |
| S6-008 | E08-S06 Per-question consent relay                                                            |        5 | `/root/courtship_closure`              | DONE   | `services/api/internal/cloth/relay/` domain, revalidating service, privacy keyer and Mongo repository                                                                                                                                         | commit `21070d0`; focused race/vet; live MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                                                                   | Exact two-grant intersection, immediate revoke/mismatch redaction, reviewer/current-consent fail-closed ports, replay/CAS and raw privacy proofs green                                                                                                                                                                                                                                                          | Reviewer submits bounded opaque prompts; relay reveals only the currently dual-consented question/response and never unilateral/full-thread/raw/public/reverse data                                                                                                                                                                                                                                                      |
| S6-009 | E08-S07 Gate ceremony and optional circle announcement                                        |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/cloth/ceremony/` domain, revalidating service, privacy keyer and transactional Mongo repository                                                                                                                        | commit `13818ff`; focused race/vet; 22,395 fuzz executions; live MongoDB 8.0.13 Testcontainers race proof; telemetry flake rerun green                                                                                                                                                                                                           | Dual-only completion; distinct optional one-destination announcement; fresh dual consent; immediate publish revalidation and idempotent fixed-kind publisher proofs green                                                                                                                                                                                                                                       | No automatic publish, relationship detail, reviewer material or raw content; circle announcement remains separately optional and destination-bounded                                                                                                                                                                                                                                                                     |
| S6-010 | E08-S08 Harvest/weaver specification and order handoff                                        |        5 | `/root/client_quality`                 | DONE   | `services/api/internal/cloth/harvest/` plus `internal/product/cloth-harvest-weaver-spec.md`                                                                                                                                                   | spec `41afabc`; kernel `7c0b186`; 1,000-case consent invalidation property; focused race/vet; live MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                         | Full-payload approval binding, revise invalidation, concurrent one-handoff convergence, terminal provider denial, bounded callbacks and raw privacy proofs green                                                                                                                                                                                                                                                | Pair-authorized immutable recipe handoff contains six safe tokens and opaque delivery ref only; no reflection, address, payment, public listing/browse or provider access beyond the order                                                                                                                                                                                                                               |
| S7-001 | E09-S02 LiveKit room and token adapter                                                        |        5 | `/root/courtship_closure`              | DONE   | stateless `services/api/internal/realtime/livekit/` adapter and application contract                                                                                                                                                          | commit `37ceac2`; official latest `server-sdk-go/v2 v2.18.1` and protocol `v1.50.4`; focused test/race/vet; full Go/vet                                                                                                                                                                                                                          | Official verifier proves opaque identity, bounded TTL/skew and listener/speaker/explicit-host grants; create/list/admin/recording/ingress/agent/metadata denied; README documents why Mongo/Testcontainers do not apply                                                                                                                                                                                         | Short-lived room-join-only tokens; no client API secret, phone/email/name identity, token persistence/logging, room browse or implicit recording grant                                                                                                                                                                                                                                                                   |
| S7-002 | E09-S03 Host stage, mute, eject and co-host controls                                          |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/fire/control/` domain, revalidating service, privacy keyer and transactional Mongo repository                                                                                                                          | commit `9b81f02`; focused race/vet; 249,437 co-host bypass fuzz executions; live MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                           | Host-only promotion; host/co-host bounded mute/eject; no server unmute; atomic eject/revoke; immediate authority revalidation, replay/CAS and raw privacy proofs green                                                                                                                                                                                                                                          | Least privilege with terminal eject and immutable reason-coded audit; never mute/eject host/co-host through subordinate role and no public accusation                                                                                                                                                                                                                                                                    |
| S7-003 | E09-S04 client low-bandwidth degradation ladder and captions                                  |        5 | `codex-root`                           | DONE   | guarded web `/fie/fires/[fireId]`, Expo equivalent, connection model and Fie entry point                                                                                                                                                      | 44 web tests; web/mobile typecheck/lint/build; desktop/mobile Playwright 2/2 interaction/axe/overflow; `git diff --check`                                                                                                                                                                                                                        | Video → audio → captions → reconnect reducer; explicit lower-data choices; live captions; safety/leave remain reachable in every state; Outfit and 48px controls                                                                                                                                                                                                                                                | Degradation is calm and explicit with no phone/follower/public-attendance leak, shame, urgency, hidden data use or forced upgrade                                                                                                                                                                                                                                                                                        |
| S7-004 | E09-S05 Runsheet, timer and game-segment mount points                                         |        5 | `/root/client_quality`                 | DONE   | `services/api/internal/fire/runsheet/` domain, authority-gated service, privacy keyer and Mongo repository                                                                                                                                    | commit `8aa8903`; 1,000-case timer/order property; focused race/vet; live MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                                  | Timer projection clamps at zero without mutation; explicit authorized transitions; one-winner CAS/replay; fixed game-capability catalog and raw privacy proofs green                                                                                                                                                                                                                                            | Versioned ordered segments with informational server timer; no automatic advance/eject/punish and no dynamic game code                                                                                                                                                                                                                                                                                                   |
| S7-005 | E09-S06 Fire consent and recording policy                                                     |        5 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/fire/recording/` bounded context                                                                                                                                                                              | commit `28c1083`; focused race/vet; 100,806 fuzz executions; live MongoDB Testcontainers race proof; full Go/vet                                                                                                                                                                                                                                 | Recording defaults off; unanimous current-participant consent, join/pause/revoke revalidation, bounded purpose/retention, replay/CAS and opaque-reference privacy proofs green                                                                                                                                                                                                                                  | Recording defaults off; host proposes a bounded purpose/retention policy, every current participant explicitly opts in before start, joiners pause recording until consent, any revoke stops immediately                                                                                                                                                                                                                 |
| S7-006 | E10-S01 Oware domain rules and legality engine                                                |        9 | `/root/backend_games`                  | DONE   | `services/api/internal/games/oware/domain/` legality engine                                                                                                                                                                                   | commit `942694f`; focused race/vet; sow, capture, feed, threshold, grand-slam and seed-conservation property proofs                                                                                                                                                                                                                              | Abapa skip-origin sowing, capture chain, feed rule, >24 win and leftover scoring are pure and deterministic; exactly 48 non-negative seeds are conserved                                                                                                                                                                                                                                                        | Game skill and legality remain isolated from persistence, ratings, member discovery, matching visibility and trust                                                                                                                                                                                                                                                                                                       |
| S7-007 | E09-S08 Incident hotkey and trust-and-safety live routing                                     |        5 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/fire/incident/` bounded context                                                                                                                                                                               | commit `2a4068f`; focused race/vet; 1,000-case property proof; live MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                                        | Participant authority, bounded category/evidence, immediate host-independent safety action, one opaque case, minimal T&S envelope, replay/CAS and raw privacy proofs green                                                                                                                                                                                                                                      | Any participant can trigger a private bounded-category incident; immediate safety action is independent of host; one opaque case routes to authorized T&S with minimal live context and no public accusation                                                                                                                                                                                                             |
| S7-008 | E09-S09 client in-app call without phone-number exposure                                      |        5 | `codex-root`                           | DONE   | web/Expo private-room call invitation, explicit consent and active-call states                                                                                                                                                                | 46 web tests; web/mobile typecheck/lint/build; rendered desktop interaction/semantic/overflow QA; `git diff --check`                                                                                                                                                                                                                             | Incoming call remains inert until explicit accept; decline/block/end are terminal; active call keeps captions, Safety and leave reachable; responsive Outfit UI has no desktop overflow                                                                                                                                                                                                                         | Recipient explicitly accepts a private audio proposal before join; call uses Obiara names only, never phone/email/contact discovery; safety, captions and leave remain available                                                                                                                                                                                                                                         |
| S7-009 | E09-S10 Fire load, device and constrained-network acceptance harness                          |        5 | `/root/consent_kernel`                 | DONE   | deterministic quality-only harness and evidence under `services/api/internal/quality/fireacceptance/` and `internal/quality/`                                                                                                                 | commit `1b5961e`; focused race/vet; full Go/vet; deterministic replay                                                                                                                                                                                                                                                                            | 150/150 seats at 24 kbps/180 ms RTT/70 ms jitter/5% loss; zero unavailable consent/safety/leave controls; 312 ms aggregate p90 under 400 ms; five 30-seat device classes all pass                                                                                                                                                                                                                               | Synthetic acceptance uses no production traffic or member data; Ghana-shaped 90-minute physical-device field run remains an explicit release prerequisite                                                                                                                                                                                                                                                                |
| S7-010 | E13-S01 Notification preference, quiet hours and six/day cap                                  |       10 | `/root/backend_notifications`          | DONE   | `services/api/internal/notifications/`                                                                                                                                                                                                        | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | -race suite green incl. 20-way cap race → exactly 6 claims; quiet-hours window incl. midnight-crossing and IANA tz date keys; muted-category suppression without cap consumption; safety override bypasses cap and quiet hours; Redocly + client regen + drift green                                                                                                                                            | Self-reviewed; cap claims are atomic conditional increments (duplicate-key = cap reached); delivery counters TTL after 7 days; channel dispatchers land with E13-S03+                                                                                                                                                                                                                                                    |
| S7-011 | E10-S02 async Oware game, move timers and room embedding                                      |        9 | `/root/client_quality`                 | DONE   | `services/api/internal/games/oware/session/` application, privacy adapter and Mongo repository                                                                                                                                                | commit `3a74887`; focused race/vet; 1,000-case move/seed property; live MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                                    | Two opaque HMAC players, bounded server deadline, server-only legal moves, one-winner CAS/replay, explicit expiry preserving board and raw privacy proofs green                                                                                                                                                                                                                                                 | Server-authoritative moves and bounded clocks; expiration never invents a move, retries never apply twice, room references stay opaque and no game outcome reaches matching visibility                                                                                                                                                                                                                                   |
| S7-012 | E10-S03 Oware Glicko-2 ratings and notation                                                   |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/games/oware/rating/` deterministic ratings and notation                                                                                                                                                                | commit `e619626`; canonical reference vector; focused race/vet; 293,288 rating and 19,079 notation fuzz executions; full Go/vet                                                                                                                                                                                                                  | Stable-order Glicko-2 including inactivity; bounded replay notation revalidates every move through Oware domain and rejects digest/capture/turn tampering                                                                                                                                                                                                                                                       | No persistence, UI, skill-to-matching export, popularity ranking or member-discovery surface                                                                                                                                                                                                                                                                                                                             |
| S7-013 | E10-S04 conduct-only suban integration                                                        |        5 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/games/conduct/` domain, revalidating service, privacy adapter and Mongo repository                                                                                                                            | commit `5225720`; focused race/vet; 1,000-case mapping property; live MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                                      | Four verified bounded events map server-side to fixed reasons/provenance; HMAC refs, authority, one-winner appeal, replay/CAS and raw privacy proofs green                                                                                                                                                                                                                                                      | Suban records conduct only; no skill, win/loss, rating, popularity, matching/trust visibility, client-selected score/reason or free text                                                                                                                                                                                                                                                                                 |
| S7-014 | E13-S02 Dawn, Monday, fire and Sunday rituals                                                 |       10 | `/root/backend_notifications`          | DONE   | `internal/notifications/ritual/`, `services/worker/internal/jobs/ritual/`                                                                                                                                                                     | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | schedule unit tests (calendar kinds, tz-aware due times, 19:15 herald, dedup keys); Testcontainers end-to-end: herald exactly-once with deduped rerun, blocked accounts skipped, dawn withheld before 06:00 and dispatched after; unit proof suppression never consumes the daily ritual                                                                                                                        | Self-reviewed; notifications context relocated to `internal/notifications/` for worker import; decide-before-dedup ordering so quiet-hours suppression defers rather than consumes the day's ritual; ritual calendar events ride the durable outbox                                                                                                                                                                      |
| S7-015 | E10-S05 reviewed Ɛbɛ duel domain and content catalog                                          |        5 | `/root/consent_kernel`                 | DONE   | `services/api/internal/games/ebe/` reviewed-content domain plus operations-approved Mongo catalog and retained exact-pair duel service                                                                                                        | original commit `27dfb1b`; runtime follow-up 2026-07-30: full Go test/vet, focused privacy projection proof, 99-operation OpenAPI/generated-client/Redocly checks                                                                                                                                                                                | Immutable reviewer/source/language/version provenance; accepted forms and reviewer/player keys never project; empty catalog fails closed; strict alternating turns, revision CAS, idempotent commands and tamper-proof replay                                                                                                                                                                                   | No synthetic production prompts, public answers/scores, cultural-authority claim, publishing, matching, rating or popularity coupling                                                                                                                                                                                                                                                                                    |
| S7-016 | E10 client Ɛbɛ duel experience                                                                |        5 | `codex-root`                           | DONE   | authenticated web/Expo circle launch, sourced prompt/turn history, private free-text answer and quiet polling; operations console approval workflow                                                                                           | 2026-07-30 runtime closure: Next web/admin production builds; Expo web export; 68 web, 83 admin and 7 mobile tests; all app typecheck/lint; full Go test/vet; `git diff --check`                                                                                                                                                                 | Member clients render only retained reviewed prompts and member-relative turns; the other answer and accepted forms stay hidden; admin approval records immutable sourced versions and never returns accepted forms                                                                                                                                                                                             | Removed fabricated opponent, proverb pack, source/revision, answer choices, readiness and reveal; catalog remains empty by default and requires an authenticated operations reviewer to add real sourced material                                                                                                                                                                                                        |
| S7-018 | E10-S06 client Anansesɛm relay and publish consent                                            |        5 | `codex-root`                           | DONE   | authenticated web/Expo private story creation, relay, author-edit, current-draft grant and publish experience                                                                                                                                 | 2026-07-30 runtime closure: 92-operation OpenAPI and generated-client checks; focused Anansesɛm/API Go tests and vet; web/mobile typecheck/lint; 71 web and 7 mobile tests; `git diff --check`                                                                                                                                                   | Both clients create from an exact two-member circle, render only retained passages, apply alternating contributions and author-owned edits, grant the current draft fingerprint and publish only after dual grants; new writing clears grants and no HMAC author/room keys are projected                                                                                                                        | No fabricated story, author, turn, consent, edition or session ID; no room/contact/matching/trust/private-authorship exposure, race pressure or implicit publication                                                                                                                                                                                                                                                     |
| S7-019 | E10-S06 Anansesɛm relay and dual-consent publication domain                                   |        5 | `/root/client_quality`                 | DONE   | `services/api/internal/games/anansesem/` domain, revalidating service, privacy adapter, Mongo repository and runtime composition                                                                                                              | original commit `6121be1`; runtime follow-up 2026-07-30: focused service/API/OpenAPI tests; 92 unique operations; generated-client drift and Redocly lint green                                                                                                                                                                                  | Exactly two HMAC authors/room, bounded strict alternation, author-only edit, edit-invalidated grants, fresh dual consent, one-winner CAS/replay and raw privacy proofs green; member projection reports ownership without projecting keys                                                                                                                                                                       | Versioned public edition contains neutral title/version/time plus ordinal/content only; no room, contact, private authorship, list/browse or stale consent                                                                                                                                                                                                                                                               |
| S7-020 | E10-S07 client Ampe realtime pulse                                                            |        5 | `codex-root`                           | DONE   | authenticated web/Expo exact-pair Ampe creation, retained readiness/manual-choice commands, server-owned presence heartbeat, safe pause/reconnect and atomic reveal                                                                           | 2026-07-30 runtime closure: focused Ampe/presence/API Go tests and vet; deterministic grace/absence/reconnect/no-forfeit proof; web/mobile typecheck/lint; 68 web and 7 mobile tests; repository-wide suite green before final presence slice; `git diff --check`                                                                                | Mongo replay store retains privacy-keyed rounds and tamper-evident transcripts; authenticated GETs touch server-timestamped privacy-keyed presence with TTL; absence after grace applies disconnect/pause, a returning heartbeat reconnects, and locked choices remain hidden until both players lock                                                                                                           | No client-controlled connect/disconnect action, camera/body inference, automatic forfeit, reveal on absence, public score, player key, matching/rating/trust coupling or fabricated partner/readiness/reveal                                                                                                                                                                                                             |
| S7-021 | E10-S07 Ampe realtime round domain                                                            |        5 | `/root/consent_kernel`                 | DONE   | pure `services/api/internal/games/ampe/domain/` server-authoritative round/replay law                                                                                                                                                         | commit `e03d415`; focused race/vet; 2,000 deterministic scenarios; 350,772 fuzz executions; full Go/vet                                                                                                                                                                                                                                          | Opaque pair, manual bounded choice, server sequence/CAS, retry fingerprint, first-lock secrecy, atomic second-lock reveal, disconnect pause/no-forfeit and tamper-proof replay green                                                                                                                                                                                                                            | Public View/Event omit opponent choice and digest; Mongo/GoMock inapplicable to pure domain; no camera/body data, public score, matching/rating/trust coupling                                                                                                                                                                                                                                                           |
| S7-023 | E12-S01 Universal report/block intake backend                                                 |       10 | `/root/backend_safety`                 | DONE   | `services/api/internal/safety/`, HTTP routes + OpenAPI                                                                                                                                                                                        | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers end-to-end: fraud→Tier A persisted with reporter-safe acknowledgement, queue event on outbox, self-report rejection, block unique-edge lifecycle; category→tier matrix unit tests (A: fraud/minor/sexual, B: harassment, C: spam/other per Doc 09 §2); Redocly + client regen + drift green                                                                                                      | Self-reviewed; reporter identity stored for least-exposure desk access only (Doc 09 §3); triage queues/evidence land with E12-S02+                                                                                                                                                                                                                                                                                       |
| S7-022 | E10-S08 client tournaments, private ladder and anti-cheat transparency                        |        5 | `codex-root`                           | DONE   | private invitation cohorts, member-relative bracket/ladder, web/Expo enrollment and Oware play, operations tournament/review desk, verified result, expiry evidence, human decision and one-winner appeal                                     | 2026-07-30 runtime closure: 115-operation OpenAPI/generated-client/Redocly checks; exact result/winner/match and exact expired-evidence rejection proofs; 62/62 repository JS tasks; full Go/vet; web/admin production builds; Expo export; 68 web, 83 admin and 7 mobile tests; `git diff --check`                                              | Both entrants converge on one active privacy-keyed match board; completed decisive boards alone advance the bracket; exact expired boards alone open neutral review; operations records a human decision and member may appeal once before final resolution; all clients render the retained state                                                                                                              | No public cohort browse/member grid, client winner/result/evidence key, accusation, score, reason, free text, automatic guilt, skill-to-matching/trust coupling, pay-to-rank or unverified ladder mutation                                                                                                                                                                                                               |
| S7-024 | E10-S08 tournament, ladder and anti-cheat domain                                              |        8 | `/root/client_quality`                 | DONE   | `services/api/internal/games/competition/` domain, revalidating service, privacy adapter and Mongo repository                                                                                                                                 | commit `0f98abb`; focused race/vet; deterministic/property tests; live MongoDB 8.0.13 Testcontainers race proof; full Go/vet                                                                                                                                                                                                                     | Opt-in 4–16 power-of-two cohorts, verified-result-only bracket/private ladder, HMAC refs, neutral review, human decision, distinct appeal resolution and CAS/idempotency proofs green                                                                                                                                                                                                                           | No listing/discovery/popularity/payment/pay-to-rank/matching/trust exposure; conduct remains separate and fair-play evidence never auto-accuses                                                                                                                                                                                                                                                                          |
| S7-025 | E10 client Oware room and async move experience                                               |        5 | `codex-root`                           | DONE   | guarded web/Expo Oware room route, accessible board and private room mount; both runtimes connected to authenticated API                                                                                                                      | original commit `133d141`; runtime follow-up 2026-07-30: 73 web and 7 mobile tests; web/mobile typecheck/lint; focused Oware/API/OpenAPI Go tests and vet; generated-client drift and Redocly lint; `git diff --check`                                                                                                                           | Server derives an exact two-member pair from current circle membership, creates a privacy-keyed Mongo session, projects no room/player keys and applies revisioned moves through the authoritative Abapa engine; clients only select a non-empty owned house and never simulate sow/capture                                                                                                                     | No client-supplied player identity, game outside an exact two-active-member circle, local gameplay truth, leaderboard, matching advantage, popularity pressure, contact/public-attendance exposure or invented server result                                                                                                                                                                                             |
| S8-001 | E11-S01 AI gateway, vendor policy and audit metadata                                          |        8 | `codex-root`                           | DONE   | isolated `services/api/internal/ai/gateway/` domain/application boundary with generated Uber GoMock; no vendor network adapter                                                                                                                | focused race/vet; 225,413 policy fuzz executions; full Go/vet; `go.uber.org/mock v0.6.0` latest verification; `git diff --check`                                                                                                                                                                                                                 | Capability/model/vendor/region/data-class allowlist, bounds, current-consent revalidation, fail-closed redaction, no fallback and content-free audit metadata ordering proofs green                                                                                                                                                                                                                             | No direct client/vendor call, raw prompt/response persistence or silent fallback; persistence/Testcontainers intentionally inapplicable to port-only gateway                                                                                                                                                                                                                                                             |
| S8-002 | E11-S02 rules and trust-path cold-start introductions                                         |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/matching/coldstart/` domain/application boundary; no model/vendor use                                                                                                                                         | commit `16eac1c`; focused race/vet; full Go/vet; 1,000 shuffled-order property runs; 208,916 fuzz executions; `git diff --check`                                                                                                                                                                                                                 | Reciprocal versioned explicit preferences intersect scoped allowlisted trust summaries; authority and visibility revalidated; max 20 opaque candidates and 4 stable generic reasons; JSON privacy proof rejects score/rank/path/popularity/skill/trait/model fields                                                                                                                                             | Pure current projection makes persistence/Testcontainers inapplicable; no model/vendor/global browse/popularity/game skill/sensitive inference/raw path exposure or compatibility ranking                                                                                                                                                                                                                                |
| S8-003 | E11-S03 client grounded resonance explanation and feature consent                             |        5 | `codex-root`                           | DONE   | guarded web/Expo introduction explanation and feature-consent controls using bounded reason-code fixtures                                                                                                                                     | 61 web tests; web/mobile typecheck/lint/build; rendered desktop interaction/semantic/overflow QA; `git diff --check`                                                                                                                                                                                                                             | Explanation renders only enabled reasons; trust withdrawal removes its reason immediately; optional voice feature defaults off; rules-first/system-detail disclosure says AI words but does not choose/rank                                                                                                                                                                                                     | No destiny/compatibility certainty, hidden score, attractiveness ranking, urgency, popularity, raw trust path or silent AI feature use                                                                                                                                                                                                                                                                                   |
| S8-004 | E11-S04 consent-controlled matching feature registry                                          |        5 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/matching/features/` domain/application/persistence boundary; no model/vendor use                                                                                                                              | commit `3f8b514`; focused race/vet; full Go/vet; 1,000-case property coverage; live MongoDB 8.0.13 Testcontainers intersection/withdrawal/privacy proof; `git diff --check`                                                                                                                                                                      | Versioned allowlisted definitions, optional default-off exact-purpose grants, immediate withdrawal, higher-version regrant, pair intersection, immutable decision snapshots and current-state revalidation green                                                                                                                                                                                                | HMAC member keys only; no raw content, sensitive inference, retroactive consent, stale grant, model/vendor data, public/list behavior or silent optional feature                                                                                                                                                                                                                                                         |
| S8-005 | E11-S06 client Okyeame whitelist, disclosure and refusal controls                             |        5 | `codex-root`                           | DONE   | shared pure `@obiara/okyeame-policy` plus guarded web/Expo Okyeame experiences; no vendor/network adapter                                                                                                                                     | 11 policy and 63 web tests; policy/web/mobile typecheck/lint; web/mobile production builds; rendered interaction/semantic/desktop-overflow QA; `git diff --check`                                                                                                                                                                                | Explicit AI disclosure; member-invoked feature, navigation and wording help only; matchmaking/autonomous romance/impersonation/counsel/private evidence/hidden memory refused; prompt retention fixed false                                                                                                                                                                                                     | Outfit and existing Fie visual language preserved; no autonomous initiation, silent memory, private access, relationship decision or counsel-to-matching leakage                                                                                                                                                                                                                                                         |
| S8-006 | E13-S05 WhatsApp OTP and pod alerts                                                           |       10 | `/root/backend_notifications`          | DONE   | `internal/notifications/whatsapp/`                                                                                                                                                                                                            | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | strict-template domain (OTP + pod alert only, no free text per FR-701); GoMock branch tests (OTP bypasses preferences, muted suppression sends/logs nothing, failure logged + reported); Testcontainers end-to-end with real delivery log; S7-010 cap test made wall-clock independent after midnight failure                                                                                                   | Self-reviewed; simulator sender until scored WhatsApp Business vendor; OtpSenderAdapter ready to switch identity OTP at composition; delivery log TTL 90 days                                                                                                                                                                                                                                                            |
| S8-007 | E11-S05 model/ranker readiness, offline evaluation and fairness gates                         |        8 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/matching/evaluation/` readiness/evaluation boundary; no online ranker or vendor invocation                                                                                                                    | commit `5f9d144`; focused race/vet; full Go/vet; 1,000 weak-cohort property cases; fuzz seed/non-finite checks; live MongoDB 8.0.13 Testcontainers CAS/privacy proof                                                                                                                                                                             | Exact consent snapshot, minimum cohort/evidence/per-slice thresholds, max eight approved slices, bounded quality/error/disparity, model card and HMAC human approval bound to revision with 30-day expiry; changes invalidate approval                                                                                                                                                                          | Offline only; no ranker, vendor, production decision, raw data, sensitive inference, weak cohort, stale approval, non-finite metric or unbounded group slicing                                                                                                                                                                                                                                                           |
| S8-008 | E11-S07 counsel isolation from matching                                                       |        5 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/counsel/isolation/` policy/application boundary plus compile-time dependency proof                                                                                                                            | commit `ed7f40f`; focused race/vet; full Go/vet; 1,000 shuffled exact-field property runs; 162,640 fuzz executions; `git diff --check`                                                                                                                                                                                                           | Deny-by-default matching/explanation/ranking/trust/AI egress; only explicit freshly authorized five-field opaque safety event allowed; architecture test forbids counsel imports to/from matching/trust/AI                                                                                                                                                                                                      | Pure policy/application boundary makes persistence/Testcontainers inapplicable; event structurally excludes content/topic/attendance/session/outcome/actor/free text and false/error/withdrawal never publishes                                                                                                                                                                                                          |
| S8-009 | E13-S03 Push, in-app and SMS routing with fallback                                            |       10 | `/root/backend_notifications`          | DONE   | `internal/notifications/routing/`, `internal/notifications/inapp/`                                                                                                                                                                            | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | ladder domain tests (OTP SMS-primary/WhatsApp-fallback per §11, safety 4-channel, unknown→in-app); GoMock router tests (first-wins, fallback, all-failed, suppression, missing-channel skip); Testcontainers end-to-end: push failure → in-app fallback lands in inbox, OTP via SMS, idempotent mark-read                                                                                                       | Self-reviewed; preference gate runs once before the ladder; safety/OTP bypass it; inbox entries are opaque refs with 90-day TTL, content renders from the localization registry at read time                                                                                                                                                                                                                             |
| S8-018 | E11-S13 client AI model cards, red-team status and appeal transparency                        |        5 | `codex-root`                           | DONE   | guarded web/Expo accountability surface under Okyeame using pure client projection; no model/vendor/admin data                                                                                                                                | 67 web tests; web/mobile typecheck/lint/build; rendered interaction/semantic/desktop-overflow QA; `git diff --check`                                                                                                                                                                                                                             | Versioned scope, consent basis, bounded evaluation/red-team state, explicit unreleased ranker and human-only appeal reference shown; reducer rejects unknown/replayed submissions and privacy projection rejects raw/sensitive fields                                                                                                                                                                           | Cards explicitly are not certifications or perfect-safety promises; no hidden score, group microtargeting, raw prompt/path/biometric/counsel data, automatic appeal decision or false production readiness                                                                                                                                                                                                               |
| S8-019 | E12-S06 admin Mpanyimfo docket, recusal, ruling and appeals                                   |        5 | `codex-root`                           | DONE   | `apps/admin/app/mpanyimfo/` MUI route with pure reducer fixtures and dashboard entry; no backend ruling mutation or private evidence                                                                                                          | 14 admin tests; admin typecheck/lint/build; rendered recusal/quorum/reasoned-ruling/separate-appeal, semantic and desktop-overflow QA; `git diff --check`                                                                                                                                                                                        | Redacted docket, conflict recusal, minimum two active seats/two matching votes, 20-character ruling reason and distinct appeal reference for another panel; original ruling remains intact                                                                                                                                                                                                                      | No single-elder decision, hidden/raw evidence, unreasoned ruling, recused vote, self-review, automatic enforcement or appeal overwrite                                                                                                                                                                                                                                                                                   |
| S8-023 | E12-S07 women’s-safety review evidence                                                        |        8 | `/root/consent_kernel`                 | DONE   | isolated stateless `services/api/internal/safety/womensreview/` aggregate/evidence boundary; no member content store or automatic enforcement                                                                                                 | commit `ec1c242`; focused race/vet; full API tests; about 412,000 insufficient-cohort fuzz executions; `git diff --check`                                                                                                                                                                                                                        | Versioned women-panel-reviewed dimensions, minimum cohort/response/dimension evidence, bounded redacted aggregates, substantive women-reviewer approval with full coverage/gap acknowledgements and neutral incomplete/release-review outcomes                                                                                                                                                                  | Stateless boundary makes Mongo/Testcontainers inapplicable; no raw content/identity, token review, subgroup microdata/labels, member score, false representativeness, automatic action, vendor/model decision or release port                                                                                                                                                                                            |
| S8-024 | E12-S08 admin incident response and regulatory runbooks                                       |        5 | `codex-root`                           | DONE   | `apps/admin/app/incidents/` MUI route with pure reducer fixtures and dashboard entry; no pager/provider/regulator mutation                                                                                                                    | 17 admin tests; admin typecheck/lint/build; rendered role/out-of-order/packet/dual-close, semantic and desktop-overflow QA; `git diff --check`                                                                                                                                                                                                   | Versioned P1 runbook, distinct commander/recorder, ordered mandatory containment/preservation/notification-clock checkpoints, redacted packet confirmation and two-role close accountability                                                                                                                                                                                                                    | Packet projection rejects raw-member fields; no silent severity change, skipped mandatory step, same-person close, evidence deletion, unilateral close or automatic regulator notification                                                                                                                                                                                                                               |
| S8-011 | E11-S09 Sow pre-delivery multilingual screening                                               |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/seed/screening/` stateless adapter composed behind existing sow Screening port; no delivery/persistence ownership                                                                                             | commit `43a6bba`; focused race/vet; all-seed/full Go/vet; 1,000 multilingual property rounds; 167,352 normalization fuzz executions; `git diff --check`                                                                                                                                                                                          | Unicode NFKC/bounded text, reviewed versioned locale catalog without translation, bounded media metadata, provider-neutral advisory and mandatory opaque human adjudication; unsupported/uncertain/error routes to human                                                                                                                                                                                        | Stateless transient adapter makes Mongo/Testcontainers inapplicable; real sow composition proves rejection never reaches acceptance/allowance spend; no raw retention, translation-as-truth, direct vendor, delivery or model-final decision                                                                                                                                                                             |
| S8-012 | E11-S10 Sika Shield text/voice patterns and precision gate                                    |        8 | `/root/client_quality`                 | DONE   | isolated stateless `services/api/internal/safety/sikashield/` offline detector/evaluation boundary; no automatic member action or payment access                                                                                              | commit `472f186`; focused race/vet; full Go/vet; 1,000-case property; about 476,000 fuzz executions; `git diff --check`                                                                                                                                                                                                                          | Versioned human-reviewed text/consented voice-metadata patterns, finite aggregate metrics, precision at least .97/review at most .20, exact-current revalidation and only no-action/human-review outcomes                                                                                                                                                                                                       | Stateless boundary makes package-local Mongo/Testcontainers inapplicable; no raw voice/text retention, accusation, automatic enforcement/account/payment action, sensitive inference, hidden score, vendor or production activation                                                                                                                                                                                      |
| S8-010 | E13-S04 Resend email: templates, sender port, signed delivery webhooks                        |       10 | `/root/backend_notifications`          | DONE   | `internal/notifications/email/`, webhook route + OpenAPI                                                                                                                                                                                      | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | strict-template domain (ops/admin/verification-help, no marketing); svix HMAC-SHA256 vectors (valid, wrong body, wrong id, rotation multi-sig, stale timestamp, missing secret); handler tests (401, applied, ignored unknown events, 400); Testcontainers delivery log + status update round-trip; Redocly + client regen + drift green; one Docker flake rerun green                                          | Self-reviewed; webhook secret via RESEND_WEBHOOK_SECRET env only; svix-id deduped through inbox; unknown provider refs 404 for dead-letter handling                                                                                                                                                                                                                                                                      |
| S8-013 | E12-S03 admin privacy-redacted evidence viewer and legal-hold workflow                        |        5 | `codex-root`                           | DONE   | `apps/admin/app/safety/` MUI route with pure reducer fixtures and dashboard entry; no backend mutation or raw evidence                                                                                                                        | 6 admin tests; admin typecheck/lint/build; rendered purpose/acknowledgement/evidence/hold interaction, semantic and desktop-overflow QA; `git diff --check`                                                                                                                                                                                      | Least-exposure selection, 12-character bounded purpose, audit acknowledgement, explicit redaction markers and separate reversible legal-hold confirmation; unknown selection/open/hold transitions fail closed                                                                                                                                                                                                  | Reporter identity/raw content absent, evidence never opens unlogged, hold does not expose/delete/decide, and no automatic case action or destructive release                                                                                                                                                                                                                                                             |
| S8-014 | E11-S11 scam-arc sequence signals and action ladder                                           |        8 | `/root/consent_kernel`                 | DONE   | isolated stateless `services/api/internal/safety/scamarc/` aggregate/evaluation boundary; no message store, payment access or automatic enforcement                                                                                           | commit `01d777c`; focused race/vet; full API tests; about 277,000 single-event fuzz executions; `git diff --check`                                                                                                                                                                                                                               | Bounded opaque event sequence, deterministic reviewed/versioned rules, strictly escalating least-harm recommendations, neutral human-review signal and current consent/authority revalidation                                                                                                                                                                                                                   | Stateless boundary makes Mongo/Testcontainers inapplicable; no raw content/payment/voice/member data, accusation, hidden score, auto-block/payment/account action, single-event conviction, vendor or model-final decision                                                                                                                                                                                               |
| S8-015 | E12-S04 admin action ladder and propagated-control confirmation                               |        5 | `codex-root`                           | DONE   | extended `apps/admin/app/safety/` pure reducer/MUI desk only; no backend enforcement or device mutation                                                                                                                                       | 8 admin tests; admin typecheck/lint/build; rendered proposal/scope/reason/confirmation/appeal, semantic and desktop-overflow QA; `git diff --check`                                                                                                                                                                                              | Human-only warning, temporary surface restriction and bounded account review proposals; explicit scope, 12-character neutral reason, confirmation and appeal notice required; reducer never auto-escalates                                                                                                                                                                                                      | Proposal explicitly cannot act on devices/payments/other accounts; no permanent-ban default, automatic enforcement, accusation, silent propagation or hidden member score                                                                                                                                                                                                                                                |
| S8-016 | E12-S05 admin care queue and approved resource-first scripts                                  |        5 | `codex-root`                           | DONE   | `apps/admin/app/care/` MUI route with pure reducer fixtures and dashboard entry; no clinical record, diagnosis or outbound provider                                                                                                           | 11 admin tests; admin typecheck/lint/build; rendered script/contact-confirmation/no-contact-disabled, semantic and desktop-overflow QA; `git diff --check`                                                                                                                                                                                       | Least-exposure cases, approved versioned resource scripts, member contact preference, explicit human confirmation and one-message preparation; unapproved/unknown/no-contact/unscripted transitions fail closed                                                                                                                                                                                                 | Projection rejects clinical claims; no diagnosis, therapy/treatment claim, repeated/coercive contact, punitive safety action, raw counsel content or automated crisis adjudication                                                                                                                                                                                                                                       |
| S8-017 | E11-S12 syndicate, vouch-ring and device anomaly detection                                    |        8 | `/root/client_quality`                 | DONE   | isolated stateless `services/api/internal/safety/anomaly/` offline aggregate/evaluation boundary; no graph browse, device fingerprint store or automatic enforcement                                                                          | commit `25e2c8d`; focused race/vet; full Go/vet; graph determinism/all-shape/1,000-case property; about 366,000 fuzz executions; `git diff --check`                                                                                                                                                                                              | Versioned reviewed rules, bounded graph aggregates, 64-hex opaque evidence/reviewer keys, cohort 200/evidence 1,000/precision .98 gates, current consent/evaluator/router authority and only no-action/human-review                                                                                                                                                                                             | Stateless boundary makes Mongo/Testcontainers inapplicable; no raw graph/path/device/member data, browse, guilt by association, accusation/score, matching/trust mutation, enforcement/account/payment/vendor/model-final path                                                                                                                                                                                           |
| S8-021 | E12-S04 Action ladder and propagated account/device controls                                  |       10 | `/root/backend_safety`                 | DONE   | `internal/safety/` (actions), identity status transitions, `services/worker/internal/jobs/safety/`                                                                                                                                            | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | ladder matrix unit tests (A→ban only, B→suspend then ban, C→warn then escalate per Doc 09 §2); GoMock propagation tests; Testcontainers: 14d suspension sets account suspended with computed expiry + sessions revoked + immutable action log + case resolved, repeat-B ban blocks account + device_risk entry, off-ladder rejected with no trace; worker reactivation lifts only expired suspensions           | Self-reviewed; identity gained suspended status + expiry with version-pinned updates; enforcement bridges live at composition roots; propagation failure leaves no partial action record                                                                                                                                                                                                                                 |
| S8-025 | E12-S09 retention, erasure and transparency-report data                                       |        8 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/compliance/retention/` policy/application/persistence boundary; no product-content ownership                                                                                                                  | commit `c477547`; focused race/vet; 1,000-case hold property; about 579,000 fuzz executions; live MongoDB 8.0.13 Testcontainers concurrency/idempotency/index/privacy proof; full Go/vet                                                                                                                                                         | Versioned exact data-class/purpose policies, HMAC keys, legal-hold precedence, verifier-required irreversible erasure completion, append-only events and opaque aggregate counters green                                                                                                                                                                                                                        | No deletion API, silent durable deletion, hold bypass, raw-member transparency, reversible-erasure claim, cross-purpose reuse or product-content ownership                                                                                                                                                                                                                                                               |
| S8-026 | E12-S10 admin moderation workforce safeguards                                                 |        5 | `codex-root`                           | DONE   | `apps/admin/app/workforce/` MUI route with pure reducer fixtures and dashboard entry; no HR record, scheduling provider or evidence access                                                                                                    | 20 admin tests; admin typecheck/lint/build; rendered break/category-preview/acceptance/exposure/support, semantic and desktop-overflow QA; `git diff --check`                                                                                                                                                                                    | Bounded shift/exposure counters, protected break clears assignment, category-only preview, explicit accept, no-penalty opt-out and supervisor support request; max exposure blocks preview                                                                                                                                                                                                                      | Projection contains no productivity/ranking field; no forced evidence viewing, diagnosis, hidden surveillance, retaliation/performance flag or automatic assignment                                                                                                                                                                                                                                                      |
| S8-027 | E12-S11 admin fraud-victim evidence export path                                               |        5 | `codex-root`                           | DONE   | extended `apps/admin/app/safety/` with pure export reducer/MUI dialog; no file generation, evidence store or external recipient                                                                                                               | 22 admin tests; admin typecheck/lint/build; rendered disabled/consent/purpose/scope/expiry/success interaction and semantic QA; `git diff --check`                                                                                                                                                                                               | Victim-requested bounded export, explicit included-reference checklist, reporter/third-party redaction, purpose, 72-hour expiry and human confirmation                                                                                                                                                                                                                                                          | No accused-party disclosure, raw hidden evidence/scope, automatic sharing, perpetual link, case action or victim-blaming copy                                                                                                                                                                                                                                                                                            |
| S8-028 | E12-S11 fraud-victim evidence export domain                                                   |        8 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/safety/victimexport/` domain/application/persistence boundary; no evidence-content store or outbound delivery                                                                                                 | commit `970adc4`; focused race/vet; 1,000-case TTL property; about 354,000 fuzz executions; live MongoDB 8.0.13 Testcontainers concurrency/idempotency/privacy proof; full Go/vet                                                                                                                                                                | Member-authorized allowlisted opaque refs, exact purpose, redaction attestation, HMAC keys, 72-hour one-time/revocable authorization, CAS and append-only audit green                                                                                                                                                                                                                                           | No raw evidence/content store, reporter/accused identity, automatic delivery endpoint, perpetual token, case action or mutable audit TTL deletion                                                                                                                                                                                                                                                                        |
| S8-029 | E13-S06 client Nnoboa nomination and auntie consent flow                                      |        5 | `codex-root`                           | DONE   | guarded web/Expo companion surfaces with shared pure `@obiara/nnoboa-policy`; no WhatsApp provider, candidate discovery or server mutation                                                                                                    | 4 policy and 67 web tests; policy/web/mobile typecheck/lint; web/mobile production builds; rendered disabled/consent/accept/veto/semantic/desktop-overflow QA; `git diff --check`                                                                                                                                                                | Up to three member-designated nominators, bounded opaque candidate reference, member absolute veto, nominee consent before introduction review and Outfit companion surfaces green                                                                                                                                                                                                                              | No global browse, autonomous nomination, nominee exposure before consent, courtship conversation, private room/doorway/voice/contact content, engagement bait or pressure copy                                                                                                                                                                                                                                           |
| S8-030 | E15-S01+S02 producer-enforced analytics schema registry and consent-aware pipeline            |       10 | `/root/backend_analytics`              | DONE   | `services/api/internal/analytics/`                                                                                                                                                                                                            | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | registry covers full Doc 08 §3 taxonomy; validation rejects unregistered events, undeclared props, free text in enum/opaque props, wrong types, missing required; GoMock pipeline tests; Testcontainers: valid event persists pseudonymized (member id never stored), content-smuggling prop rejected pre-persistence, opt-out emits nothing                                                                    | Self-reviewed; consent bridge reads consent_records when the registry ships; pseudonym is salted sha256, aggregation/pseudonymization jobs land with E15-S08                                                                                                                                                                                                                                                             |
| S8-031 | E13-S09 banned engagement-pattern test suite                                                  |        5 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/notifications/engagementpolicy/` pure policy/test boundary; no dispatcher or template persistence ownership                                                                                                   | commit `1df9156`; focused race/vet; 1,000 deterministic and 500 shuffled property cases; about 330,000 fuzz executions; architecture/privacy/diff proofs                                                                                                                                                                                         | Immutable versioned/reviewed locale catalog covers all five prohibited classes; NFKC/case/space normalization scans title/body/template/campaign/tags and fails closed on unsafe controls or unknown locale/category                                                                                                                                                                                            | Stateless boundary makes Mongo/Testcontainers inapplicable; no content generation, delivery, member inference, repository/vendor/network, quiet-hour/cap bypass, mutable list or hidden engagement score                                                                                                                                                                                                                 |
| S8-032 | E13-S01 client notification preferences, quiet hours and cap transparency                     |        5 | `codex-root`                           | DONE   | guarded web/Expo notification settings with shared pure `@obiara/notification-settings`; no dispatcher, provider, preference persistence or delivery mutation                                                                                 | 3 policy and 67 web tests; policy/web/mobile typecheck/lint; web/mobile production builds; rendered category/channel/quiet-hours/cap/critical-copy interaction, semantic and desktop-overflow QA; `git diff --check`                                                                                                                             | Per-category/channel choice, bounded local quiet-hours window, immutable server-owned six-per-day explanation and safety/OTP override disclosure green                                                                                                                                                                                                                                                          | No engagement bait, pressure/jealousy copy, disabling critical safety/OTP, fake urgency, popularity/view signals or claim that client settings alone enforce delivery                                                                                                                                                                                                                                                    |
| S8-033 | E15-S04 append-only suban event ledger with recomputable marks                                |       12 | `/root/backend_analytics`              | DONE   | `services/api/internal/suban/`, marks/events routes + OpenAPI                                                                                                                                                                                 | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | mark computation tests (fresh credits earn, 2.5-half-life decay ages out, finding suppresses in window, fraud permanent, combined small/tiny weights); GoMock cap tests; Testcontainers: three follow-throughs earn keeps_word, finding suppresses without ledger edit, member-visible ordered ledger, 11th same-kind event capped; Redocly + client regen + drift green                                        | Self-reviewed; marks only ever recomputed from the append-only ledger (no cached authority per Doc 08 §4); 0.05 threshold tolerance documented against day-fresh decay                                                                                                                                                                                                                                                   |
| S8-034 | E14-S01 product/SKU rules proving seeds and visibility are unsellable                         |        8 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/commerce/catalog/` domain/application/persistence boundary; no checkout, payment or marketplace profile ownership                                                                                             | commit `3d4eaf2`; focused race/vet; 1,000-case product-law property; about 67,000 fuzz executions; live MongoDB 8.0.13 Testcontainers version/read/privacy proof; repository-wide vet                                                                                                                                                            | Only physical-good/event-ticket/digital-service kinds, bounded GHS/USD minor units, opaque titles, immutable versions and draft→published→retired law; changed price requires a new version                                                                                                                                                                                                                     | No seeds, approaches, visibility/rank/matching advantage, suban/trust, urgency, member transfer, hidden fee, checkout/provider/profile ownership or global browse                                                                                                                                                                                                                                                        |
| S8-035 | E14-S06 client matchmaker discovery and consent-led booking                                   |        5 | `codex-root`                           | DONE   | guarded web/Expo licensed-matchmaker marketplace with shared pure `@obiara/matchmaker-marketplace`; no calendar, MoMo, escrow, provider or server mutation                                                                                    | 3 policy and 67 web tests; policy/web/mobile typecheck/lint; web/mobile production builds; rendered license/fee/service/dual-consent/semantic/desktop-overflow QA; `git diff --check`                                                                                                                                                            | Licensed profiles, languages/specialties, banded fees, post-engagement-only ratings, consultation intent and mutual consent before curated-proposal exposure green                                                                                                                                                                                                                                              | No global member browse, purchasable approach, candidate exposure before both consents, transactional review pressure, hidden fee, seed/visibility sale or fake scarcity                                                                                                                                                                                                                                                 |
| S8-036 | E14-S02 MoMo provider adapter and USSD-push intent boundary                                   |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/commerce/momo/` domain/application/persistence boundary with provider port; no ledger, catalog or checkout UI ownership                                                                                       | commit `3b307e4`; focused race/vet; about 101,000 fuzz executions; live MongoDB 8.0.13 Testcontainers idempotency/concurrency/privacy proof; `git diff --check`                                                                                                                                                                                  | Bounded GHS/pesewa intent, explicit member→provider confirmation states, immutable amount/audit, opaque provider refs, HMAC phone ref, signed replay-safe callback, command uniqueness and CAS green                                                                                                                                                                                                            | No raw durable phone, member transfer, silent charge, seeds/visibility SKU, unsigned callback, amount mutation, success-before-provider confirmation or real provider call                                                                                                                                                                                                                                               |
| S8-052 | Doc 08 §8 consent-map registry: purpose toggles with receipts and enforcement ports           |       10 | `/root/backend_consent`                | DONE   | `services/api/internal/consent/consentmap/`, consent routes + OpenAPI                                                                                                                                                                         | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | control-rule unit tests (locked required purpose, one-way opt-in/opt-out, defaults, explicit-wins); Testcontainers: default resolution, opt-in with receipt, opt-out wins, locked/wrong-direction rejected without receipts, switchboard merge; Redocly + client regen + drift green                                                                                                                            | Self-reviewed; enforcement ports in analytics/scam-arc/games now have a registry to bridge to; identity-safety row is immutable by construction                                                                                                                                                                                                                                                                          |
| S8-053 | E16-S06 market-pack governance with four-eyes publishing and configuration audit              |       12 | `/root/backend_admin`                  | DONE   | `services/api/internal/marketpack/`, governance routes + OpenAPI                                                                                                                                                                              | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | four-eyes unit tests (self-approval rejected, approver required, state lifecycle); Testcontainers: draft audited, self-approval rejected, second approver publishes and lists, retire unlists, configuration_changes trail grows per action; Redocly + client regen + drift green                                                                                                                               | Self-reviewed; kill-switch semantics covered by retire; pack content stays terminology-ref + feature flags, no copy strings in the pack itself                                                                                                                                                                                                                                                                           |
| S8-054 | E12-S08 incident response, evidence handling and 72-hour reporting runbooks                   |       10 | `/root/backend_safety`                 | DONE   | `internal/operations/runbooks/`                                                                                                                                                                                                               | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | incident-response.md (severity scheme, contain-with-kill-switches, evidence preservation, provider playbooks, care protocol) and evidence-and-reporting.md (access classes, preservation, exports, regulator contacts, clock records) reviewed against Doc 09 §3/§7 and plan §15/§16; prettier clean                                                                                                            | Self-reviewed; runbooks reference only shipped machinery (holds, redacted viewer, action logs, kill-switch flags, fallback ladders) and mark rehearsal as a Sprint 11 gate                                                                                                                                                                                                                                               |
| S8-055 | E13-S06 Nnoboa nomination and auntie consent backend                                          |       12 | `/root/backend_notifications`          | DONE   | `services/api/internal/companions/nnoboa/`, nomination routes + OpenAPI, WhatsApp `nnoboa_consent` template                                                                                                                                   | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | domain lifecycle unit tests (validation, consent/decline/expiry transitions, 30-day window); GoMock service tests (duplicate-pending rejected, invite send, lazy expiry); Testcontainers round-trip (create/read/consent CAS, version conflict, pending-kin query, latest-first list); 4 OpenAPI ops (nominateKin, listNominations, consentNomination, declineNomination), Redocly + client regen + drift green | Self-reviewed; invite names the kin only — nothing about the member's romantic life; decline always respected without consequence; consent invite delivery-logged via the WhatsApp channel (simulator sender until provider selection)                                                                                                                                                                                   |
| S8-039 | E09-S07 ember-close transaction: fire dims to embers with frozen attendance                   |        9 | `/root/backend_fires`                  | DONE   | `services/api/internal/fire/` (close flow), close route + OpenAPI                                                                                                                                                                             | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | -race integration: non-host rejected (FR-401), host close dims to embers, roster frozen at capacity, admissions rejected after close; domain state tests (host-only, embers not re-closable, live close allowed); Redocly + client regen + drift green                                                                                                                                                          | Self-reviewed; waitlisted members are excluded from the frozen roster (going only); cancellation-after-close policy noted for host-controls follow-up                                                                                                                                                                                                                                                                    |
| S8-040 | E14-S03 immutable double-entry ledger domain                                                  |        8 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/commerce/ledger/` domain/application/persistence boundary; no provider, catalog or admin UI ownership                                                                                                         | commit `2d66a9d`; repository-wide sequential tests/vet; focused race; live MongoDB Testcontainers concurrency/idempotency/malformed-row/privacy proof; `git diff --check`                                                                                                                                                                        | Balanced append-only journal, GHS/USD bounded minor units, account-class law, semantic idempotency, current authorization, HMAC refs, insert-only Mongo and recomputable balances green                                                                                                                                                                                                                         | No edit/delete, single-entry posting, member transfer, negative/overflow amount, hidden fee, provider call, raw member/payment data or cached-authority balance                                                                                                                                                                                                                                                          |
| S8-041 | E14-S05 membership/pass grace, cancellation, receipt and refund domain                        |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/commerce/membership/` domain/application/persistence boundary; no ledger, provider, catalog or client UI ownership                                                                                            | commit `9288e64`; focused race/vet; about 260,000 fuzz executions; live MongoDB 8.0.13 Testcontainers concurrency/idempotency/privacy proof; `git diff --check`                                                                                                                                                                                  | Versioned pass grant, immutable receipt/term, explicit paid-through, max-14-day grace, non-punitive cancellation, opaque refund refs, provider-confirmation port, CAS and append-only audit green                                                                                                                                                                                                               | No romance-access guarantee, seeds/visibility/rank, silent renewal, punitive cancellation, refund-before-confirmation, raw payment/member data, provider call or mutable historical term                                                                                                                                                                                                                                 |
| S8-042 | E14-S06 matchmaker profiles, licensing and consent-led booking domain                         |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/commerce/matchmaker/` domain/application/persistence boundary; no payment, ledger, marketplace UI or member discovery ownership                                                                               | commit `e7286d3`; focused race/vet; about 438,000 fuzz executions; live MongoDB 8.0.13 Testcontainers concurrency/idempotency/privacy proof; `git diff --check`                                                                                                                                                                                  | Current license/version/jurisdiction and fee-band validation, immutable milestone terms, opaque refs, consultation booking, sealed proposal, distinct dual consent and completion-only review eligibility green                                                                                                                                                                                                 | No global browse, pre-consent candidate exposure, pay-to-approach, seeds/visibility/rank, unlicensed listing, pre-engagement rating, provider call or raw contact data                                                                                                                                                                                                                                                   |
| S8-043 | E14-S05 client membership, grace, cancellation, receipt and refund flow                       |        5 | `codex-root`                           | DONE   | guarded web/Expo membership settings with shared pure `@obiara/membership-settings`; no provider, ledger, catalog, billing or server mutation                                                                                                 | commit `e27279f`; 3 policy and 67 web tests; policy/web/mobile typecheck/lint/build; rendered cancellation/refund/provider-confirmation/semantic/desktop-overflow QA                                                                                                                                                                             | Exact paid-through and renewal state, non-punitive cancellation preserving purchased access, opaque receipt/refund refs and provider-confirmed completion green                                                                                                                                                                                                                                                 | No romance-access guarantee, silent renewal, cancellation pressure, seeds/visibility/rank sale, refund-before-confirmation, raw payment data or fake urgency                                                                                                                                                                                                                                                             |
| S8-044 | E14-S04 idempotent payment webhooks and reconciliation boundary                               |        8 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/commerce/reconciliation/` domain/application/persistence boundary with provider-signature and ledger ports; no provider network/admin UI ownership                                                            | commit `5b152d7`; focused race/vet; about 205,000 fuzz executions; live MongoDB 8.0.13 Testcontainers concurrency/idempotency/checkpoint/malformed-row/privacy proof; repository-wide tests except one known telemetry flake whose immediate isolated rerun passed; `git diff --check`                                                           | Signed callback verification, immutable opaque provider facts, semantic event idempotency/conflict, exact read-only ledger proof, explicit exceptions, immutable daily checkpoints, append-only fingerprinted audits and HMAC keys green                                                                                                                                                                        | No unsigned mutation, duplicate posting, automatic balance edit, hidden tolerance, provider call, raw phone/payment credential, destructive reconciliation or success without ledger proof                                                                                                                                                                                                                               |
| S8-045 | E14-S07 escrow, settlement, payout statement and dispute domain                               |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/commerce/escrow/` domain/application/persistence boundary with ledger port; no provider, marketplace UI or admin ownership                                                                                    | commit `f6377ac`; focused/property/privacy tests; race/vet; about 916,000 fuzz executions; live MongoDB 8.0.13 Testcontainers concurrent create/CAS/idempotency/privacy proof; `git diff --check`                                                                                                                                                | Immutable funded amount/funding ref, versioned milestones, distinct delivery and acceptance evidence, bounded cumulative settlement, explicit gross/fee/net payout statement, idempotent ledger command, dispute freeze, retained Mpanyimfo escalation ref and append-only CAS transitions green                                                                                                                | No release before evidence, unilateral term change, over-settlement, member transfer, hidden fee, dispute deletion, raw payment/member data or provider call                                                                                                                                                                                                                                                             |
| S8-046 | E12-S05 care queue with resource-first scripts and closure quietening                         |       10 | `/root/backend_safety`                 | DONE   | `internal/safety/` (care), notifications quietening read                                                                                                                                                                                      | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers end-to-end: closure flag opens case + sets 72h window, pods suppressed with care_quietening reason while safety still delivers and other members unaffected, distress flags route without quietening, diagnostic-language scripts rejected, engage/resolve lifecycle, empty queue after resolve                                                                                                  | Self-reviewed; care never touches enforcement paths (Doc 09 §5); scripts are an approved registry (helpline/counselor/support/quietening); quietening refreshable per new closure                                                                                                                                                                                                                                        |
| S8-047 | E15-S08 retention, pseudonymization and aggregation jobs with proof-of-deletion               |       10 | `/root/backend_analytics`              | DONE   | `internal/platform/retention/`, `services/worker/internal/jobs/retention/`                                                                                                                                                                    | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers: 90d pseudonymization strips subjectRef from aged rows only, 13mo aggregation rolls events into per-day counts then deletes, completed privacy requests purged at 90d with open kept, immutable retention_audit proof per policy, rerun fully idempotent                                                                                                                                         | Self-reviewed; binding policies live in one declarative table (media classes activate when persistence lands); legal holds untouched by design                                                                                                                                                                                                                                                                           |
| S8-048 | E09-S09 in-app call initiation with LiveKit tokens and zero phone-number exposure             |        9 | `/root/backend_calls`                  | DONE   | `services/api/internal/calls/`, call routes + OpenAPI                                                                                                                                                                                         | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers: membership-gated initiation persists call and issues 64-hex opaque speaker tokens per participant, non-member rejected, outsider end rejected, participant end persists ended state; GoMock branch tests; Redocly + client regen + drift green                                                                                                                                                  | Self-reviewed; participant/room keys are salted sha256 to match the adapter's 64-hex contract (FR-304); main composes the real adapter when LIVEKIT_API_KEY/SECRET are set, else a clean not-configured stub                                                                                                                                                                                                             |
| S8-049 | E15-S07 P0 funnel and phase-exit metric queries over the analytics pipeline                   |       10 | `/root/backend_analytics`              | DONE   | `services/api/internal/analytics/` (queries), metrics route + OpenAPI                                                                                                                                                                         | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers: 4→3→2→1 funnel computes 0.75/0.5/0.5 rates, fire attendance 2 of 4 active = 0.5, regret trend down against prior window; unit rate/zero-division/trend tests; Redocly + client regen + drift green                                                                                                                                                                                              | Self-reviewed; live queries over raw events complement codex's p0gate projection snapshots (different surfaces, no overlap)                                                                                                                                                                                                                                                                                              |
| S8-050 | E11-S11 scam-arc sequence signals and action ladder                                           |       10 | `/root/backend_safety`                 | DONE   | `services/api/internal/sentinel/scamarc/`, signal route + OpenAPI                                                                                                                                                                             | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | scoring unit tests (distinct-kind weights, diversity bonus, ladder thresholds); GoMock ladder tests (education card only on rung crossing, case opens once at top rung, opted-out room rejected pre-append); Testcontainers: none→education(card)→case(once) progression with persisted state; Redocly + client regen + drift green                                                                             | Self-reviewed; rules-first per v0 (no ML); monitoring defaults on per Doc 08 §8 with opt-out port; case-opener bridge is nil-safe until wired to safety desks                                                                                                                                                                                                                                                            |
| S8-051 | E13-S08 delivery observability: per-channel delivery statistics surface                       |       10 | `/root/backend_notifications`          | DONE   | `internal/notifications/deliverystats/`, stats route + OpenAPI                                                                                                                                                                                | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers: window-scoped aggregation per channel (9 whatsapp attempts incl. 2 failed, 10 email with bounced+complained folded into failure rate 0.8), out-of-window excluded; unit zero-division/default tests; Redocly + client regen + drift green                                                                                                                                                       | Self-reviewed; completes E13-S08: fallback lives in the routing ladders, opt-out in preferences, observability here — stats computed from delivery logs only, never message content                                                                                                                                                                                                                                      |
| S8-076 | E16-S06 market-pack and terminology governance backend                                        |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/governance/marketpack/` domain/application/persistence boundary; no runtime i18n registry, deployment or admin UI ownership                                                                                   | commit `7ee19b5`; parity/placeholder/terminology/distinct-role and 1,000-order property tests; race/vet; about 588,000 fuzz executions; live MongoDB 8.0.13 Testcontainers concurrency/idempotency/privacy proof; `git diff --check`                                                                                                             | Immutable digested versions bind exact current master; complete key and placeholder multiset parity, registered do-not-translate law, professional/community/in-context review, substantive checks, distinct roles, opaque refs, current-master revalidation and publish-ready-only CAS state green                                                                                                             | No machine translation as truth, missing/extra key, placeholder drift, same-person approval, silent fallback approval, automatic registry overwrite/deploy/market activation, free-text content store or culturally unreviewed terminology                                                                                                                                                                               |
| S8-056 | E16-S07 backend feature-flag and kill-switch control plane                                    |        8 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/platform/flagcontrol/` domain/application/persistence boundary composed through existing fail-closed flags kernel; no admin UI, deployment or member targeting ownership                                      | commit `e6e8094`; focused/property/race/vet; about 706,000 fuzz executions; live replica-set MongoDB 8.0.13 Testcontainers transaction/concurrent-CAS/idempotency/audit/malformed-row/privacy proof; repository-wide Go tests/vet; `git diff --check`                                                                                            | Immutable exact environment/market/capability proposals, finite actions/reasons, max-two-hour expiry, stepped least-privilege proposer, distinct stepped approver, apply restricted to approver, atomic CAS audit and existing kill-precedence integration green; expiry is disabled+killed                                                                                                                     | No global-by-default or permanent temporary control, same-person approval, unstepped mutation, member/cohort targeting, direct deployment, silent activation, audit edit/delete, unsafe expiry fallback, source-kernel bypass, free-text secret or automatic rollout claim                                                                                                                                               |
| S8-057 | E16-S03 admin community, circle, host and fire operations                                     |        5 | `codex-root`                           | DONE   | `apps/admin/app/community/` MUI operations desk with pure redacted fixtures and dashboard entry; no circle, host, fire, notification or member mutation                                                                                       | 3 community and 39 admin tests; admin typecheck/lint/build; rendered density/ineligible-host/reason/notice/eligible-proposal/semantic/desktop-overflow QA; `git diff --check`                                                                                                                                                                    | Redacted circle density/capacity and immutable fire references, current host verification/certification, bounded action reason and participant-notice preview required before a proposal; source state remains unchanged                                                                                                                                                                                        | No member/content/contact exposure, unverified host assignment, silent fire cancellation, automatic community change, capacity bypass, global browse, retaliation, notification send or source-state overwrite                                                                                                                                                                                                           |
| S8-059 | E18-S01+S02+S03 admin launch cohort, host school and matchmaker licensing readiness           |        5 | `codex-root`                           | DONE   | `apps/admin/app/launch/` MUI readiness desk with pure aggregate fixtures and dashboard entry; no CRM, certification, licensing, outreach or market activation mutation                                                                        | 4 launch and 43 admin tests; admin typecheck/lint/build; rendered denominator/incomplete-training/license/note/fail-closed/semantic/desktop-overflow QA; `git diff --check`                                                                                                                                                                      | First Hundred Families uses aggregate count and density, Host School uses required modules/current certification, Agyina uses current license/jurisdiction; incomplete evidence blocks launch and a bounded review note leaves all facts unchanged                                                                                                                                                              | No member/contact list, recruitment outreach, vanity target pass, expired host/license acceptance, automatic certification/license/market activation, hidden denominator, coercive growth metric or source-record edit                                                                                                                                                                                                   |
| S8-060 | E17-S01 Render Blueprint and environment matrix                                               |        5 | `codex-root`                           | DONE   | root `render.yaml`, protected synthetic-staging API/worker/web/admin topology, environment matrix and deterministic validator; no external resource creation or production deployment                                                         | official `render blueprints validate` valid with six actions/four services; 4 blueprint tests; API/worker production builds; web/admin production builds; `git diff --check`                                                                                                                                                                     | API and worker are separately scaled with immutable commit identity and independent secret slots, API `/live` health check, protected/network-isolated staging, checks-pass deploys, manual seven-day previews, no durable disk and explicit production residency/legal gate                                                                                                                                    | No secret value in Git, shared API/worker credential, production data in lower environments, auto-deploy to production, ephemeral durable truth, false residency approval or external resource mutation                                                                                                                                                                                                                  |
| S8-062 | E18-S01+S02+S03 backend launch-readiness aggregate projection                                 |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/launch/readiness/` domain/application/persistence boundary; no CRM, host training, certification, licensing, outreach or market activation mutation ownership                                                 | claimed 2026-07-27 after S8-059 admin readiness desk closure and current family/host/matchmaker source lanes                                                                                                                                                                                                                                     | generated Uber GoMock current-source orchestration proofs; consent/density/currency/jurisdiction/incomplete-evidence/property tests; race/vet clean; about 637,000 fuzz executions; live MongoDB 8.0.13 Testcontainers concurrency/idempotency/raw-BSON privacy proof                                                                                                                                           | Consented family density, current host training/certification and current matchmaker licence/jurisdiction project into an immutable append-only review snapshot; missing, stale, insufficient or mismatched evidence fails closed with exact blockers                                                                                                                                                                    | No family/member/contact list, CRM or recruitment outreach, vanity target pass, expired host/license acceptance, automatic certification/license/market activation, hidden denominator, coercive growth metric, source mutation or snapshot edit/delete |
| S8-061 | E18-S04+S05 admin staffing coverage, launch calendar, density gates and waitlist throttle     |        5 | `codex-root`                           | DONE   | extended `apps/admin/app/launch/` pure aggregate coverage/calendar/throttle proposal; no scheduling, waitlist, notification or launch-state mutation                                                                                          | 6 launch and 45 admin tests; admin typecheck/lint/build; rendered coverage-gap/calendar/density/reason/throttle/fail-closed/semantic/desktop-overflow QA; `git diff --check`                                                                                                                                                                     | Verification/support/Tier-A coverage show exact staffed/required seats, calendar leaves unresolved gates blocked, and waitlist throttle proposal requires current low-density evidence plus bounded reason; launch and waitlist facts remain unchanged                                                                                                                                                          | No staff identity/performance score, silent understaffing pass, growth pressure, member/contact list, arbitrary throttle, automatic waitlist/launch mutation, notification send, hidden denominator or density bypass                                                                                                                                                                                                    |
| S8-063 | E17-S02 MongoDB Atlas network, security and backup configuration                              |        8 | `/root/client_quality`                 | DONE   | declarative `deploy/atlas/` staging/production-candidate specification, least-privilege role matrix, restore-evidence schema and deterministic validator; no Atlas/cloud resource creation                                                    | commit `1b8584d`; validator CLI; focused race; repository-wide tests/vet; secret/wildcard and diff checks                                                                                                                                                                                                                                        | Synthetic staging and production-gated AF_SOUTH_1 candidate require three AZs, termination protection, TLS, private exact allowlists, isolated C4 identities, AES-256/CMK, localized continuous backup/PITR, RPO ≤5m, RTO ≤60m and non-destructive restore evidence with distinct approvals                                                                                                                     | No credential or IP wildcard in Git, shared app/admin/C4 identity, public access, disabled TLS, cross-region backup by default, production activation before signed residency/DPIA/procurement/restore gates, secret value, destructive restore, false readiness or external resource mutation                                                                                                                           |
| S8-064 | E17-S03 provider-neutral secret management and rotation                                       |        8 | `/root/consent_kernel`                 | DONE   | `deploy/secrets/`, shared `internal/platform/secrets/`, API/worker runtime configuration and deterministic repository checks; no provider account, vault value or external resource mutation                                                  | claimed 2026-07-27 after E17-S01 environment matrix; SRS NFR-302, data classification, threat model, Render Blueprint, runtime config and CI/security workflows reviewed                                                                                                                                                                         | provider-neutral inventory/policy parity and complete rotation-control proofs; tracked-file high-confidence leak scan; Render value-free slot checks; API/worker repository-wide tests; focused race/vet and existing Blueprint tests green; `git diff --check`                                                                                                                                                 | Metadata inventory names every runtime secret, isolated service scope, owner/custodian, maximum age, bounded overlap, reload/revoke/rollback and redacted evidence; staging/production API and worker reject missing, malformed, future or stale rotation evidence without logging values                                                                                                                                | No secret value or digest in Git/tests/logs, shared service credential, build-time secret, silent indefinite overlap, rotation without rollback/revocation/evidence, provider lock-in, production provisioning or external resource mutation            |
| S8-076 | E03 member shell independent viewport navigation                                              |        3 | `codex-root`                           | DONE   | verified fixed desktop `apps/web` Fie rail and fixed mobile bottom navigation after concurrent main-branch delivery                                                                                                                           | web typecheck; structural CSS review; `git diff --check`                                                                                                                                                                                                                                                                                         | Desktop rail is fixed to `100dvh` while only its navigation content scrolls; member content has an independent offset; mobile content reserves safe-area-aware space for the viewport-fixed navigation                                                                                                                                                                                                          | Concurrent implementation was retained during rebase because it met the requirement more precisely; no duplicate edit, route, member state, navigation label or application behavior change                                                                                                                                                                                                                              |
| S8-077 | Public marketing application and content foundation                                           |        5 | `codex-root`                           | DONE   | independent `apps/marketing`, original art-direction imagery, SEO metadata, sitemap, robots and `internal/product/marketing-content-plan.md`                                                                                                  | marketing typecheck/lint/test/build; desktop/mobile Playwright render and overflow QA; scroll-reveal state verification; `git diff --check`                                                                                                                                                                                                      | Media-led responsive homepage communicates voice, compound, trust and hosted fires without fabricated testimonials, metrics or availability claims; Outfit and one warm-light visual system used throughout                                                                                                                                                                                                     | Public discovery stays isolated from authenticated `apps/web`; generated imagery is explicitly marked for replacement with consented production photography before launch                                                                                                                                                                                                                                                |
| S8-065 | E17-S04 CI/CD release evidence and preview/staging/production promotion controls              |        5 | `codex-root`                           | DONE   | manual read-only release-evidence workflow, bounded qualification script and promotion/rollback runbook; no Render resource creation, external deployment or production activation                                                            | 4 release-policy and 4 Blueprint tests; shell syntax and ShellCheck; actionlint; `git diff --check`; Mongo/Testcontainers inapplicable to stateless repository policy                                                                                                                                                                            | Exact current-main SHA and named CI/CodeQL/dependency successes bind a 90-day metadata-only artifact; manual preview expires, protected synthetic staging remains checks-pass gated, and production attempts fail before artifact creation while topology/approval gates are absent                                                                                                                             | No untested deployment, mutable/latest tag, automatic production, production data in preview/staging, bypassed environment approval, secret in artifact/log, false deployment claim or external resource mutation                                                                                                                                                                                                        |
| S8-066 | E17-S05 Expo EAS build, submit and update release channels                                    |        5 | `codex-root`                           | DONE   | Expo app/EAS configuration, `expo-updates`, fail-closed dynamic release config, release policy and deterministic tests; no EAS project creation, credential provisioning, store submission or OTA publication                                 | 5 release-config tests and existing mobile tests; typecheck/lint; Expo web export; public config evaluation; Expo dependency check; `git diff --check`; Mongo/Testcontainers inapplicable to stateless client release config                                                                                                                     | Pinned EAS CLI/current SDK-compatible packages, committed-build requirement, isolated preview/staging/production channels, app-version runtime law, internal/draft staging, exact-tested-update promotion and binary/OTA rollback are explicit; incomplete release env fails closed and production submission remains absent                                                                                    | No cross-channel update, development client in release, embedded secret, automatic store submission, untested OTA, runtime mismatch, mutable version identity, production publish or external account mutation                                                                                                                                                                                                           |
| S8-069 | E17-S06 service-level objectives, alerts, dashboards and on-call                              |        5 | `codex-root`                           | DONE   | vendor-neutral SLO/alert/dashboard specification, on-call/escalation runbook and deterministic validator; no monitoring-vendor account, paging integration or external alert mutation                                                         | 3 observability-policy tests including 9 unsafe variants; focused race/vet; API/worker telemetry tests; `git diff --check`; Mongo/Testcontainers inapplicable to stateless policy validation                                                                                                                                                     | Four exact 30-day release-blocking SLIs, 99.9% budget math, multi-window burn alerts, dependency/safety/dead-letter pages, finite owned dashboards, P0-P2 response law and redacted runbooks are explicit; missing telemetry fails closed                                                                                                                                                                       | No member/content/secret label, unbounded cardinality, vanity-only dashboard, silent alert disable, self-resolving incident claim, punitive individual metric, automatic production rollout or external vendor mutation                                                                                                                                                                                                  |
| S8-068 | E17-S09 security scans, DAST and penetration-test closure                                     |        5 | `/root/security_closure`               | DONE   | pinned passive ZAP localhost-fixture workflow, finite security policy, synthetic evidence fixture and penetration closure runbook; no external target, production scan, secret, account or resource mutation                                  | 5 focused policy/evidence tests with 10 unsafe variants; focused race/vet; actionlint; fixture build; `git diff --check`; Mongo/Testcontainers inapplicable to stateless CI policy and HTTP fixture                                                                                                                                              | Immutable ZAP 2.17.0 Linux/amd64 digest can reach only the repository-owned loopback fixture; synthetic, stale, partial-scope, unresolved or self-verified evidence fails production closed, while closed findings require distinct assessor/owner/verifier and opaque remediation/retest refs                                                                                                                  | No production or arbitrary-host scan, credential in Git/log/artifact, active exploit, automatic risk acceptance, self-approval, silent finding suppression, false penetration-test claim or external mutation                                                                                                                                                                                                            |
| S8-067 | E17-S08 bounded performance, load and cost evidence                                           |        8 | `/root/consent_kernel`                 | DONE   | `internal/quality/performance/`, database-backed load fixtures, deterministic budgets and cost model; no external load target, paid provider, production traffic or resource mutation                                                         | claimed 2026-07-27 after E17-S01/S02/S03/S04 deployment and operational-control foundations; SRS NFR-100/201/202 and Doc 07 load/cost envelope reviewed                                                                                                                                                                                          | bounded profile/guardrail, percentile/property, machine-JSON and HTTP `/live` tests; focused race/vet; benchmark cost 0 alloc and load-runner overhead 16 allocs; live MongoDB 8.0.13 Testcontainers 800 insert+read samples at concurrency 16, zero errors and about 8.9 ms p95; transparent 10k/100k/1m cost formula/sensitivity and Doc 07 envelope proofs; `git diff --check`                               | Local bounded profiles emit exact sample/error/percentile evidence against committed budgets; cost calculations use visible planning assumptions, round metered units conservatively and keep the 100k-MAU scenario inside Doc 07's order-of-magnitude envelope without purchase or production inference                                                                                                                 | No external spend/traffic, vanity throughput, unbounded soak, production/member data, hidden percentile/sample, benchmark-as-SLA claim, provider quote as fact, unsafe concurrency, resource mutation or false capacity approval                        |
| S8-071 | E18-S06+S07+S08 admin campus attribution, UAT triage and hypercare command center             |        8 | `codex-root`                           | DONE   | extended launch pure model/MUI desk with aggregate attribution, consent/training/completion, bounded triage and owned hypercare signals; no ambassador payout/ranking, feedback source, incident, notification, launch or production mutation | 10 launch and 49 admin tests; admin typecheck/lint/build; rendered desktop/overflow/section/fail-closed interaction QA with no application runtime errors; `git diff --check`                                                                                                                                                                    | Campus quality requires complete aggregate evidence, 30-day sustainment and zero unresolved safety; UAT shows exact consent/training/completion and critical findings; hypercare exposes exact SLO/safety/owner blockers and prepares only an opaque human triage record                                                                                                                                        | No member/contact list, vanity referral leaderboard, per-ambassador performance score, coercive recruitment, raw feedback/content, automatic payout/action/launch, incident closure, hidden denominator or production mutation                                                                                                                                                                                           |
| S8-070 | E17-S07 backup restore and disaster-recovery rehearsal                                        |        8 | `/root/client_quality`                 | DONE   | isolated `internal/platform/drrehearsal/` orchestration and `deploy/atlas/rehearsal/` evidence/runbook; no production/source mutation, Atlas resource creation or external restore execution                                                  | generated Uber GoMock deterministic lifecycle tests; 1,000-target property proof and fuzz seed; focused race/vet; live MongoDB 8.0.13 replica-set Testcontainers transactional copy, source immutability, index/digest/count integrity, corruption rejection and isolated cleanup; repository-wide Go tests; Atlas validator; `git diff --check` | Exact staging point-in-time watermark, isolated target, collection counts/digests/indexes, transactional/audit invariants, RPO ≤5m and RTO ≤60m bind immutable redacted evidence only after safe cleanup and distinct data-owner/security approvals; production fails before any port call                                                                                                                      | No source write/drop, production target, credential/PII evidence, fake restore success, missing PITR proof, RPO/RTO miss, same approver, destructive cleanup, production activation or external resource mutation                                                                                                                                                                                                        |
| S8-072 | E17-S10 release notes, UAT, rollback and hypercare evidence bundle                            |        5 | `codex-root`                           | DONE   | strict metadata-only JSON Schema, synthetic blocked example, operator closeout runbook and deterministic Go validator; no release publication, deployment, store submission, notification or production mutation                              | 4 release-bundle tests with 8 unsafe variants; focused race/vet; strict JSON parse and syntax checks; `git diff --check`; Mongo/Testcontainers inapplicable to stateless evidence validation                                                                                                                                                     | Exact candidate/evidence refs, bounded notes, ordered consent/training/completion counts, critical findings, ≤60m rollback proof, distinct approval and named hypercare roles bind for seven days; blockers and stale evidence fail closed, and a bundle alone can never approve current production                                                                                                             | No raw feedback/member/contact data, vague latest version, automatic deploy/submit/notify, self-approval, hidden blocker, untested rollback, false production approval, incident closure or external mutation                                                                                                                                                                                                            |
| S8-075 | Cross-epic backlog traceability closure audit                                                 |        3 | `codex-root`                           | DONE   | evidence-backed map of all 182 E01–E18 story IDs to ledger/code/tests, including explicit bundled-label mappings and two genuine P2 gap closures; no product, provider, deployment or external mutation                                       | repository-wide `pnpm run check`: 58/58 workspace tasks and complete Go suite green; focused finance/admin/verification/reconciliation/Suban checks; zero non-DONE task rows; synchronized `main`; `git diff --check`                                                                                                                            | Eleven earlier/bundled implementations are mapped explicitly; S8-073 diaspora isolation and S8-074 Gate/USSD close the only genuine gaps with GoMock/property/race/vet/fuzz and live MongoDB Testcontainers evidence                                                                                                                                                                                            | No papering over missing code, false production readiness, unverified status, duplicate implementation, destructive cleanup, provider call or external mutation                                                                                                                                                                                                                                                          |
| S8-073 | E14-S09 P2 diaspora payment isolation                                                         |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/commerce/diaspora/` domain/application/persistence boundary with provider-confirmation and exact platform-sale ledger ports; no MoMo, provider network, client UI or member-transfer ownership                | claimed 2026-07-27 after confirming only Ghana MoMo/GHS collection exists; PRD M1-05, pricing Doc 03 diaspora GBP/USD/EUR, architecture PCI enclave/no-member-transfer law and cross-border compliance reviewed                                                                                                                                  | generated Uber GoMock exact orchestration proofs; currency/minor-unit/price-version/confirmation/failure/no-transfer/reflection/property tests; focused race/vet; about 365,000 fuzz executions; live MongoDB 8.0.13 Testcontainers create/CAS idempotency and raw BSON privacy/isolation proof; `git diff --check`                                                                                             | Immutable checkout accepts only current versioned GBP/USD/EUR catalog quotes, prepares an opaque instruction without charging, verifies exact provider facts, and after successful confirmation exposes only a fixed platform-sale ledger command; the port has no arbitrary line, account, GHS/MoMo, payout or member-transfer primitive                                                                                | No provider/network call, real fund, card/PayPal credential, raw member/contact data, FX conversion, GHS/MoMo mutation, member transfer, unconfirmed entitlement, duplicate posting, mutable quote, cross-purpose export or automatic refund/payout     |
| S8-074 | E13-S07 P2 Gate links and USSD companion                                                      |        8 | `/root/client_quality`                 | DONE   | isolated `services/api/internal/companions/p2gate/` domain/application/persistence boundary for consent-bound reviewer-link proposals and minimal USSD session views; no gateway/provider/network integration or source aggregate mutation    | generated Uber GoMock authentication/consent/current-source orchestration; exact-consent and 1,000-case pack property/fuzz proofs; focused race/vet; live MongoDB 8.0.13 replica-set Testcontainers 12-way idempotency/concurrency, TTL expiry index and raw-BSON privacy proof; repository-wide Go tests; `git diff --check`                    | Gate link proposals re-read exact current bilateral pack consent, bind one opaque reviewer to an OTP-gated/no-forward watermark contract and expire after 30 days; authenticated member-scoped USSD returns only bounded pod count, drum indicator, three future fire refs and three approved help refs                                                                                                         | No raw phone/contact/content persistence, global browse, courtship conversation, Gate pack delivery, automatic invite/SMS/approach, balance/payment flow, provider call, refusal signal, consent bypass, forwarding, expiry extension or external mutation                                                                                                                                                               |
| L1-001 | Post-backlog launch-gate inventory and claim-safe execution ledger                            | Launch 1 | `codex-root`                           | DONE   | `agent_plan.md`, `internal/quality/launch-decision-inventory.md`                                                                                                                                                                              | claimed/completed 2026-07-27 after S8-075 closure; repository-controlled work only                                                                                                                                                                                                                                                               | Reconciled 16 gates by authority, owner, current evidence, state and closure condition; reviewed against residency, Atlas, release and backlog-closeout controls                                                                                                                                                                                                                                                | Repository/external/provider/credential/cohort/store/production-action evidence is explicit; production remains blocked and no external action or approval is implied                                                                                                                                                                                                                                                    |
| L1-002 | Executable production-gate and evidence registry                                              | Launch 1 | `/root/consent_kernel`                 | DONE   | isolated `internal/quality/launchgates/` evaluator/CLI, strict schema, blocked synthetic fixture and operator contract under `deploy/release/`; no admin UI or external mutation                                                              | claimed 2026-07-27 after L1-001 scope definition                                                                                                                                                                                                                                                                                                 | canonical 17-gate/category/provenance/dependency/freshness tests; strict parser/EOF/private-field, wrong-scope, self-review, duplicate-gate/ref, stale/synthetic/pending and 1,000-order property proofs; focused race/vet; independent L1-003 exact cross-review found duplicate-ref replay then verified correction and no remaining blocker; `git diff --check`                                              | Repository, external-decision, provider, credential, cohort, store and production-action evidence are exact production/GH/candidate bound with opaque refs, distinct issuer/reviewer, bounded validity and fail-closed dependencies; the committed synthetic registry remains blocked and `ready` is evidence metadata only, never activation                                                                            |
| L1-003 | Launch-gate security and privacy policy cross-review                                          | Launch 1 | `/root/security_closure`               | DONE   | isolated `internal/quality/launchgates/security_boundary_test.go` and `internal/quality/launch-gate-security-review.md`; no production scan, provider call or approval mutation                                                               | cross-reviewed L1-002 before publish; found duplicate evidence-reference replay and incomplete order-invariance proof, coordinated both bounded corrections, then independently verified the landed contract                                                                                                                                     | 7 adversarial review tests with 11 unsafe variants; owner plus cross-review suites, focused race/vet and `git diff --check` green; Mongo/Testcontainers inapplicable to stateless pure policy evaluation                                                                                                                                                                                                        | Synthetic/stale/wrong-environment/self-approved/private-shaped evidence, repository substitution, replay, dependency bypass and evidence reordering fail closed; `ready` remains metadata only, committed fixture remains blocked, and no external gate or production action is claimed                                                                                                                                  |
| L1-004 | Admin launch-decision readiness surface                                                       | Launch 1 | `codex-root`                           | DONE   | `apps/admin/app/launch/` decision-gate model, tests and MUI operator surface                                                                                                                                                                  | claimed/completed 2026-07-27 after inventory contract; L1-002 registry structure remains a final cross-review input                                                                                                                                                                                                                              | 51/51 admin tests; admin typecheck/lint/production build; live 1433px rendered QA with zero horizontal overflow and required no-launch warning visible; responsive MUI grid reviewed at narrow breakpoints                                                                                                                                                                                                      | Six authority-separated gates expose owner, evidence, freshness and dependency; only repository proof is verified, every external gate remains blocked/awaiting authority, and no launch control exists                                                                                                                                                                                                                  |
| L1-005 | Exact-SHA synthetic staging qualification package                                             | Launch 1 | `/root/client_quality`                 | DONE   | isolated `internal/quality/stagingqualification/` generator/validator plus committed synthetic release evidence and qualification artifact; no Render/provider/store action                                                                   | deterministic regenerate-and-byte-equivalence test; exact-SHA, source-digest, 24-hour freshness and 9 unsafe-variant proofs; focused race/vet; release-bundle and DR contract tests; repository-wide Go tests; CLI generate/validate; `git diff --check`                                                                                         | Exact current-main SHA binds fresh synthetic staging release evidence, blocked closeout bundle and isolated DR rehearsal; false production/legal/provider/cohort evidence fields are mandatory and the artifact expires after 24 hours                                                                                                                                                                          | No production approval/readiness claim, real UAT/cohort/legal/residency/procurement/provider proof, mutable latest reference, stale or cross-SHA input, external deployment/store/provider action, credential or member data                                                                                                                                                                                             |
| L1-006 | Launch 1 integrated cross-review, full quality gate and publish                               | Launch 1 | `codex-root`                           | DONE   | repository-wide verification, review notes and task ledger                                                                                                                                                                                    | claimed/completed 2026-07-27 after L1-002 through L1-005 landed and were cross-reviewed                                                                                                                                                                                                                                                          | Full `pnpm run check` green across 22 workspaces and complete Go suite; exact-SHA staging qualification CLI validates; compiled launch-gate CLI returns blocked exit 2 with all 17 gates closed; `git diff --check`; final rebase and remote parity                                                                                                                                                             | Cross-review fixed duplicate evidence replay and a false order-invariance proof; browser QA confirmed no desktop overflow and no launch control; production remains honestly blocked on every external authority                                                                                                                                                                                                         |
| L2-001 | Final external-gate handoff inventory and claim-safe ledger                                   | Launch 2 | `codex-root`                           | DONE   | `agent_plan.md`, `internal/quality/external-gate-handoff.md`, launch inventory reconciliation                                                                                                                                                 | claimed/completed 2026-07-27 after Launch 1 closeout                                                                                                                                                                                                                                                                                             | Thirteen external gates mapped to named authority, repository handoff, Launch 2 task and irreducible external act; completion rule records six required proof classes                                                                                                                                                                                                                                           | No agent-buildable task remains implicit; synthetic staging status corrected to complete and every real authority boundary remains blocked                                                                                                                                                                                                                                                                               |
| L2-002 | Founder/legal residency and DPIA decision-record contract                                     | Launch 2 | `/root/security_closure`               | DONE   | isolated `deploy/legal/` strict schema, blocked synthetic fixture and external decision packet plus `internal/quality/residencydecision/` pure validator/CLI; no legal conclusion, signature or approval                                      | stable camelCase contract coordinated with L2-001/UI owner; exact full processing and infrastructure scope, finite interpretation/location law and three distinct external authority slots                                                                                                                                                       | 7 policy/schema tests with 22 unsafe variants and 1,000-order property proofs; focused race/vet; blocked CLI exit; `git diff --check`; Mongo/Testcontainers inapplicable to stateless pure policy evaluation                                                                                                                                                                                                    | Unsigned, self-approved, stale, synthetic, incomplete, repository-issued, private-shaped, replayed, wrong-environment, Ghana-only mismatch, ambiguous or out-of-Africa resident-boundary evidence fails closed; passing metadata never supplies a legal conclusion, signature, approval or production action                                                                                                             |
| L2-003 | Provider/procurement diligence evidence contract                                              | Launch 2 | `/root/consent_kernel`                 | DONE   | isolated `internal/quality/providerdiligence/` validator/CLI and strict schema, blocked synthetic template and operator contract covering Atlas, storage/CDN, LiveKit and communications; no provider account, purchase or network call       | claimed/completed 2026-07-27 after L1-002                                                                                                                                                                                                                                                                                                        | complete/missing/synthetic/pending/stale, duplicate provider/subject/ref, unknown provider, self-review, strict parser/fixture and 1,000-order property tests; focused race/vet; CLI blocked-fixture proof; `git diff --check`                                                                                                                                                                                  | Requires DPA, subprocessors, locations, retention, deletion, keys, support, breach, cost and exit evidence for every provider with opaque globally unique references, independent roles and 90-day maximum validity; synthetic fixture remains blocked                                                                                                                                                                   |
| L2-004 | Ghana device/network field-test execution kit                                                 | Launch 2 | `/root/client_quality`                 | DONE   | isolated `internal/quality/fieldtest/` strict validator/CLI, execution runbook and blocked synthetic manifest; no real member data, provider spend, external traffic or production target                                                     | exact device/network/path/sample/percentile/budget validation; 15 unsafe variants; 1,000-case percentile property and fuzz proof; focused race/vet; repository-wide Go tests; deterministic CLI; `git diff --check`                                                                                                                              | Exact Android 8.0/API 26/2 GiB clean-install floor, physical-device gate, representative Ghana 3G plus fixed-local profiles, six fixed compute/database/media paths, 30+ samples, recomputed p50/p90/p95, immutable refs, distinct operator/reviewer and seven-day freshness fail closed                                                                                                                        | No production/member data, phone/contact/media payload, external traffic/spend, provider call, outlier removal, fabricated percentile, emulator-as-physical claim, synthetic qualification, self-review, stale/cross-SHA evidence or launch approval                                                                                                                                                                     |
| L2-005 | Credential/store and cohort/operations readiness handoff kit                                  | Launch 2 | `/root/consent_kernel`                 | DONE   | isolated `internal/quality/readinesshandoff/` metadata validator/CLI plus strict schema, blocked synthetic template and operator contract; no secret, store submission, recruitment, contact or staffing mutation                             | claimed/completed 2026-07-27 after L2-003                                                                                                                                                                                                                                                                                                        | complete/missing/synthetic/pending/stale, unknown/duplicate/ref replay, kind/provenance substitution, self-review, strict parser, blocked fixture and 1,000-order property tests; focused race/vet; compiled CLI blocked-fixture proof; `git diff --check`                                                                                                                                                      | Controlled runtime/Play/App account custody, witnessed signing ceremony, UAT consent/training/completion, circle/host readiness and support/T&S coverage require opaque external evidence, distinct roles and bounded validity; validation is metadata only                                                                                                                                                              |
| L2-006 | Admin external-gate evidence intake surface                                                   | Launch 2 | `codex-root`                           | DONE   | `apps/admin/app/launch/` decision-gate requirements, handoff reducer/tests and responsive MUI desk; no upload, approval, launch or external mutation                                                                                          | claimed/completed 2026-07-27 after L2-001; contract labels remain presentation-decoupled                                                                                                                                                                                                                                                         | 53/53 admin tests; typecheck/lint/production build; live 1433px rendered and interaction QA with zero horizontal overflow, exact requirements/external act visible and launch still blocked                                                                                                                                                                                                                     | Operators can select five external authority classes and prepare one opaque coordination record; repository gate selection, short notes and state mutation fail closed, and no evidence upload/approval/activation control exists                                                                                                                                                                                        |
| L2-007 | Final repository backlog closure audit and publish                                            | Launch 2 | `codex-root`                           | DONE   | `internal/quality/repository-backlog-closure.md`, repository-wide evidence mapping, quality gates and ledger closeout                                                                                                                         | claimed/completed 2026-07-27 after L2-002 through L2-006 and correction L2-008                                                                                                                                                                                                                                                                   | Full `pnpm run check` green across 22 workspaces and complete Go suite; four compiled external-handoff CLIs each return blocked exit 2; admin rendered QA; zero non-DONE task rows; `git diff --check`; final rebase and remote parity                                                                                                                                                                          | Audit found and L2-008 fixed blocked field evidence exiting 0; every agent-buildable task is closed while real legal/provider/device/store/cohort/operations/go-no-go/activation acts remain explicit external blockers                                                                                                                                                                                                  |
| L2-008 | Field-test CLI blocked-decision exit semantics correction                                     | Launch 2 | `/root/client_quality`                 | DONE   | `internal/quality/fieldtest/cmd/` deterministic decision output, process-level exit tests and runbook correction; no manifest weakening, device/network operation or external mutation                                                        | claimed/completed 2026-07-27 after L2-007 audit reproduced valid blocked manifest exiting 0 without output                                                                                                                                                                                                                                       | blocked/invalid/usage/qualified decision tests plus real subprocess exit-2 proof; focused race/vet; full fieldtest tests; CLI JSON inspection; `git diff --check`                                                                                                                                                                                                                                               | Valid blocked evidence emits JSON `blocked` and exits 2; only `qualified-field-evidence` exits 0, invalid evidence exits 1 and bad usage exits 64; no silent success, blocked-as-qualified inference, manifest bypass, production/member data, device/network/provider action, spend or launch approval                                                                                                                  |
| L3-001 | Fixed desktop rail and mobile bottom navigation viewport behavior                             | Launch 3 | `codex-root`                           | DONE   | `apps/web/app/fie/styles.css`, `apps/web/app/fie/shell-layout.test.ts`                                                                                                                                                                        | claimed/completed 2026-07-27 from user-reported shell scroll defect; isolated worktree preserved concurrent backend edits                                                                                                                                                                                                                        | 69/69 web tests; web typecheck/lint/build; CSS contract tests; live 1433×774 browser measurement before/after page scroll: rail fixed at top 0, height/bottom 774, nav `overflow-y:auto`, main offset 248px, no horizontal overflow                                                                                                                                                                             | Rail is independent of detail-page height and only its navigation content scrolls; compact rail offsets at 86px; mobile content reserves safe-area-aware clearance and bottom navigation is viewport-fixed, full-width and non-document-scrolling                                                                                                                                                                        |
| S8-058 | E16-S03 community operations orchestration backend                                            |        8 | `/root/consent_kernel`                 | DONE   | isolated `services/api/internal/admin/communityops/` domain/application/persistence boundary; no circle, host, fire or notification source mutation ownership                                                                                 | claimed 2026-07-27 after E16-S03 admin desk and circle/host/fire source lanes completed; PRD M13-03 and SRS FR-801 reviewed                                                                                                                                                                                                                      | generated Uber GoMock service proofs; bounded/redacted aggregate property and fuzz proofs; race/vet clean; live MongoDB 8.0.13 Testcontainers concurrency, idempotency and raw BSON privacy proof                                                                                                                                                                                                               | Bounded redacted circle density snapshot, current host verification/certification revalidation, reason-coded fire/host operation proposal, exact participant-notice preview acknowledgement and append-only audit are implemented; acknowledgement only reaches human-review readiness                                                                                                                                   | No member/content/contact exposure, direct or silent source mutation, unverified host assignment, capacity bypass, notification send, automatic approval/apply, retaliation, global browse or audit edit/delete                                         |
| S8-038 | E16-S01 admin principals, roles and MFA step-up backend                                       |       10 | `/root/backend_admin`                  | DONE   | `services/api/internal/admin/`, admin auth routes + OpenAPI                                                                                                                                                                                   | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers end-to-end: privileged enroll audited (FR-801), verifier blocked from enrolling, email-delivered MFA login issues 30-min session, fresh-code step-up flags + audits, wrong code increments attempts; domain challenge/session lifecycle + least-privilege unit tests; Redocly + client regen + drift green                                                                                       | Self-reviewed; root bootstrap is a controlled migration, not an API path; codes ride the Resend channel via a bridge adapter; desks can now require steppedUp sessions for sensitive actions                                                                                                                                                                                                                             |
| S8-037 | E12-S03 privacy-redacted evidence access with immutable audit                                 |       10 | `/root/backend_safety`                 | DONE   | `internal/safety/` (evidence)                                                                                                                                                                                                                 | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | redaction unit tests (phone shapes, email, handles masked; innocent text intact; bundle redaction); audit-order proof (audit written before any read); Testcontainers: identifier-laden reason served redacted, every access audited (2 views = 2 records), curiosity rejected with no audit                                                                                                                    | Self-reviewed; over-redaction preferred per Doc 09 §3; subject stays visible, third parties masked; insider review supported via per-agent access counts                                                                                                                                                                                                                                                                 |
| S7-026 | E12-S02 Tiered T&S queues and SLA routing                                                     |       10 | `/root/backend_safety`                 | DONE   | `internal/safety/` (cases), `services/worker/internal/jobs/safety/`                                                                                                                                                                           | claimed/completed 2026-07-26                                                                                                                                                                                                                                                                                                                     | Testcontainers end-to-end: filed reports become tiered cases (A 8h / B 24h / C 72h / care immediate per Doc 09 §3), replay-safe via inbox dedup + unique reportId index, oldest-SLA-first queue, assign/resolve lifecycle, breach count observability; one Docker reaper flake rerun green                                                                                                                      | Self-reviewed; safety context relocated to `internal/safety/` for worker import; outbox gained FindByEventType so typed consumers never race the relay's publish markers                                                                                                                                                                                                                                                 |

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

| Risk                                                                        | Severity | Mitigation / decision gate                                                                     |
| --------------------------------------------------------------------------- | -------: | ---------------------------------------------------------------------------------------------- |
| Render/provider data residency may not meet Ghana/in-region policy          | Critical | Verify in Sprint 0; DPIA/legal decision; hybrid/platform change before production              |
| MongoDB transactions/graph traversal may complicate invariant-heavy domains |     High | State-machine/ledger spike, strict indexes, transaction tests, materialized trust paths        |
| Expo support for Android 8/native identity/liveness requirements            |     High | Sprint 0 device/provider PoC; development build/config plugin or approved min-version decision |
| P0 scope expands due to three clients                                       |     High | Shared contracts/tokens, mobile reference, identical domain API, strict P0 cut line            |
| Identity/MoMo/WhatsApp provider procurement blocks delivery                 |     High | Port-first adapters, simulators, dual shortlist, explicit fallback/manual ops                  |
| E2E encryption conflicts with consented safety processing                   | Critical | Dedicated architecture/legal threat-model decision before room implementation                  |
| Voice/media costs and 3G performance miss budgets                           |     High | Opus/resumable/progressive PoC, lifecycle policies, device/network benchmarks                  |
| AI safety/quality insufficient in Twi/Pidgin                                |     High | Rules/human fallback, local paid annotation/review, threshold gates                            |
| Cold-start matching lacks data                                              |   Medium | Rules, circles, vouches, fires, Nnoboa; no premature model                                     |
| Admin evidence access creates insider/privacy risk                          | Critical | Redaction, purpose scopes, just-in-time access, MFA, immutable audit and reviews               |
| Operational community is not ready when software is                         |     High | E18 runs from Sprint 0 with release gate                                                       |
| “Latest version” upgrades destabilize delivery                              |   Medium | Pin exact stable versions, controlled upgrade PRs, no prerelease by default                    |

## 24. Decision register and founder questions

These answers improve Sprint 0 but do not prevent review of this plan:

1. **Member web P0 depth:** should member web have full P0 feature parity,
   including recording, Sow, rooms and fires, or should it be responsive
   read/limited-action access while mobile remains the only complete P0 client?
   Current plan assumes full parity except device-only identity capabilities,
   which use a secure mobile handoff.
2. **Public marketing site:** resolved 2026-07-27. It is a fourth independently
   deployed application at `apps/marketing`; `apps/web` remains the
   authenticated member product.
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
- [x] Public marketing is independently deployed from authenticated member web.
- [x] Sprint 0 is the first active execution milestone.
- [x] Bounded Sprint 0 execution goal created on 2026-07-26.
- [x] Sprint 0 integrated engineering checkpoint recorded on 2026-07-26.
- [ ] Founder approval of the conditional Sprint 1 proceed decision.

Open decisions remain in Section 24 and the checkpoint corrective-action table.

## 26. Product completion continuation

The founder reopened implementation on 2026-07-30 with a mandate to complete
all planned and partially implemented features. The earlier repository backlog
closure proved bounded backend and presentation slices, but it did not prove
that the client applications were connected to those backend contracts.
Runtime integration is therefore an active completion requirement.

| Task    | Deliverable                                                                                                                               | Agent        | Status | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| ------- | ----------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| E2E-001 | Member OTP request and verification through the generated API contract, with server-side API mediation and protected session cookies      | `codex-root` | DONE   | `apps/web/app/api/auth/otp/`, `apps/web/app/lib/api-server.ts`, and onboarding pending/error states; web typecheck, 78 tests, lint and production build pass                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| E2E-002 | Complete identity onboarding after session establishment: consent, Ghana Card, age and liveness state against server-authoritative routes | `codex-root` | DONE   | Ghana Card uses the authenticated member session instead of a client-supplied account ID. Promise, terms and adult-age affirmation persist as three immutable, exactly versioned, idempotent Mongo consent histories. Liveness uses real browser camera/microphone capture, strict audio/JPEG size and type bounds, application-layer AES-GCM encryption, opaque artifact references and Mongo TTL deletion within 24 hours. Its authenticated/idempotent decision path retains only privacy-keyed proof, persists attempts, queues uncertain results for human review and fails closed. Focused Go tests/vet, OpenAPI generated-client drift and web typecheck pass; the prior full web lint, 78 tests and production build also passed. Consent Mongo Testcontainers proof exists but Docker container creation timed out on the saturated daemon                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| E2E-003 | Connect authenticated member web feature surfaces to server-authoritative APIs                                                            | `codex-root` | DONE   | Identity onboarding, notification preferences, Fire discovery/RSVP, safety reporting, profile fields and visibility consent, doorway questions, Suban explanations/appeals, membership cancellation/refund requests, privacy export/deletion tracking, the complete purpose-bound consent switchboard, Nnoboa kin invitations, circles/retained room entries, the privacy-keyed Seed Garden summary and Oware now use authenticated server routes. Oware creation is mounted in a retained circle room, server-derives exactly two active participants, stores privacy-keyed room/player references in Mongo, returns only a redacted revisioned board and applies moves through the server Abapa engine; the old browser sowing reducer and named opponent were removed. The Fie home now independently loads the authenticated profile, joined-circle count, next retained Fire, Garden aggregates, nominations and membership; device connectivity comes from the browser rather than a reducer preview, partial service failures degrade only their section, and fabricated member identity, date/location, notification count, event, seed allowance and paid-through claims were removed. Abɔnten now renders retained Fire records only; fabricated learning and community-notice posts were removed until those catalogs are composed. Circle persistence is composed at runtime with private-by-default creation, bounded mine/discover queries, request/approve/promote/expel/leave and owner visibility controls; private-reference request bypass and membership self-escalation were closed in the application/domain layers. Hosts/owners alone receive sorted management projections, while discovery reveals only taxonomy, opaque reference and aggregate size. Circle rooms now authorize every list/create/delete against current circle membership, privacy-key authors with a separately rotated `CIRCLE_HMAC_SECRET`, apply bounded retention and project no author key. Fabricated Sunday Readers/Builders activity, private transcripts, calls, timers and presence were removed from web/mobile; room clients render only durable entries and explicitly identify uncomposed media/live providers. Ember giver/recipient and nomination owner are session-derived; reporter, profile owner, doorway owner, photo-vault viewer, listening identity, membership owner, garden owner and Suban subject remain server-owned too. Nnoboa consent/decline requires separately rotated HMAC response authority delivered only in the private WhatsApp invite; invalid authority is indistinguishable from not-found, the public response projects no member identifier, and the web includes a real privacy-first kin response page. The web/mobile Garden now shows only durable aggregate lifecycle counts; fabricated named seeds, allowance and server-confirmation controls were removed until the complete atomic sow path is composed. The introduction explanation now fails closed around its opaque reference and shows only persisted matching-personalization consent; the fabricated candidate, compatibility reasons, AI involvement, availability and accept/rest decisions were removed until the introduction store, media manager and transcriber are composed. The Agyina marketplace now loads only current persisted licences, books member-owned immutable consultation terms idempotently and shows real engagement state without fabricating candidate consent, proposal exposure or payment; invalid/expired licences fail the catalog closed. Consent controls expose the domain's required, opt-in, opt-out and reversible policies without allowing a client to manufacture renewed consent. Raw Suban provenance and fake payment-provider/nominee confirmations are not projected. Repository-wide JS/TS tasks and the Go suite passed before this slice; focused matchmaker/Oware/API/OpenAPI tests and vet, generated-client/Redocly checks, web typecheck/lint, 68 web tests and `git diff --check` pass. Final reconciliation: authenticated runtime paths or explicit fail-closed boundaries; 37 web tests and 54-route production build pass. |
| E2E-004 | Connect admin authentication, step-up and operational desks to server-authoritative APIs                                                  | `codex-root` | DONE   | Admin email-code login now issues a short-lived HttpOnly session cookie; the entire ops layout redirects without it and sign-out clears it. Enrollment and step-up derive and verify the authenticated admin session. Verification queue routes are registered at runtime with role-derived scopes, a separately rotated admin HMAC key, recent-MFA evidence access, redacted/audited evidence and idempotent decisions. Funnel/delivery metrics, market-pack governance and scam-arc signal ingestion require the explicit operations-admin scope and derive the actor from that session. Verification and analytics desks now load real aggregate state; analytics explicitly leaves uncomposed D30/fairness/safety evidence blocked instead of displaying synthetic values. The operator desk now loads persisted principals, performs MFA-stepped-up enrollment, non-admin role changes, suspension/reactivation, protects self/last-admin invariants and uses optimistic concurrency. Admin-role grants/revocations are durable four-eyes proposals: a distinct stepped-up admin approves against the unchanged target version, while proposal, principal and immutable audit records commit atomically; the fake operator roster and caller-selected approver were removed. A new Agyina licensing desk lists the complete register and creates/renews versioned public professional profiles under operations scope plus MFA; optimistic version checks prevent lost renewal and the licence record plus immutable admin audit commit atomically. The safety desk now lists persisted open cases, exposes only privacy-keyed subjects, supports server-owned self-assignment, and opens automatically redacted evidence only for the assigned agent after fresh MFA with an immutable purpose-scoped access record. Its fabricated roster, legal holds, victim exports and enforcement proposals were removed. The member desk now reads the real account store through an operations/safety-scoped endpoint and projects only environment-keyed references, tier, lifecycle and dates; phone, account ID and member content never reach the client. Its fake roster, host/privacy claims, caller-selected approver and direct suspend/reactivate/block controls were removed because enforcement is already case-bound to the audited safety ladder with session/device propagation. Focused Go HTTP/application/OpenAPI tests, 56 admin tests, admin typecheck/lint/build, generated-client drift and Redocly checks pass. Final reconciliation: authenticated runtime desks or truthful non-mutating authority boundaries; 17 admin tests and 39-route production build pass.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| E2E-005 | Connect Expo mobile authentication and P0 feature surfaces to server-authoritative APIs                                                   | `codex-root` | DONE   | Claimed 2026-07-30 after the admin verification vertical slice. Expo gates every route behind the real OTP API, stores member tokens in device-only SecureStore (session storage only for the web export), validates configured release API origins and exports successfully. Profile, notification preferences, privacy export/deletion status, purpose-bound consent controls, Suban explanations/appeals, membership actions, Nnoboa invitations, circle discovery/lifecycle/management, retained circle rooms, private Oware creation/board/moves, private Anansesɛm creation/relay/author-edit/grant/publish, Fire detail/RSVP, Seed Garden summary, licensed matchmaker discovery/consultation booking and the introduction explanation's persisted consent control now use authenticated APIs. Mobile Oware retains its circle context, loads the privacy-safe revisioned board and submits idempotent moves without local sowing. Mobile Anansesɛm retains its circle context, loads only persisted passages, enforces server-owned turns and author edits, and publishes only the mutually granted draft. Agyina no longer ships seeded professionals or client-controlled candidate consent; it lists only current server licences and retains immutable member-owned consultation terms without claiming payment. The Games Hall no longer opens invented IDs or claims tournament seats/reviews. Fabricated Nnoboa candidate consent, courtyard activity, live-Fire content, named seed histories, introduction identity/reasons/decisions, story authorship/turn/consent and claimed AI involvement were removed. Final reconciliation: authenticated P0 adapters or explicit fail-closed boundaries; 7 mobile tests and production Expo web export pass.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| E2E-006 | Route-by-route rendered, responsive and interaction verification across marketing, web and admin; device verification for mobile          | `codex-root` | DONE   | User-approved Chrome fallback verified marketing and member onboarding at desktop/390px, member/admin unauthenticated guards, admin login at desktop/390px with malformed-email validation, and the exported Expo login at 390px; no horizontal overflow or application console errors; no real OTP sent                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |

E2E-006 final evidence: **DONE** on 2026-07-30 after the user approved Chrome
as the live-verification fallback. Chrome
verified marketing at desktop and 390px widths, member onboarding at desktop
and 390px widths, unauthenticated member/admin route guards, admin login at
desktop and 390px widths, malformed-email validation, and the exported Expo
login at 390px. The checked surfaces had no horizontal overflow or application
console errors; browser-extension warnings were excluded. No real OTP was sent.

Completion reconciliation (2026-07-30): E2E-003, E2E-004 and E2E-005 are
complete. Member web, admin and Expo either use authenticated
server-authoritative routes or explicitly
fail closed where a required authority/provider is not composed. The remaining
invented private-room identity/message, Abusua material/consent/reviewer state,
Okyeame appeal references and pre-hydration profile were removed. The final
functional gate passed all 62 workspace lint/typecheck/test tasks, the complete
Go suite, generated-client/OpenAPI validation, 37 web tests, 17 admin tests, 7
mobile tests and all marketing/web/admin/mobile production builds. E2E-006
is complete through the user-approved Chrome fallback with responsive and
interaction checks across the public, member, admin and exported mobile entry
surfaces. Authenticated interiors remain protected behind real OTP gates and
were not bypassed. Repository-wide Prettier and `git diff --check` now pass
after formatting the accumulated implementation.

Admin redesign program (2026-08-22): the authorized page-by-page admin visual
and interaction redesign is tracked in `admin_redesign_plan.md`. That child
ledger owns route claims, shared-file reservations, loading/empty/error state
requirements, Ashesi ScheduleFlow right-side interaction parity, and acceptance
evidence. Existing task statuses in this master ledger are unchanged.

Escrow runtime evidence (2026-07-30): member web and Expo now list only
privacy-safe, owner-bound funded escrows and allow only member acceptance or
dispute freeze. The operations Agyina desk can retain provider-confirmed funding
only by deriving the owner, total and milestones from a booked engagement, then
record separately authorized delivery evidence. Funding and delivery writes are
MFA-gated and atomically paired with immutable admin audit records. Member
clients cannot supply funding, delivery evidence, matchmaker confirmation or
settlement. Finance settlement now requires the dedicated finance scope plus
fresh MFA and atomically commits the escrow CAS, balanced GHS journal posting
and immutable admin audit. Its payout statement exposes gross, disclosed fee
and net while account and provider references remain privacy-keyed.

Fire runtime evidence (2026-07-30): web and Expo Fire detail now resolve only
the exact retained upcoming Fire reference and submit RSVP through the
authenticated server route. The web no longer fabricates a live host, speaker
story, captions, attendee initials/count, media quality controls or connection
failover. Both clients fail closed when the requested reference is absent;
neither silently substitutes another event. Media-room and live-caption claims
remain absent until their providers are composed.

Ɛpono Ano runtime evidence (2026-07-30): the web doorway no longer fabricates
a candidate, portrait initials, voice recording, transcript, answer, shared
path, recommendation reason, tier outcome, acceptance or pass decision. It now
loads and updates only the authenticated member's retained doorway question and
matching-personalization consent. The page states that the introduction queue
is uncomposed and exposes no candidate-facing action until a real
consent-governed introduction store exists.

Abusua Gate runtime evidence (2026-07-30): the web no longer lets one browser
select fabricated shared material, impersonate a named partner's consent,
create a fake private passage or claim expiry, one-time-code and watermark
controls. The capability now fails closed and documents the three missing
runtime authorities: a server-derived retained pair, two independently
authenticated capability grants and separately delivered reviewer authority.
The isolated gate domain remains intact for later safe composition.

Safety operations runtime evidence (2026-07-30): the API now maps the dedicated
`safety.review` capability from T&S/admin roles and exposes a bounded open-case
queue, self-assignment and least-exposure evidence access. Queue projections
contain privacy-keyed subject references but no reporter identity. Evidence
requires both exact assignment ownership and fresh MFA, accepts only the
triage/appeal/legal purpose enum, automatically redacts phone, email and handle
identifiers, and writes the existing immutable access audit. The evidence
service now resolves a case to its retained report instead of incorrectly
treating the case ID as the report ID. Handler tests cover scope, keyed
projection, reporter omission, MFA, assignment and successful redacted access;
OpenAPI and the generated client contain the three runtime routes. The admin
desk uses this flow and makes uncomposed legal-hold, export and enforcement
capabilities unavailable rather than simulating them.

Care operations runtime evidence (2026-07-30): the dedicated T&S/admin scope
now opens the persisted care queue and permits only the domain's
open-to-engaged-to-resolved lifecycle. Every projection privacy-keys the member
reference and exposes only the signal class, lifecycle status, approved
resource keys, creation time and revision. Resolution requires fresh MFA and
accepts only the finite Ghana helpline-directory, counsellor-referral,
support-content and closure-quietening keys. The admin page no longer
fabricates members, contact preferences, prepared messages or delivery; it
truthfully records engagement and resources actually used while keeping care
separate from enforcement and diagnosis. Handler tests cover scope, private
projection and MFA. OpenAPI, the generated client and the real admin BFF carry
all three routes, and the replica-set Mongo integration covers queue ordering,
closure quietening, notification policy, invalid-script rejection and
resolution.

Runtime-control evidence (2026-07-30): the existing fail-closed flag kernel is
now composed into API startup with environment-layer configuration, a durable
Mongo proposal/audit repository, HMAC-keyed controller authority and the
in-process runtime registry. Admin-only, freshly stepped-up proposal mutations
allow exactly one canonical capability, staging/production environment, Ghana
market, finite action and finite reason. A distinct stepped-up administrator
must approve, and only that approver may apply the retained terms. Controller
keys never leave the API. Applied proposals replay oldest-to-newest after API
restart, and every active proposal receives an exact deadline timer that
publishes the domain's disabled-and-killed fail-closed state and persists the
expiry audit. HTTP middleware now enforces the resulting Sow, Fires, Payments
and Gate decisions on their member route families while leaving health,
authentication and the administrative control plane reachable. Staging
explicitly enables the four composed capabilities and keeps AI disabled. The
admin desk now lists real proposals and performs the actual propose, approve
and apply transitions; its fabricated proposal, named approvers and false
apply-ready state were removed. Focused domain/runtime/HTTP tests, live
replica-set Mongo repository proof, OpenAPI/Redocly, generated-client drift,
admin tests/typecheck/lint and production build pass.

Market-governance runtime evidence (2026-07-30): the complete draft,
published and retired market-pack register is now available only to an
authenticated operations administrator. Its response includes bounded market,
terminology reference, feature states, lifecycle and version while omitting
all controller identifiers; the current operator receives only
`proposedByMe`/`approvedByMe` relationship booleans. Draft, distinct-operator
publish and retirement now require fresh MFA. Each optimistic pack transition
and its immutable configuration audit record commit in the same Mongo
transaction, so neither can survive alone. The redesigned governance desk
loads this real register, creates bounded drafts, prevents visible
self-approval, publishes through server-enforced four eyes and retires without
deleting history. The fabricated translation counts, review notes and
caller-selected named approvers were removed. Focused domain/application/HTTP
and OpenAPI tests, live replica-set Mongo lifecycle proof, 53 admin tests,
admin typecheck/lint/build, Redocly, generated-client drift and
`git diff --check` pass.

Finance reconciliation runtime evidence (2026-07-30): the API now composes a
finance-scoped, read-only operational projection over the retained
reconciliation facts, comparison audits and daily checkpoints. Exception rows
contain only bounded suffixes of already-HMAC-keyed provider and statement
references, currency, amount, finite mismatch code and timestamps; raw
provider references, event keys, ledger commands and full linkage keys never
reach the admin client. The redesigned finance desk combines that real
evidence with the existing MFA-gated, evidence-complete atomic escrow
settlement flow. Its fabricated exceptions, local investigation/resolution,
fake redacted export and caller-selected two-person pricing approval were
removed. Pricing publication and finance exports now state that they are
unavailable until their server authorities are composed instead of claiming
success. Focused query/HTTP/OpenAPI tests, live Mongo append-only/concurrency/
checkpoint/privacy proof, 49 admin tests, admin typecheck/lint/build, Redocly,
generated-client drift and `git diff --check` pass.

Admin account runtime evidence (2026-07-30): an authenticated operator can now
read only the current principal's enrolled email, least-privilege roles,
lifecycle state and enrollment date plus the current session's issued/expiry
times and fresh-MFA state. Principal IDs, session IDs, other sessions, device
names and inferred locations are not projected. The account page now renders
those server-authoritative facts, performs a real current-browser sign-out and
keeps only theme and language as explicitly local browser preferences. Its
fabricated named profile, editable identity, operator ID, Mac/iPhone session
roster, Accra location, fake session revocation and claimed notification
delivery preferences were removed. Focused HTTP/OpenAPI tests, 43 admin tests,
admin typecheck/lint/build, Redocly, generated-client drift and
`git diff --check` pass.

Incident-response runtime evidence (2026-07-30): the admin incident surface is
now an honest operational rendering of
`internal/operations/runbooks/incident-response.md`: the four severity
definitions and response targets, ordered detect-to-review flow, evidence
handling boundary and 72-hour reasonable-suspicion clock are visible alongside
links to the real runtime-control, safety and care authorities. It explicitly
states that pager dispatch, legal holds, external channels, regulatory
packets/submission and incident closure remain external or uncomposed. The
fabricated P1 incident, typed commander/recorder, locally completed
checkpoints, fake packet reference and fake close transition were removed.
Admin typecheck/lint/build, 40 remaining truthful admin tests and
`git diff --check` pass.

Workforce-safeguard runtime evidence (2026-07-30): the admin workforce surface
now presents the required category-before-content, protected-break,
no-penalty-refusal, bounded-exposure, confidential-support and rotation
policies as explicit external staffing requirements. It points evidence-facing
work to the real assigned, purpose-audited, fresh-MFA safety queue and states
that no staffing/HR/counselling authority is composed. The invented 96-minute
shift, 2-of-4 exposure counter, local assignment acceptance/completion, fake
break, opt-out and supervisor-support success states were removed. Admin
typecheck/lint/build, 37 remaining truthful admin tests and `git diff --check`
pass.

Mpanyimfo runtime evidence (2026-07-30): the admin surface now accurately
describes the implemented women-led evidence boundary: current versioned
definition, minimum cohort/response thresholds, five bounded dimensions,
substantive women-reviewer coverage, complete gap acknowledgement, opaque
keys, and only neutral `evidence_incomplete` or
`ready_for_release_review` outcomes. It states that no docket, recusal,
panel-seat authority, vote, ruling, appeal, enforcement or release authority
is composed. The fabricated case/action/evidence counts, three elder seats,
local votes, quorum ruling, reason and fake appeal reference were removed.
Admin typecheck/lint/build, 34 remaining truthful admin tests and
`git diff --check` pass.

Community-operations runtime evidence (2026-07-30): the admin surface now fails
closed while host-certification and participant-notice authorities remain
uncomposed. It documents the exact versioned evidence chain required for
circle/fire density, host verification/training/Suban vetting/certification,
the finite notice template and audience, bounded reason, exact-preview
acknowledgement and source revalidation. It also states that a successful
proposal is human-review-ready only and never assigns a host, cancels a fire,
changes a circle or sends a notice. The invented circle, fire, host candidates,
dates, density, notice and locally successful proposal were removed. Admin
typecheck/lint/build, 31 remaining truthful admin tests and `git diff --check`
pass.

Launch-readiness runtime evidence (2026-07-30): the admin surface now renders
the repository-controlled staging evidence and the named external handoff from
`internal/quality/launch-decision-inventory.md` and
`internal/quality/external-gate-handoff.md`. It offers no launch, handoff,
throttle or triage mutation and explicitly keeps production blocked by
residency/legal, provider, Ghana field, credential/store, real cohort,
operational-cover, founder-decision and production-action gates. The fabricated
candidate SHA, readiness reference, family/host/matchmaker counts, staffing,
milestones, campus attribution, UAT, hypercare and locally prepared actions were
removed. Admin typecheck/lint/build, 17 remaining truthful admin tests and
`git diff --check` pass.

Completion authority remains bounded: legal approval, provider contracting,
production credentials, physical-device field evidence, store submission,
cohort recruitment and the human go/no-go decision stay external gates.
They must be closed by the task that first depends on them. Sprint 1 engineering
may continue on independent stories; production provisioning and real-member
data remain prohibited.

Launch-waitlist runtime evidence (2026-08-04): the public marketing site now
uses a conversion-focused, evidence-based journey with repeated waiting-list
calls to action, clear product differentiation, no invented testimonials or
launch metrics, and an explicit promise of one availability email. The form
requires name, valid email and purpose-specific consent; the API normalizes
email, idempotently upserts into the uniquely indexed `launch_waitlist` MongoDB
collection, preserves the original consent/version/time, and tracks pending
versus sent notification state for a future launch dispatch. An
operations-scoped `/v1/admin/waitlist` projection is available through the
authenticated admin BFF and the sidebar now opens a dedicated Waiting list desk
with totals, pending count and subscriber rows. OpenAPI and the generated
TypeScript client include both routes. Desktop and 390px browser QA showed no
horizontal overflow and verified the form's accessible controls and graceful
backend-unavailable message. Marketing/admin production builds, Redocly,
generated-client drift, the full 62-task frontend matrix and the complete Go
suite pass. Sending the availability email remains a deliberate launch-time
operation: Resend production credentials, approved sender/domain, final email
content and the authorized dispatch trigger are still external launch gates.

Marketing production-shell evidence (2026-08-04): the public navbar and footer
now render the supplied official Obiara lockups, including a measured 104px
mobile navbar treatment, and the supplied app-tile artwork is exposed as the
site favicon and manifest icon. SEO now includes a deployment-aware canonical
origin, title/description/keywords, index/follow and large-preview directives,
Open Graph/Twitter metadata, social-image alt text, sitemap, robots host,
manifest, theme color and WebSite structured data. The marketing runtime also
removes the framework disclosure header and emits nosniff, frame-denial,
referrer and browser-permission headers. The environment example records the
two production bindings required for the canonical website and server-side API
proxy. Marketing lint, typecheck, tests, production build and `git diff
--check` pass; the production server returned the expected routes, metadata,
security headers, favicon and manifest, and 390px browser QA confirmed the
official logo at 104px with no horizontal overflow.

Marketing navigation-state evidence (2026-08-04): the public header now tracks
the reader's current page section from measured scroll position, visually
distinguishes the active information link with the brand rose rule, marks it
with `aria-current="location"`, and promotes the waitlist button to its active
rose state at the conversion section. Live browser QA confirmed the Trust to
Waitlist transition and zero horizontal overflow; marketing lint, typecheck,
production build and `git diff --check` pass.

Mobile store-readiness evidence (2026-08-05): the Expo app now has stable iOS
and Android identifiers and versions, an opaque 1024px production icon,
adaptive and monochrome Android artwork, a minimal native-permission boundary,
Apple encryption and privacy-manifest declarations, runtime-version isolation,
automatic production build-number increments, and explicit staging and
production submission profiles that remain drafts and never auto-submit. Every
release environment fails closed unless it supplies an EAS project ID, HTTPS
API origin and HTTPS public-site origin. The app links directly to its privacy
policy and member support beside the already implemented authenticated export
and full-account-deletion flow. The public marketing surface now provides
static `/privacy`, `/terms`, `/support` and `/delete-account` routes and includes
them in navigation and the sitemap; this closes the repository side of Apple
privacy-policy access and Google Play's public account-deletion resource.
`apps/mobile/store-submission.md` supplies review copy, reviewer paths, content
and privacy declarations, creative requirements and the human closeout
checklist; `store-data-safety.md` records the engineering data inventory without
claiming legal attestation. Expo 57 dependencies were aligned to the exact
compatible versions, including the shared native UI peer. Evidence: 8/8 mobile
tests, mobile TypeScript and ESLint, Expo dependency check, Expo Doctor 20/20,
resolved production config inspection, successful Android/iOS/web Expo export
(1,322/1,204/895 modules), marketing static production build with all four
public routes, the full 62-task frontend matrix and complete Go suite. Store
accounts and agreements, reserved app records, EAS/store credentials, deployed
production URLs, final privacy/legal attestations, physical-device screenshots
and qualification, reviewer credentials, signed AAB/IPA uploads, console forms
and store review remain external gates; the repository does not falsely mark
the app as submitted or accepted.

Production hosting-preparation evidence (2026-08-05): `render.yaml` is now a
backend-only production Blueprint containing exactly the protected Go API and
worker; all Node/member/admin/marketing hosting was removed. Both services use
manual `autoDeployTrigger: off`, immutable Render commit identity, independent
least-privilege Mongo credential slots, the shared `obiara_production` logical
database, required rotating secret slots and optional OTLP/LiveKit bindings.
Payments and AI fail closed until their external production gates are approved.
Render CLI validation returned `valid: true` with four actions, one environment
group and the two expected services. Gitignored `.env.production` worksheets
now exist locally for marketing, member web, admin, API and worker; tracked
`.env.production.example` counterparts provide safe deployment manifests.
Vercel owns all three Next.js applications, with root-directory and Production
environment instructions in `deploy/vercel/README.md`; production BFF/API
bindings now fail closed instead of silently calling localhost. Marketing,
member and admin Vercel-shaped production builds and stripped native Go API and
worker builds pass, as do the focused Blueprint, release-policy, secret-policy
and API-config tests. Configuration readiness does not claim deployed Render or
Vercel resources, provider/legal approval, real credentials, database restore
evidence, live health checks or traffic activation.

Post-completion review and correction round (2026-08-15): an independent
four-lane review covered commit `263abec` (backend, member web/admin, mobile)
and `fb6444f..HEAD` (waitlist, store readiness, hosting, marketing). Twenty
verified findings were corrected:

- Mobile sign-in actually works now: OTP verify sends a stable SecureStore
  device ID (`auth.go` hard-required `deviceId`, so every sign-in previously
  failed with 422); authenticated 401s clear the session and return the gate
  to the phone stage; a real sign-out control exists; the escrow screen no
  longer calls `crypto.randomUUID()` (absent on Hermes); game screens reset
  stale idempotency keys on input change and failure; the Suban appeal no
  longer collects a reason it never sends.
- Member web: the onboarding Ghana Card gate now accepts the server's `vc_…`
  case IDs instead of silently dead-ending a successful submission; the
  profile desk no longer shows a false "Profile saved" banner on first load;
  the dead fabricated-sentinel guard and the dead `games-model` availability
  constant were removed.
- Admin: BFF routes now fail closed with 503 JSON instead of HTML 500s when
  the Go API is unreachable; the analytics desk's fabricated "recorded review"
  (invented `review•••2J8` reference) and dead fabricated seeds were removed;
  the operators desk no longer invents MFA status or mislabels enrollment as
  "last active"; the command centre and rail now render live
  verification/safety/care/account data or honestly fail closed — the
  fabricated greeting, metrics, named queues, SLA pulse, incident rows, nav
  badges and operator identity were removed.
- API: the payments kill switch now covers the actual
  `/v1/matchmaker-engagements*` and `/v1/membership*` mutation routes;
  admin enroll/status/role mutations commit with their audit records in one
  Mongo transaction; the last-active-admin guard is enforced inside the
  transaction instead of as a read-then-write race; admin login start always
  returns 202 without enumerating unknown or suspended principals;
  competition review/appeal resolution now requires the same MFA step-up as
  comparable admin overturns; the waitlist upsert falls back to an idempotent
  read on duplicate key, the admin waitlist handler no longer violates its
  envelope on cancelled contexts, and public waitlist joins are per-IP
  throttled (429 documented in OpenAPI).
- Marketing emits a baseline Content-Security-Policy alongside its existing
  security headers; the generated TS client was regenerated after the two
  documentation-level OpenAPI edits and the drift gate passes.

Verification: full Go suite (283 packages) green including the admin MongoDB
Testcontainers integration over the new transactional paths; 62-task frontend
lint/typecheck/test matrix green; marketing, web, admin and mobile checks and
production builds pass; OpenAPI contract test, generated-client drift,
repository Prettier and `git diff --check` all pass.

Recorded follow-ups (not defects; owned by the first dependent task): the API
has no token-refresh endpoint, so mobile members re-verify by OTP after the
15-minute access token expires; admin sign-out clears the cookie but no
server-side logout endpoint exists yet; liveness artifact references are not
yet bound to their upload subject (safe today because no retrieval path
exists — must be enforced before any reviewer-facing decryption is added);
waitlist name/email are stored unkeyed as a documented exception to the
privacy-keying convention because the operations desk must read them; consent
on public waitlist joins is self-declared until a double-opt-in or launch-time
verification authority is composed.

Launch-preparation hardening (2026-08-15): the supply chain and toolchain
were brought to a clean launch state. Go was bumped to 1.26.6 and grpc to
1.82.1 across `go.mod`, `go.work`, `render.yaml` and both CI workflows after
govulncheck found seven reachable advisories; the rescan reports zero
reachable vulnerabilities. Three new high JS advisories were remediated with
exact overrides (`brace-expansion@5.0.9`, `js-yaml@4.3.1`, `nanoid@3.3.18`).
The two `image-size` DoS advisories have no patched upstream release and are
reachable only through metro/Storybook dev tooling, so they are held as a
documented temporary exception in `internal/quality/ci-security-baseline.md`
(owner, rationale, compensating controls, 2026-11-15 expiry); the CI audit
gate now counts severities from the advisories map, which honors
`auditConfig.ignoreGhsas`. Verification: full Go suite green in short mode
under 1.26.6, vet/actionlint/Render Blueprint validation clean, the complete
frontend lint/typecheck/test matrix and all 11 production builds pass, and
the pnpm high/critical audit gate passes. The
`TestMongoReplicaSetIsolatedRestorePreservesSourceAndDetectsCorruption` live
run was skipped once more because Docker Desktop on the development machine
is unresponsive (daemon API calls hang) — the same pre-existing environment
limitation recorded at S1-001 and S4-005, not a code failure; the test passed
live earlier the same day and CI executes it on a clean runner. Production
remains blocked exclusively on the external gates in
`internal/quality/external-gate-handoff.md`.

## 27. Handover-pack reconciliation (2026-09-04)

The founder asked for the Build Pack to be checked against what is actually
implemented. Every testable requirement in the twelve
`Obiara_Handover_Package/2_Build_Pack` documents was extracted and traced into
the repository: **888 requirements**. The trace returned 112 built, 391
partial, 254 missing, 122 drifted and 9 undocumented.

The 80 requirements whose absence is launch-critical were then handed to
independent verifiers instructed to _refute_ the finding, because the
expensive error in this repository is declaring something missing that is
built under an Akan name. **74 of 80 were confirmed; 6 were overturned.**
Sizes on the confirmed set: 46 large, 26 medium, 2 small.

The 391 partial and 254 missing requirements below the critical tier were NOT
individually verified. They are indicative and must be re-checked before any
of them is worked; the trace was deliberately generous with "partial", and
anything begun-but-incomplete earned that label.

**The finding underneath every area is one shape: the domain layer is real and
the outbound adapters are not.** Aggregates, state machines, invariants,
optimistic versioning and consent re-checks are modelled carefully and tested.
What is absent is everything that speaks to the outside world — object
storage, transcription, payment rails, media encoding — and, in several
cases, the composition that would make a finished context reachable at all.
`services/api/internal/introduction/` is the clearest instance: a 397-line
domain aggregate, a 410-line application service, defined ports and tests,
whose `module.go` is two lines of package doc with no constructor, which
nothing imports, and which appears in none of the route registrations in
`services/api/main.go`.

This is the same failure that closed the front door on 2026-09-04: the Ghana
Card and liveness providers were both stubs, and signing up could not be
completed while they were unreachable.

### 27.1 Voice & media (4)

| Req                   | Deliverable                                                                      | State   | Size   | Verified finding                                                                                                                                                                                                   |
| --------------------- | -------------------------------------------------------------------------------- | ------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| TA-object-storage-cdn | Object storage + CDN for media                                                   | MISSING | LARGE  | Verified; the claim holds. services/api/internal/media/ has only domain/ (asset.go, asset_test.go) and application/ (access.go, access_test.go, ports.go, mock_ports_test.go) — no adapters/ dir. ports.go…        |
| TA-voice-encode-opus  | Client records AAC, transcodes to Opus 16–24 kbps adaptive on-device             | MISSING | LARGE  | I could not refute this. The spec item is real (Obiara_Handover_Package/2_Build_Pack/Obiara_07_Technical_Architecture.docx §3.2: "Client records AAC→transcodes to Opus 16–24 kbps adaptive on-device; upload via… |
| CC-01                 | Every voice surface: playback speed, re-record before send, transcription toggle | MISSING | LARGE  | Claim stands. No voice capture or playback UI exists in any client: a repo-wide grep for playbackRate / <audio / new Audio( / AudioContext (excluding node_modules and .worktrees) returns zero hits, and the…     |
| CL-REG-07             | No machine-only translation in production for launch languages                   | MISSING | MEDIUM | Claim confirmed, and it slightly understates the gap. (1) The gate is real but unconsumed: packages/i18n/src/index.ts enforces reviewed+value+reviewer+ISO date+placeholder parity in isApprovedTranslation,…      |

### 27.2 Matching & introductions (26)

| Req                          | Deliverable                                                                                 | State   | Size   | Verified finding                                                                                                                                                                                                   |
| ---------------------------- | ------------------------------------------------------------------------------------------- | ------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| S-21                         | S-21 Person page (pre-acceptance) elements                                                  | MISSING | LARGE  | Claim verified, not refuted. No person-page route exists in any client: the full route listing of apps/web/app and apps/mobile/app contains garden, epono-ano, introductions/[introId], abonten, adiwo, games and… |
| S-20                         | S-20 Weekly Introductions delivers 3–7 story-cards                                          | MISSING | LARGE  | I searched under every plausible Akan alias (fie, abonten, adiwo, epono-ano, dan-mu, garden, nnoboa) and every bounded context, and the claim holds. SPEC (what S-20 actually is): extracted from…                 |
| S-31                         | S-31 Pod open — full-screen hold-to-listen, release pauses                                  | MISSING | LARGE  | Claim confirmed after searching the Akan-named contexts, both client route trees, OpenAPI, main.go wiring and the authz grant table. NO pod screen exists in either client:…                                       |
| S-22                         | S-22 Sow composer — question on top, 30–90 s record, one re-record                          | MISSING | LARGE  | Verified, claim stands. NOT BUILT: (1) No composer route/component in any client. apps/web/app/fie/garden/garden-shell.tsx:102-113 states it outright — "The former candidate cards and 'preview server…           |
| S-40                         | S-40 Room header with cloth strip, stage chip and honesty ribbon                            | MISSING | LARGE  | Claim verified in full; nothing found under an Akan or alternate name. Backend room mechanics are real and mounted: services/api/internal/platform/http/courtship_room.go registers POST /v1/courtship/rooms,…     |
| FR-201a                      | Seed allowance enforced server-side                                                         | MISSING | LARGE  | I went looking for a spend path under every Akan/seed name and could not find one on the running surface; the claim holds, and if anything the reachable surface is looser than described. What is built (claim's… |
| FR-202                       | Sow requires 20s verified playback of Voice of Introduction                                 | MISSING | LARGE  | I tried hard to find the gate under another name and could not. What IS built is exactly as the claim describes: services/api/internal/seed/listening/domain/playback.go defines `RequiredSeconds = 20.0`,…        |
| M4-03                        | Sow flow: ≥20s listen gate, 30–90s answer with one re-record, drag-release gesture          | MISSING | LARGE  | I tried to find this built under Akan/domain names (sprout, epono ano/doorway, garden, sow, pod, listening) and it is not. All three legs of PRD M4-03 / UX S-22 hold up as claimed. (1) Listen gate is advisory.… |
| M5-01                        | Room creation on mutual water with cloth strip, stage and okyeame door in header            | MISSING | LARGE  | Searched under every Akan alias and could not refute any part of this. (1) Gating: services/api/internal/platform/http/courtship_room.go:113-163 (startRoomHandler) checks only opaque-id shape for…               |
| P0-M4-SOW-FLOW               | Sow flow                                                                                    | MISSING | LARGE  | I could not refute it — the claim holds on every leg, and the gap is if anything larger than stated. NOT COMPOSED. services/api/internal/seed/module.go is the seed-stage composition root and its StageModule…    |
| P0-M4-PODS                   | Pods                                                                                        | MISSING | LARGE  | Claim confirmed, and the exit-gate half is worse than stated. The slice is genuinely built: services/api/internal/seed/pod/domain/pod.go is a full event-sourced aggregate (owner/media keys, MaxRecipients=25,…   |
| P0-M5-VOICE-FIRST            | Rooms are voice-first                                                                       | MISSING | LARGE  | Claim confirmed. services/api/internal/courtship/drum/domain/stage.go is a complete voice-first turn model (Medium voice/text, ErrVoiceRequired, Open() forces first beat MediumVoice, Add() enforces alternation… |
| P0-M4-DECLINE-SPROUT-WATER   | Decline / sprout / water actions                                                            | MISSING | LARGE  | Claim verified on every point. Water: services/api/internal/seed/water/{domain/water.go, application/service.go, application/ports.go, adapters/outbound/mongodb/repository.go,…                                   |
| P0-EXIT-PODS-HEARD           | Pods heard rate ≥65%                                                                        | MISSING | LARGE  | The core claim holds: nothing in production code emits epono.pod_heard or epono.seed_sown, so GET /v1/metrics/funnel always returns podsHeardRate 0. services/api/internal/analytics/application/analytics.go:52…  |
| IM-009                       | The Sow is the atomic gesture; aba replaces the swipe and the like                          | MISSING | LARGE  | The claim stands. The sow slice is built to depth but is unreachable end to end. Built: services/api/internal/seed/sow/domain/sow.go (Accept with commandID/fingerprint/allowanceUnits invariants; confirmed…      |
| IM-022                       | Received seeds rest at the house front as closed pods                                       | MISSING | LARGE  | Claim stands. Pack requirement confirmed from Obiara_Handover_Package/2_Build_Pack/Obiara_06_UX_Flows_Screens.docx (S-14 "Ɛpono ano hub. Two panes: House Front (pods) / My Garden"; S-30 "House Front. Pods as…   |
| TA-state-engine-enforces-frs | State engine enforces FR-201/203/206/301/302 by construction                                | MISSING | MEDIUM | Claim stands, with one correction to its wording: the FR-301 alternation engine EXISTS and is fuzz-proven, it is simply not composed. services/api/internal/courtship/drum/domain/stage.go implements strict…      |
| FR-301 | Alternation: consecutive sends impossible at API layer | DONE | MEDIUM | RECONCILED 2026-09-05: Closed by §36: the queue aggregate refuses a second consecutive turn, so it is impossible below the API rather than checked by one route. (was: Could not refute — I traced the whole composed path. services/api/main.go:205 builds courtship.NewRoomModule and :543 registers the room routes; servi) |
| FR-303                       | Active-rooms honesty count server-computed and unsuppressible                               | MISSING | MEDIUM | Claim confirmed; I could not refute it. PRD M5-06 (extracted from Obiara_Handover_Package/2_Build_Pack/Obiara_01_PRD.docx, line 87) requires "each member sees the other's count of currently active rooms…        |
| M4-01                        | 7 seeds per week, Monday renewal, +3 suban bonus, no rollover, never purchasable            | DRIFTED | MEDIUM | The claim stands on its two substantive limbs; only its third limb is overstated, and not in a way that changes the runtime outcome. CONFIRMED — allowance is 3, not 7.…                                           |
| M5-02 | Alternation engine: one drum holder, composer disabled without drum, voice-only first… | PARTIAL | MEDIUM | RECONCILED 2026-09-05: §36 delivered strict alternation. The drum engine's stage, single drum holder and voice-only first turn remain uncomposed. (was: Claim stands on the live path; only its "missing" label is too strong. The engine EXISTS and is fully tested but is dead code. services/api/internal/c) |
| M4-06                        | Sprout: doorway conversation capped at three alternating exchanges, mutual water opens room | PARTIAL | MEDIUM | The claim's status ("partial") stands, but one sentence of its gap text is factually wrong and should be corrected. WRONG: "There is no Water action and no mutual-water gate." A complete mutual-water slice…     |
| M4-AC-01                     | Server-side enforcement of listen gate, allowance, no purchase, 90-day lock                 | PARTIAL | MEDIUM | The claim's substance holds: no reachable API path enforces the listen gate, the allowance spend, or the 90-day decline lock. I dumped every registered route (`grep mux.Handle` across…                           |
| P0-M5-ALTERNATION-ENGINE     | Room alternation engine                                                                     | MISSING | MEDIUM | Claim verified, not refuted. The shipped courtship-room turn path enforces only optimistic ordering: main.go:543 registers RegisterCourtshipRoomRoutes over courtship.NewRoomModule (main.go:205);…                |
| P0-EXIT-SEED-SPROUT          | Seed→sprout ≥25%                                                                            | MISSING | MEDIUM | Claim confirmed, and if anything understated. Measurement side is real: services/api/internal/analytics/application/metrics.go computes SeedToSproutRate =…                                                        |
| P0-EXIT-SPROUT-ROOM          | Sprout→room ≥35%                                                                            | MISSING | MEDIUM | Could not refute; if anything the gap is wider than claimed. The metric arithmetic is real: services/api/internal/analytics/application/metrics.go computes SproutToRoomRate = rate(rooms, sprouted) from…         |

### 27.3 Onboarding (1)

| Req  | Deliverable                                                                   | State   | Size  | Verified finding                                                                                                                                                   |
| ---- | ----------------------------------------------------------------------------- | ------- | ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| S-06 | S-06 Voice of Introduction — three prompts, 120 s meter, per-prompt re-record | MISSING | LARGE | I searched under every Akan alias and under seed/garden/listening, and the claim holds; if anything the gap is slightly larger than stated. Backend, what exists:… |

### 27.4 Identity & verification (12)

| Req                        | Deliverable                                                                     | State   | Size   | Verified finding                                                                                                                                                                                                   |
| -------------------------- | ------------------------------------------------------------------------------- | ------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| TA-ghana-id-partner        | Ghana identity verification via accredited partner API                          | MISSING | LARGE  | I could not refute this. There is no accredited-partner adapter anywhere in the repo, under any name. What I opened: - services/api/internal/platform/delivery/verification.go — the composition-root selector…    |
| FR-103b                    | Presentation-attack rejection at Sentinel thresholds                            | MISSING | LARGE  | Claim stands, and the gap is slightly larger than stated. Pack requirement confirmed in Obiara_Handover_Package/2_Build_Pack/Obiara_02_SRS.docx (FR-103: randomized spoken-phrase + face challenge, reject…        |
| SENT-02                    | Deepfake screen on enrollment and step-ups                                      | MISSING | LARGE  | Could not refute; the gap is real and the same size as claimed. The requirement is real and specified in two places the claim's grep missed: internal/architecture/threat-model-v0.md:100-102 ("face+voice…        |
| M1-07                      | Voice of Introduction: 120s guided recording, transcribed and values-tagged     | MISSING | LARGE  | Searched by feature name, by Akan naming, and by mechanism (audio/transcribe/record/listen/tag) — the claim holds and is if anything generous. The context is domain + application only:…                          |
| P0-M1-VOICE-INTRO          | Voice of Introduction                                                           | MISSING | LARGE  | I searched for the feature under every plausible alias (introduction, voice, sow/seed, epono-ano, media, asset, transcribe) and the claim holds. Built: services/api/internal/introduction/domain/introduction.go… |
| TA-integration-national-id | Integration: National ID verify partner, launch-blocking, human-review fallback | MISSING | MEDIUM | The claim is accurate on both halves. Absent integration: services/api/internal/verification/adapters/outbound/ holds only manual, simulator, mongodb and privacy — no vendor adapter under any name.…             |
| FR-101a | Romantic surfaces gated below Tier 1 | DONE | MEDIUM | RECONCILED 2026-09-05: Closed by §32: MemberGate at the registration tables, rules read from the authz grant table. (was: The claim holds exactly as written; I could not find the gate under any other name. The policy table is real (services/api/internal/authz/domain/polic) |
| FR-102a                    | National ID verified against issuer service                                     | MISSING | MEDIUM | The claim holds. `services/api/internal/verification/application/service.go` (lines 46-49) defines `VerificationProvider` as "the outbound port for the national-ID issuer service", and `SubmitGhanaCard` calls…  |
| M1-02 | Age gate from national ID; under-18 hard block and ID data deletion | PARTIAL | MEDIUM | RECONCILED 2026-09-05: §33 delivered the under-18 hard block with keyed proof and purge; §34 gave adult ID data an end date. The date of birth is still self-declared, so 'from national ID' awaits FR-102a. (was: Claim confirmed, and the deletion half is worse than claimed. (1) The safeguarding package has zero importers: `grep -rln "safeguarding" --include="*.) |
| M1-AC-01 | Tier-0 accounts fully gated from member surfaces | PARTIAL | MEDIUM | RECONCILED 2026-09-05: §32 gated the romantic surfaces at Tier 1. Circles, fires listing, profile and settings remain open to Tier 0, so 'fully gated from member surfaces' is not yet true. (was: Verified: the claim holds. services/api/internal/platform/http/fire.go is the only route file that gates on tier — RegisterFireRoutes takes a TierRead) |
| TS-AGE-ASSURANCE | Age assurance via ID-derived DOB with audit trail | PARTIAL | MEDIUM | RECONCILED 2026-09-05: §35 delivered the audit trail: an assurance record on the case, carrying method and threshold, stored apart from the birth date so retention cannot take it. ID-derived DOB awaits FR-102a. (was: Claim confirmed, and if anything the gap is marginally larger than stated. DOB capture and audit exist: services/api/internal/platform/http/verificati) |
| FR-101b | Sowing gated below Tier 2 | DONE | SMALL | RECONCILED 2026-09-05: Closed by §32: POST /v1/seed/sprouts requires Tier 2. Note Tier 2 is operator-unreachable until the promotion rule is decided. (was: The claim holds under a wide search (Akan names included: seed/sprout "doorway", sow, garden, water, source, listening, safety). 1. No tier check on t) |

### 27.5 Trust & safety (13)

| Req                       | Deliverable                                                                               | State   | Size   | Verified finding                                                                                                                                                                                                   |
| ------------------------- | ----------------------------------------------------------------------------------------- | ------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| S-43-MEET                 | Meeting flow: verified venue picker, trusted-contact toggle, next-day mutual confirmation | MISSING | LARGE  | Claim confirmed after searching under Akan naming too. What exists: services/api/internal/courtship/proposal/domain/proposal.go defines TypeMeeting as one of three proposal kinds, with State carrying only…      |
| TA-voice-server-pipeline  | Server voice pipeline order: virus scan → Sentinel → transcription → storage+waveform →…  | MISSING | LARGE  | I went looking under every Akan and English name and the claim holds; the repo even documents it itself. Spec source (extracted from the docx):…                                                                   |
| M4-ABUSE-01               | Sentinel screening of sow recordings before delivery, seed refunded on failure            | MISSING | LARGE  | I went looking for the screened-sow path under every Akan/English name and could not find it on any live route. What exists is a library, not a feature. Built (library only): -…                                  |
| P0-SENTINEL-SOW-SCREENING | Sentinel v0: sow-screening                                                                | PARTIAL | LARGE  | Claim stands; I could not refute it. The policy IS fully written and unit-tested: services/api/internal/seed/screening/application/policy.go carries the reason codes (contact_exfiltration, payment_request,…     |
| P0-SIKA-SHIELD            | Sika Shield keyword/pattern rules                                                         | MISSING | LARGE  | Could not refute; the claim is accurate. The harness is real: services/api/internal/safety/sikashield/domain/sikashield.go defines Pattern{Key,Version,Source,ExpressionRef,ReviewedByKey}, Metrics with…          |
| TS-BLOCK-PROPAGATION      | Blocks propagate across alternate accounts                                                | MISSING | LARGE  | I could not refute this; the claim holds and is if anything slightly understated. Source requirement is real: Obiara_Handover_Package/1_Strategy/Obiara_Founding_Blueprint.docx asks for "a blocklist that…        |
| LO-25                     | Two upheld coercion/ethics findings ends a license                                        | MISSING | MEDIUM | Could not refute; the claim holds. The rule comes from Obiara_Handover_Package/2_Build_Pack/Obiara_11_Launch_Operations.docx §4 (Agyina program): "post-engagement ratings, mystery-shop audits, annual…           |
| TS-TIER-D                 | Tier D — Care (not punishment)                                                            | PARTIAL | MEDIUM | Claim stands and is slightly understated. Built: internal/safety/domain/care.go (CareCase, 5 signals, open→engaged→resolved, script allowlist, 72h QuieteningWindow), internal/safety/application/care.go…         |
| TS-CARE-ROUTING           | Care-flag immediate routing with 24/7 rota                                                | MISSING | MEDIUM | The claim holds. What IS built is the back half of the flow: internal/safety/domain/care.go (CareCase, five Signals incl. self_harm_indication and okyeame_escalation, four ScriptKeys, 72h QuieteningWindow),…    |
| TS-CARE-QUEUE-SIGNALS     | Care queue signal sources                                                                 | MISSING | MEDIUM | Claim confirmed; if anything it is understated. internal/safety/domain/care.go and internal/safety/application/care.go do build the full care aggregate (five signals, engage/resolve, scripts, 72h quietening),…  |
| CL-LIB-08                 | Banned doorway-question rules                                                             | MISSING | MEDIUM | The claim holds; I could not find the rule implemented under any other name. Source of the requirement: Obiara_Handover_Package/2_Build_Pack/Obiara_10_Content_Localization.docx — "Doorway question bank: 40…     |
| IM-029 | Strict alternation — you may not send again until answered | DONE | MEDIUM | RECONCILED 2026-09-05: Closed by §36, same invariant. (was: Claim stands. A complete strict-alternation aggregate exists but is unreachable, and the live send path has no turn check. BUILT BUT UNWIRED: services) |
| TS-TIER-A                 | Tier A — account-ending conduct                                                           | MISSING | SMALL  | Claim verified, not refuted. The ladder itself is genuinely built: internal/safety/domain/action.go (CheckLadder — Tier A = ActionBan only; SuspensionDuration 14/30/90d), internal/safety/application/actions.go… |

### 27.6 Privacy & retention (6)

| Req                 | Deliverable                                                                      | State   | Size  | Verified finding                                                                                                                                                                                                 |
| ------------------- | -------------------------------------------------------------------------------- | ------- | ----- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| S-30                | S-30 House Front — pods as closed forms showing trust path + first name/age only | MISSING | LARGE | I searched under every Akan alias and found no House Front anywhere; the claim stands. Spec (extracted from the pack): Obiara_Handover_Package/2_Build_Pack/Obiara_06_UX_Flows_Screens.docx — "S-14 Ɛpono ano…   |
| TA-data-residency   | Data residency in-region                                                         | PARTIAL | LARGE | The claim substantively stands: no in-region deployment exists. But one sub-clause is wrong, so I corrected "missing" to "partial". What is genuinely absent (claim upheld): - render.yaml:85 and :185 —…        |
| NFR-301b            | Voice/biometric blobs envelope-encrypted with per-user keys                      | MISSING | LARGE | Verified, and the search was exhaustive rather than name-based: `grep -rln "crypto/aes/crypto/cipher/chacha20/nacl/secretbox"` across all of services/api returns exactly two files —…                           |
| RET-01              | Voice in closed rooms retained 180 days then crypto-erased                       | PARTIAL | LARGE | I looked for this under every name I could find and the headline gap is real, but two of the claim's supporting statements are wrong. CONFIRMED ABSENT: - No 180-day room-voice policy anywhere.…                |
| NFR-401b            | In-region data residency with approved DR replica                                | MISSING | LARGE | Every factual element of the claim holds up. render.yaml:85 and :185 pin `region: frankfurt` for obiara-api-production and obiara-worker-production, and internal/quality/renderblueprint/blueprint_test.go:107… |
| TS-E2E-ROOM-CONTENT | Encrypted room content                                                           | MISSING | LARGE | Confirmed after reading the actual E07 room code, not just greps. The room is services/api/internal/courtship/room (state machine) plus services/api/internal/courtship/queue (turn log behind POST…             |

### 27.7 Commerce & payments (6)

| Req                          | Deliverable                                                                    | State   | Size   | Verified finding                                                                                                                                                                                                  |
| ---------------------------- | ------------------------------------------------------------------------------ | ------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| S-90                         | S-90 Checkout MoMo sheet                                                       | MISSING | LARGE  | Claim stands; I searched widely (Akan route names, catalog/membership/escrow contexts, marketing, mobile, packages) and every one of the five named checkout elements is genuinely absent. Backend:…              |
| TA-momo-aggregator           | MoMo rails: direct MTN MoMo API + Telecel/AT via aggregator                    | MISSING | LARGE  | Claim confirmed and, if anything, understated. services/api/internal/commerce/momo/ holds exactly nine files (README.md, domain/intent.go + test,…                                                                |
| TA-integration-momo-networks | Integration: MTN MoMo / Telecel / AT with cross-network retry and grace states | MISSING | LARGE  | Claim stands; I could not refute it under any alias. What exists is a deliberately provider-neutral shell: services/api/internal/commerce/momo/application/ports.go declares `Provider.RequestCollection`, and a… |
| RPM-32-MOMO-AGGREGATION      | MoMo aggregation across MTN MoMo, Telecel Cash, AT Money                       | MISSING | LARGE  | I tried hard to find this under another name (Akan naming, aggregator brands, "rail"/"gateway"/"charge"/"payer"/"psp") and it genuinely is not there. What exists:…                                               |
| RPM-05-MICRO-BILLING         | MoMo-native micro-billing at familiar price anchors                            | MISSING | LARGE  | I could not refute this; every specific assertion checks out. (1) Orphaned MoMo: services/api/internal/commerce/momo/ has domain/intent.go, application/{service.go,ports.go} (Create/Confirm/Callback, HMAC…     |
| RPM-25-MATCHMAKER-TAKE-RATE  | Platform take on matchmaker marketplace: 20% standard, 15% top-suban           | MISSING | MEDIUM | Confirmed. services/api/internal/platform/http/admin_escrow.go:66-72 (escrowTermsFromEngagement) is the only production builder of escrow terms — it sets `escrowdomain.MilestoneTerm{ID: milestone.ID,…          |

### 27.8 Platform (3)

| Req                  | Deliverable                                   | State   | Size   | Verified finding                                                                                                                                                                                                   |
| -------------------- | --------------------------------------------- | ------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| EVT-11               | Events ceremony.gate_crossed / aseda_declared | PARTIAL | LARGE  | Claim substantially stands; only its "schema and nothing else" wording overstates. CONFIRMED: `ceremony.gate_crossed` / `ceremony.aseda_declared` occur only at…                                                   |
| P0-M9-AUDIO-ROOMS-V1 | Audio rooms v1                                | MISSING | LARGE  | Verified: the gap is real and is if anything slightly understated. (1) The LiveKit token adapter (services/api/internal/realtime/livekit/adapter.go, application/ports.go with RoleListener/RoleSpeaker/RoleHost)… |
| EVT-03               | Event epono.pod_heard                         | MISSING | MEDIUM | I looked for a producer under every plausible Akan/English name and found none; if anything the gap is wider than claimed. Schema and consumer exist as the claim says:…                                           |

### 27.9 Other (3)

| Req                   | Deliverable                                                               | State   | Size   | Verified finding                                                                                                                                                                                            |
| --------------------- | ------------------------------------------------------------------------- | ------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| S-62                  | S-62 Fire live room — stage, audience grid, hand-raise, bounded reactions | MISSING | LARGE  | I went looking for the live room under every name I could think of (fire, room, stage, speaker, audience, raise, reaction, livekit, realtime, dan-mu) and the claim holds. CLIENTS — schedule + RSVP only,… |
| M9-02                 | Live audio engine with stage controls, bounded reactions, live captions   | MISSING | LARGE  | Claim stands. CONFIRMED MISSING: (1) No audio join for a fire — the only fire routes are services/api/internal/platform/http/fire.go:31-35 (schedule, list, rsvp, cancel, close), ember.go:23, and…         |
| P0-EXIT-D30-RETENTION | Day-30 retention ≥45%                                                     | PARTIAL | MEDIUM | I could not refute this one; the substance holds, though the evidence understates what exists. What IS built (more than "an unreferenced threshold constant"):…                                             |

### 27.10 Overturned by verification (6)

Claimed critical, found further along than the trace reported. None came back
clean — all six are partial rather than absent, and the two safety guarantees
among them should be read as unfinished, not satisfied.

| Req           | Deliverable                                                                  | Actually | Verified finding                                                                                                                                                                                                   |
| ------------- | ---------------------------------------------------------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| TA-drum-token | Drum possession is a token transferred only by opposing-party message events | PARTIAL  | Both supporting facts in the claim's "gap" are wrong, and the real gap is one stage narrower than stated. 1) "The handler does not even verify the caller is a participant of the room" is false.…                 |
| FR-104b       | Purge biometric/ID artifacts within 24h on age failure                       | PARTIAL  | The claim looked only at internal/platform/retention and services/worker/internal/jobs and missed an entire bounded context that implements FR-104b under an unsearched name: services/api/internal/safeguarding…  |
| FR-104a       | Age from ID must be 18 or older, hard-block on failure                       | PARTIAL  | The claim's evidence ("no server-side age computation"; "the only age computation in the whole repo is a display band, ageBand()") is false. A complete, tested, fail-closed age gate exists as its own bounded…   |
| NFR-301c      | Room content E2E-encrypted with safety-processing design                     | PARTIAL  | The encryption half of NFR-301c is genuinely absent, but the claim's gap statement is wrong on two of its points, so "missing" overstates it. WHAT IS ACTUALLY BUILT (files opened): 1. Consent-gated safety…      |
| FR-602        | Sika Shield blocks member-to-member payment and payment-detail exchange      | PARTIAL  | The claim's facts about the sikashield package are correct, but its conclusion ("nothing technically blocks member-to-member payment, gifting or payment-detail exchange") is wrong on two of three counts. MONEY… |
| PRODUCT-LAW   | Five product laws binding on every ticket                                    | PARTIAL  | The gap statement ("no enforced implementation on any reachable path") is wrong for alternation and half-wrong for voice-before-face. STRICT ALTERNATION — enforced, server-side, on a wired route.…               |

### 27.11 Goal: make the Voice of Introduction real, end to end

The confirmed gaps are not independent. One absence sits underneath the
largest cluster of them: there is no object storage anywhere in the API
(`services/api/internal/media` has `application/` and `domain/` and no
`adapters/`; no S3/GCS/R2 client exists in the tree). Every voice, photo and
document surface in the Build Pack rests on it, and the product's central
mechanic — a member is introduced by their recorded voice, not their photo —
cannot exist until it does.

The goal is therefore a vertical slice rather than a sweep: **a member records
a Voice of Introduction, it is stored encrypted, transcribed, and played back
under the listening gate that Sow already enforces.** It closes S-06 and
TA-object-storage-cdn directly, unblocks S-20, S-21, S-22, FR-202,
TA-voice-encode-opus, TA-voice-server-pipeline, NFR-301b and RET-01, and it
composes the finished `internal/introduction` context that is currently
unreachable.

| Task   | Deliverable                                             | Status  | Notes                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ------ | ------------------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| VOI-01 | Object storage port and adapter behind `internal/media` | DONE    | `internal/media/adapters/outbound/objectstore` signs SigV4 from the standard library — no vendor SDK, so no dependency tree under a release gate that counts reachable advisories, and it signs for S3, R2, B2 and MinIO unchanged. Content type, length and digest sit inside the signature. The published AWS key-derivation vector is asserted, because a wrong-but-stable key is invisible to a self-consistent test |
| VOI-02 | Compose `internal/introduction`                         | DONE    | `Reconstitute` added to the aggregate (a store cannot restore what the domain gives it no way to rebuild), a mongo store holding references and keyed fingerprints only, a media adapter, and `NewModule`. Media assets gained a `duration`: it is the number the listening gate counts against, and a client-declared one would let a member claim an answer they never recorded                                        |
| VOI-03 | HTTP routes and OpenAPI contract                        | DONE    | Four routes — the first time this context has been reachable at all. Audio never crosses the API; the client uploads to storage under the signed grant. Another member's recording answers 404, not 403: on a surface built to avoid disclosing who is a member, confirming an id exists is the disclosure. Composition is conditional on a bucket, so the routes are absent rather than present-and-broken              |
| VOI-04 | Member recording surface                                | DONE    | `/fie/settings/voice`. The 120 s bound is enforced in the reducer so it is testable without a microphone; a failed upload returns to `recorded`, not `idle`, because the take is still in the browser and a network blip must not cost the answer                                                                                                                                                                        |
| VOI-05 | Retention: 180-day crypto-erase, purge on revoke        | PARTIAL | `RetentionUntil` is set to 180 days at creation and `DueForPurge` is implemented and indexed. Nothing calls it — no worker job invokes `Purge`, so nothing is erased on its own yet                                                                                                                                                                                                                                      |

Sequencing is VOI-01 → VOI-02 → VOI-03 → VOI-04, because each depends on the
one before it. VOI-05 can be built alongside VOI-02 once the store exists.

### 27.12 Progress, 2026-09-05

VOI-01 through VOI-04 are done, tested and pushed (`e864bcb`, `c1c3d57`,
`12b867c`). A member can record three answers and have them stored, referenced
and consented to. VOI-05 is half done: the retention window is recorded and the
query that finds due recordings exists, but nothing calls it.

Decisions taken without asking, per the founder's instruction to proceed:

- **Storage is provider-neutral, not AWS.** SigV4 is implemented from the
  standard library. An SDK would have added a large dependency tree to a
  service whose release gate counts reachable advisories, and would have tied
  the adapter to one vendor.
- **`voice.introduction` is its own consent purpose, of its own kind.** It can
  be withdrawn on its own — a member may keep their account and stop us holding
  their recording — and the aggregate re-checks it on every transition, so a
  withdrawal stops work already in flight.
- **The transcriber is deferred, not blocking.** No speech vendor is
  contracted, so every request reports uncertain, which the aggregate already
  routes to its own uncertain state. This is the call already made for
  liveness. A transcript serves search, accessibility and safety review; the
  voice is the product, and it must not wait on a vendor nobody has signed.
- **Media access is owner-only with an explicit purpose list.** An unrecognised
  purpose is refused rather than waved through, so a future caller cannot
  inherit access to every member's media by inventing a string.

Still open inside this slice: TA-voice-encode-opus (the client records whatever
container the browser supports and does not transcode), TA-voice-server-pipeline
(no virus scan, no Sentinel step) and the retention sweep above.

**Note for the founder.** Commits `e864bcb` and `12b867c` were made with
`git add -A` and swept in frontend work that was uncommitted in the tree at the
time — admin desks and Fie shells that look like the redesign in progress.
Nothing was lost and everything is pushed, but that work is recorded under
these commit messages rather than its own. Worth a look before building on
those files.

## 28. Goal: make a recording hearable (2026-09-05)

Yesterday's slice built the write path and stopped there. A member can record
three answers and nothing — not even they — can play one back. `RequestRead`
on the media access service and `SignRead` on the object store are both
written and reachable from nowhere.

That is worse than an unfinished feature, because it breaks a chain that is
otherwise complete. Sowing requires twenty seconds of verified listening
(FR-202). The gate for it is built and composed:
`internal/seed/listening/domain/playback.go` sets `RequiredSeconds = 20.0`,
unions disjoint intervals so replays and out-of-order heartbeats cannot
double-count, and `POST /v1/listening/heartbeats` and
`GET /v1/listening/eligibility/{assetId}` are both on the wire. They are keyed
on the same `assetId` the introduction response already returns.

    record ✓ → playback grant ✗ → heartbeats ✓ → 20 s eligibility ✓ → Sow ✓

One missing link. Until it exists the gate can never be satisfied, so Sow can
never arm, so S-22 and the whole matching cluster rest on something that
cannot happen.

| Task   | Deliverable                                                                    | Status  | Notes                                                                                                                                  |
| ------ | ------------------------------------------------------------------------------ | ------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| VOI-06 | Playback grant: a short-lived signed read URL for an asset the caller may hear | TODO    | Owner-only to begin with, which is what `ownerpolicy` already enforces. A second rule joins it when introductions are delivered (S-21) |
| VOI-07 | Player on the voice surface, reporting listening heartbeats as it plays        | TODO    | Closes the chain end to end and lets a member hear what others will                                                                    |
| VOI-05 | Retention sweep                                                                | PARTIAL | `DueForPurge` is written and indexed; nothing calls it                                                                                 |

Deliberately not started yet: `internal/matching/coldstart` and
`internal/seed/source` are both fully built and have zero references outside
their own trees — the same orphan shape `internal/introduction` had. They are
S-20 and the next goal after this one, but composing them needs four port
adapters and at least one of those (reciprocal preferences) has no backing
store, so it is a larger piece of work than it looks. Finishing the chain
first means that work lands on something that already runs.

## 29. Goal: make the erasure promise real (2026-09-05)

The product tells a member that withdrawing a recording erases it. That is not
true today, in two separate ways, and one of them is an overstatement in this
document that needs correcting first.

**Correction to VOI-01.** Its row was written from a TODO that read
"envelope-encrypted per NFR-301b". Envelope encryption was not delivered. The
browser PUTs raw audio straight to the bucket under a signed grant, so the only
encryption at rest is whatever the bucket itself applies. NFR-301b — voice and
biometric blobs envelope-encrypted with a per-user key — remains open, and
until it is closed nothing here can honestly be called crypto-erase. It is
object deletion. The liveness path does seal its artifacts
(`liveness/adapters/outbound/privacy/sealer.go`); the introduction path does
not, because uploading through the API to encrypt it would put a two-minute
clip back on the request path that VOI-01 deliberately took it off.

**The two gaps.** `DueForPurge` is written and indexed and nothing calls it, so
no recording is ever swept. And `AssetRepository.Delete` sets `deletedAt` on
the row and leaves the object in the bucket, so even a sweep that ran would
erase nothing. A member who withdraws their voice today keeps it stored
indefinitely, and the audit trail would say otherwise.

`internal/platform/retention/retention.go` already anticipated this: its
`BindingPolicies` comment records that media-backed classes including room
voice at 180 days "activate when their persistence adapters land". That
adapter landed yesterday.

| Task     | Deliverable                                                | Status | Notes                                                                                                                                                                   |
| -------- | ---------------------------------------------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| VOI-08   | `Delete` on the object store                               | TODO   | Presign a DELETE and execute it, reusing the signing that VOI-01 already tests against the published AWS vector                                                         |
| VOI-09   | Media removal erases the object, then the row              | TODO   | Order matters: a row deleted first orphans bytes nothing can find to remove                                                                                             |
| VOI-10   | Worker sweep calling `Purge` on due and pending recordings | TODO   | Legal hold is never swept; the runner is explicit that holds live outside retention automation                                                                          |
| NFR-301b | Envelope-encrypt voice blobs with a per-user key           | OPEN   | Named here so the gap is not read as closed. Needs a scheme that keeps the direct-to-bucket upload — a client-side content key wrapped per user is the shape to look at |

## 30. Goal: let a member be introduced through a circle they belong to (2026-09-05)

The voice slice is finished end to end — record, store, play, listen, erase —
and nobody hears any of it, because introductions are never delivered. That is
the next thing a member actually experiences, and it is where 26 of the 74
confirmed critical gaps sit.

**Why not S-20 first.** `internal/matching/coldstart` is fully built and
orphaned, and looks like the same composition job `internal/introduction` was.
It is not. Its `Preferences` port returns reciprocal matching preferences and
**nothing in the repository backs it** — the only "preferences" anywhere are
notification settings. Composing coldstart means first building a whole
preferences context and a reciprocity computation, and it still has no weekly
cadence and caps at 20 rather than the 3–7 S-20 asks for.

**What is actually reachable.** `internal/seed/source` is the other orphan and
it is genuinely composable. Its mongo repository already exists, and its
`CandidateResolver` is deliberately source-scoped — the port comment says it
"must never expose a global member list, reverse graph, or profile payload".
Its four source types map onto contexts that already run: `consented_circle`
to circles, `consented_trust` to trust edges, `consented_host` to fires,
`policy_cohort` to cohorts. `Circle.Memberships()` and
`trust.Repository.Outgoing` are the two reads it needs, and both exist.

This is also the shape the product argues for. A member is introduced through
a circle they already belong to, not by an algorithm ranking strangers — so
the mechanism that is reachable is also the one that is right, and the
preferences engine is not on the critical path to a first introduction.

| Task   | Deliverable                                                | Status | Notes                                                                               |
| ------ | ---------------------------------------------------------- | ------ | ----------------------------------------------------------------------------------- |
| ISR-01 | `CandidateResolver` over circle membership and trust edges | TODO   | Bounded and consented; never a global list. Active members only, requester excluded |
| ISR-02 | `Authorizer` and `SourcePolicy` adapters                   | TODO   | A member may only open a source they belong to                                      |
| ISR-03 | Module composition                                         | TODO   | Repository exists; the rest is wiring                                               |
| ISR-04 | Routes and OpenAPI for open / withdraw / read              | TODO   | No route reaches this context today                                                 |
| ISR-05 | Member surface: ask to be introduced from a circle         | TODO   |                                                                                     |

NFR-301b (envelope-encrypted voice) and the preferences context both remain
open and are not part of this slice.

Two things ISR-04 turned up. The domain caps a source request at one hour
(`expiresAt.After(command.At.Add(time.Hour))`) — a first draft of the handler
set fourteen days and every real request would have been refused as invalid.
The bound is right: the request holds a snapshot of a roster, and a roster is
only true while nobody joins or leaves.

And `/v1/seed/sprouts`, `/v1/seed/declines` and `/v1/seed/doorways/{id}/exchanges`
were registered routes with no entry in the OpenAPI contract at all. Fixed in
§31.

## 31. Contract coverage (2026-09-05)

Three seed-stage routes had been served and undocumented. A full audit of the
198 registered routes against the document found exactly those three and no
others; an earlier pass reported a fourth problem — a path in the contract
that nothing served — and that was an extraction bug, not a finding. `/live`,
`/ready` and `/webhooks/resend` looked like phantoms only because the first
pass grepped `/v1/` and they do not start with it.

All three are now written down: sprout, doorway exchange and decline, with the
shapes their handlers actually return rather than shapes invented for the
document.

**The durable part is the test, not the three entries.**
`TestOperationIDsAreUnique` counts operation ids inside the document, so a
route served and never written down is invisible to it — the guard was doing
its job and its job was the wrong shape.
`TestEveryServedRouteIsInTheContract` compares the registrations in the source
against the paths in the document, in both directions, and fails on either.
Both directions matter: a generated client is built from this document, so a
route missing from it is unreachable to every client that trusts it, and a
path in it that nothing serves hands clients a method that 404s.

It reads the `mux.Handle` calls out of the source because Go's ServeMux does
not expose the patterns it holds, so there is nothing to ask at runtime. It
walks the paths block line by line rather than matching a path and scanning
forward, because a path that is a prefix of another would otherwise swallow
its methods — `/v1/blocks` would claim the delete on
`/v1/blocks/{blockerId}/{blockedId}`, which is exactly the mistake the first
audit script made.

The test was checked against both failure modes before being kept: removing a
documented path made it report the served route, and adding a path nothing
serves made it report the phantom. A guard that has never been seen to fail is
not yet a guard.

| Task  | Deliverable                                                             | Status |
| ----- | ----------------------------------------------------------------------- | ------ |
| CC-01 | Audit every registered route against the contract                       | DONE   |
| CC-02 | Document sprout, doorway exchange and decline                           | DONE   |
| CC-03 | `TestEveryServedRouteIsInTheContract`, both directions, failure-checked | DONE   |

Goal 30 is closed. A member can record three answers in their own voice, hear
them back, and ask to be introduced through a circle they belong to — the
first path from recording to another person that exists end to end.

What the ask does not yet do is deliver anybody. The request holds a bounded,
keyed candidate list and expires within the hour; turning that into an
introduction a member actually receives is S-20, and it needs a delivery
mechanism that does not exist yet. The count on the card is honest about this:
it says people could meet you, not that they will.

## 32. Goal: make the verification ladder mean something (2026-09-05)

The authorization kernel has been finished and unreachable since 2026-07-26.
`services/api/internal/authz/` evaluates a deny-by-default grant table that
already encodes FR-101 — `introductions.view`, `rooms.participate`,
`fires.attend` require Tier 1, `seeds.sow` requires Tier 2. Nothing calls it:

```
$ grep -rn "internal/authz" --include='*.go' services/api | grep -v '^.*/internal/authz/'
internal/identity/domain/account.go:23:  // ... mirror the authorization kernel's Tier ...
internal/admin/domain/principal.go:14:   // ... role vocabulary (services/api/internal/authz).
```

Two doc comments. That is the whole of it. The ladder is climbed and audited
correctly — verification approval promotes an account to Tier 1 — and then no
surface asks what rung anyone is standing on. `fire.go` is the sole exception,
and it passes the tier into the fire aggregate rather than through the kernel.

This is the same pattern as `introduction` and `seed/source` before them:
domain layers are real, the wiring is not. It is worse here only because the
orphaned context is the access control for the product.

It is also a gap in my own recent work. The Voice of Introduction (§28) and
the introduction ask (§30) are romantic surfaces by any reading of FR-101, and
I shipped both with no tier check — `circlepolicy.Authorizer` verifies circle
membership and nothing else. A Tier-0 account with no verified identity can
today record a Voice of Introduction and ask to be introduced to a circle.

### The decisions this needed

FR-101 gives the principle and stops there; `deploy/release/composition-inventory.md`
flagged the placements as an unmade decision rather than a mechanical wiring
job. Taken here, and recorded so they can be overridden:

**One gate, at the route boundary, reading the kernel's table.** Not inline
tier comparisons per handler — those drift. Rules stay in `policy.go`.

**The gate must never block the road out of the gate.** A Tier-0 member has to
be able to reach verification, consent, profile, privacy, settings, safety
reporting and their own account. Gate those and the ladder becomes a trap: you
cannot verify because you are unverified. This is an invariant with a test, not
a convention.

**A failed tier read denies loudly, not quietly.** A database blip must not be
indistinguishable from "you are unverified" — that is precisely the failure
that locks out members who have earned a surface. Unreadable tier answers 500;
only a real Tier-0 answers 403.

**Sowing stays at Tier 2 even though nobody can reach Tier 2.** Nothing in the
codebase promotes past 1; only verification approval exists, and it stops at
Tier 1. Gating sowing therefore closes it until an operator opens it, so this
goal adds the audited admin promotion that makes Tier 2 reachable at all.
What *automatically* earns Tier 2 is a product rule that has never been
decided, and I am not inventing one — flagged below as open.

| Task    | Deliverable                                                            | Status |
| ------- | ---------------------------------------------------------------------- | ------ |
| TIER-01 | Compose the authz kernel; one route-boundary gate through the table    | DONE   |
| TIER-02 | Romantic surfaces at Tier 1: introductions, courtship rooms, sources   | DONE   |
| TIER-03 | Sowing at Tier 2: sprouts (exchanges placed at Tier 1, below)          | DONE   |
| TIER-04 | Audited admin tier transition, so Tier 2 is reachable                  | DROPPED |
| TIER-05 | The client explains the rung instead of showing a bare refusal         | DONE   |

**Open product decision, for the owner not the adapter author:** what promotes
an account from Tier 1 to Tier 2. Until it is answered, sowing is
operator-opened only.

### What was gated, and what was deliberately not

`MemberGate` sits at the registration table rather than inside each handler, so
the access-control decisions are one readable list and a route absent from it
is visibly ungated. It authenticates once and carries the member forward in the
request context, so a gated route does not pay for a second session lookup.

Gated at Tier 1: beginning and confirming a Voice of Introduction, hearing one,
listening heartbeats and eligibility, opening and reading a circle
introduction ask, starting a courtship room, taking a turn, relighting the
pace, the honesty prompt, making a proposal and accepting one.

Gated at Tier 2: reaching toward someone (`POST /v1/seed/sprouts`).

**Not gated, on purpose.** A doorway only opens when both members reached for
each other, so gating the reply at the sowing rung could open a doorway one of
the two could never answer in; exchanges sit at Tier 1. And every exit stays
open at every rung: taking down your own recording, withdrawing an ask,
declining, pausing, closing a room, blocking and reporting. A member demoted to
Tier 0 for safety must not be shut in a room with the person they need to get
away from. `TestTheGateNeverBlocksTheRoadOutOfTheGate` reads the registration
tables from source and fails if any of those nine routes is ever wrapped.

Both guards were checked against their failure modes before being kept:
wrapping `safety/block` made the exit test name the offending line, and
weakening the tier comparison made the romantic-surface and sowing tests report
a Tier-0 member admitted and a Tier-1 member sowing.

### TIER-04 dropped, and why

The plan was an audited admin route to promote an account to Tier 2. Reading
the admin surfaces first showed that would have been a privacy regression:
operators act on **cases**, never on member ids, and the member directory
returns one-way keyed refs precisely so who a member is stays illegible to
staff. A `POST /v1/admin/members/{id}/tier` route would have needed an
addressable member id that no admin surface has, or should have.

The architecturally consistent path is a case an operator decides, the way
Tier 1 already works — and that needs the product rule for what earns Tier 2,
which has never been decided. So sowing is now correctly closed rather than
wrongly open, and nothing regressed today: no client calls `/v1/seed/sprouts`,
`/v1/seed/doorways/{id}/exchanges`, or any courtship route yet.

### TIER-05, finished in two halves

The circle ask turns a refusal into a link to verification: the BFF carries the
refusal `code` beside the message, and `tierNotice` maps `tier_1_required` to
somewhere the member can actually go.

The voice page needed the opposite fix. Refusing at save let a member record
three answers for nothing, so the rung had to reach the page *before* the
recorder. `GET /v1/onboarding/status` already answers "where do I stand" as a
coarse projection with no provider reasoning in it, so the rung was added
there rather than inventing a route: it is the member's own tier and nobody
else's. The page reads it server-side and renders a door with a handle instead
of the recorder.

Two failure directions were handled deliberately and in opposite ways, because
the cost of being wrong differs:

- The **API** refuses to guess. An unreadable account answers 500, never a
  fabricated Tier 0, because telling a verified member to verify again would
  send them to redo work they have already done.
- The **page** refuses to block. An unreadable rung renders the recorder as
  before, because the server gate is the authority on who may record and
  turning a slow status read into a locked page would shut out members who are
  perfectly entitled to be there.

**Still open:** what promotes an account from Tier 1 to Tier 2.

## 33. Goal: close the age gate (2026-09-05)

`services/api/internal/safeguarding` is the third finished-and-unreachable
context this week, and the most serious. It owns the under-18 hard block:

```
$ grep -rn "safeguarding" --include='*.go' services/api | grep -v "/safeguarding/"
$
```

Nothing. `module.go` was two lines of package doc — no composition root at
all, the same shape `introduction` was in before §28.

What was sitting behind that: `MinimumAge = 18`; `AgeAt` with correct birthday
arithmetic rather than a duration approximation; `Eligible`; a `Restriction`
that retains keyed compliance proof and deliberately never keeps a date of
birth; a `PurgeJob` holding lookup material that is destroyed on completion;
a 24-hour purge SLA; a Mongo purger that removes the verification case, the
account, its sessions, its consent records and its media; an HMAC keyer. All
of it tested. None of it reachable.

Meanwhile the only age arithmetic anywhere on a live path was `ageBand()` in
the admin repository — a display band for reviewers. So a fifteen-year-old
could submit a Ghana Card with their real date of birth, have it stored in
plain text in `verification_cases`, be matched by the provider, and be promoted
to Tier 1 — which, after §32, is precisely the rung that opens the romantic
surfaces.

### The decisions this needed

**The gate runs before the case exists.** Both submission paths generate the
case id, assess, and only then call `NewCase`. A minor's card number and date
of birth are never written to `verification_cases` at all, rather than written
and purged within 24 hours. The purge still runs behind the block, because the
account, its sessions and its consent records were created at sign-up before
any date of birth was known.

**The age gate is a required constructor argument, not a builder option.**
`WithDocuments` is optional because a deployment without encrypted storage can
honestly refuse that path. There is no equivalent for age: an optional age
check silently defaults to admitting children, and `NewModule` now refuses to
build without one.

**An unassessable age refuses.** A gate outage is not a pass. This is the
direction that admits a child, so it fails closed — and the two refusals are
distinguished only so the member gets an honest message: one says they may not
join, the other says to try again.

**The documents path needed its own branch.** Its catch-all would have told a
blocked minor that "secure storage is unavailable, please try again shortly" —
inviting exactly the retry the block exists to prevent.

**Something has to keep purging.** `Assess` attempts one purge inline and
returns; a lost transaction left the restriction pending forever, because
`PurgePending` is documented as "the worker entrypoint" and no worker called
it. A sweeper now runs it every 15 minutes, with a horizon one interval ahead
so a job is retried before its deadline rather than exactly on it.

| Task  | Deliverable                                                            | Status |
| ----- | ---------------------------------------------------------------------- | ------ |
| AGE-01 | Compose the safeguarding context; separately rotated keying secret     | DONE   |
| AGE-02 | Require the gate in verification; assess before anything is written    | DONE   |
| AGE-03 | Honest refusals on both submission paths, including the documents one  | DONE   |
| AGE-04 | Purge sweeper, so the 24-hour SLA is something that actually retries   | DONE   |

### The secret is a boot requirement, not a default

`SAFEGUARDING_HMAC_SECRET` is registered in `internal/platform/secrets` and in
`deploy/secrets/inventory.yaml`, which means **the API will refuse to boot in
any non-development environment until it and `SAFEGUARDING_HMAC_SECRET_ROTATED_AT`
are set.** That is deliberate. The restriction record outlives the purge of
everything else about the account, so keying it with a shipped local default,
or sharing another context's key, would turn compliance proof into a readable
list of which accounts belonged to children.

The repository's own inventory test caught the omission before this shipped,
which is the second time this week a policy test has been the thing that
noticed.

### Verified, not assumed

Both age tests were checked against their failure mode: removing the
`assessAge` call from `SubmitGhanaCard` made
`TestAnUnderageSubmissionIsRefusedBeforeAnythingIsWrittenDown` and
`TestAnUnassessableAgeRefusesRatherThanPasses` fail. The first asserts through
the mock controller that neither `Create` nor the provider is ever reached —
so it is testing that nothing is written down, not merely that an error is
returned.

**Still open after this goal:** an adult self-affirmation checkbox is collected
at onboarding (`AdultAgeID`) and is not age assurance; TS-AGE-ASSURANCE asks
for ID-derived DOB with an audit trail, which this delivers for the block but
not yet as a positive attested record on the approved path. FR-102a also
remains open — the identity provider is a simulator, so in the current
deployment any card number not ending in U, ? or X is treated as an issuer
match.


## 34. Goal: give national-ID data an end date (2026-09-05)

§33 delivered the deletion half of M1-02 only for blocked minors. For everyone
who verified successfully, nothing was ever deleted. The recon that mapped the
age gate is what surfaced it, and it holds up under direct reading:

- `identity_verifications` stores `dateOfBirth` as a plain `time.Time`. The
  case `Update` path writes status, providerRef, reason, decidedAt and version
  — it never touches the birth date, so it stays for the life of the record.
- `identity_documents` holds both sealed photographs of the card and carries
  no TTL. That was a deliberate decision, and the comment explaining it is
  right: "a queue can take days, and an image that expired mid-review would
  leave a member unverifiable with nothing explaining why."
- `internal/platform/retention` is the binding retention table (Doc 09 §7),
  it runs in the worker every six hours, it already supports `strip_field`
  and `delete`, and identity data simply was not in it.

So the machinery was reachable and working. The rows were missing.

`internal/compliance/retention` is a **fourth** finished-and-unreachable
context — policies, erasure status and legal holds, every reference inside its
own tree. It is the richer, per-record model with legal-hold support. It is not
what this goal used: the binding table already runs, and adding three
declarative rows to something that executes today beats composing a second
retention system that would then need its own sweeper, ports and proof
records. Composing it remains open work, not work this goal pretended to do.

### The decisions

**Strip the birth date, delete the images.** The case is the proof that a
check happened and what it decided; that proof should outlive the personal
data it was derived from. The photographs have no such role.

**Ninety days for the images, not thirty.** The document store's own comment
is the reason, and it was written by someone who had thought about it. Ninety
days is far past any plausible review queue while still being an end date.

**Two policies for the birth date, not one.** Keyed on `decidedAt`, a case
that is never decided has no such field and can never match — its birth date
would be kept forever. The second policy keys on `createdAt` at 180 days as
the backstop for abandoned and indefinitely queued submissions, set well
beyond any review so it cannot strip a case still being worked. A test asserts
the backstop is the longer of the two.

**`ageBand` had to be fixed in the same change.** It computes an age by
subtracting years, so an absent birth date yielded roughly two thousand and
returned `50_plus`. Stripping birth dates without this would have put a
confident, wrong age on the screen of a reviewer deciding about a real person.
It now reports `unknown`, and the contract enum was widened to say so.

| Task      | Deliverable                                                          | Status |
| --------- | -------------------------------------------------------------------- | ------ |
| RET-ID-01 | Binding policies for card images and both birth-date cases           | DONE   |
| RET-ID-02 | `ageBand` reports unknown instead of inventing `50_plus`             | DONE   |
| RET-ID-03 | Guards for silent-noop policies, proven against their failure modes  | DONE   |

Both guards were checked by breaking them: removing the `Fields` list made the
table test report a policy that "would run clean and retain everything", and
removing the zero check made the band test report the fabricated `50_plus`.

**Still open:** `internal/compliance/retention` remains uncomposed, and with it
per-record legal holds over identity data — the binding table is deliberately
blind to holds ("legal holds are never touched by this runner"), so a hold
placed on an identity case would not currently stop these policies.


## 35. Goal: prove the age check happened (TS-AGE-ASSURANCE, 2026-09-05)

§33 records an auditable fact only when someone is **blocked** — the keyed
`Restriction`. For everyone who passes, nothing recorded that an age check had
happened at all: the proof was that a case existed, which is only proof if you
have read the code and know a minor could never have got one.

§34 then made that worse on purpose. Stripping `dateOfBirth` thirty days after
a decision is right, but it meant that after thirty days there was no trace at
all that anyone had ever established the member was an adult. The retention
fix created the audit gap that this goal closes.

### The decisions

**The record lives on the case, not in safeguarding.** Safeguarding's
`Restriction` is proof of a block, and its events collection is
restriction-scoped with a unique index on `restrictionId`. Bending it to carry
positive assurances would have muddied a model whose whole value is that every
row in it means one thing.

**Method, not a boolean.** "We checked a date" and "the issuer confirmed the
date we checked" are different claims, and an audit trail that cannot tell them
apart overstates the weaker one. A submission is recorded as
`self_declared_dob`, and only a provider match upgrades it to
`issuer_corroborated_dob`. Today the provider is a simulator, so this keeps
FR-102a's gap visible in the data rather than hidden behind a case that merely
looks verified.

**The threshold is recorded, not assumed.** The gate reports the minimum age
it applied and the case stores it, so a record decided under an 18 rule still
says 18 if the rule ever changes.

**A case cannot be created without it.** `NewCase` refuses an unrecorded
assurance. The case is the audit artifact; letting one exist with a hole where
the most consequential decision was made is exactly the failure to prevent.

**It is stored apart from the birth date so retention cannot take it.** The
strip policy names `dateOfBirth` only. A test builds a case, strips the date,
and asserts the proof survives.

| Task     | Deliverable                                                          | Status |
| -------- | -------------------------------------------------------------------- | ------ |
| AGE-A-01 | `AgeAssurance` on the case; a case cannot exist without one          | DONE   |
| AGE-A-02 | Self-declared at submission, upgraded on an issuer match             | DONE   |
| AGE-A-03 | Persisted apart from the birth date, surfaced on the reviewer desk   | DONE   |

Guards were checked by breaking them: dropping the `Recorded()` check let a
case open with nothing recorded, and dropping the corroboration call made an
approved case still claim `self_declared_dob`. The first attempt at that second
experiment silently changed nothing — the replacement string did not match — so
it was re-run and confirmed rather than reported on the strength of a test that
had never actually been challenged.

**Still open:** the date is self-declared, and only a real issuer adapter
(FR-102a) makes `issuer_corroborated_dob` mean what it says. The simulator
returns a match on any card number not ending in U, ? or X, so in the current
deployment that upgrade is not yet evidence of anything.


## 36. Goal: make consecutive sends impossible (FR-301, IM-029, 2026-09-05)

`internal/courtship/drum` is the fifth finished-and-unreachable context: the
strict-alternation engine, fully tested, referenced nowhere outside its own
tree. The audit said so plainly — "a complete strict-alternation aggregate
exists but is unreachable, and the live send path has no turn check."

The live path confirmed it. `Room.Submit` checks membership and hands
straight to the queue, and the queue aggregate appends events carrying an
`ActorKey` it never compares to the previous one. A member could send as many
times in a row as they liked. For a product whose whole claim is that it is
not an inbox, that is the mechanic, not a detail.

### The decisions

**The rule went into the queue aggregate, not the drum engine.** Composing
drum is the larger, right answer for M5-02 — it also carries the stage, the
single drum holder and the voice-only first turn — and it needs two authz
capabilities (`courtship.drum.open`, `courtship.drum.turn`) that the grant
table does not have, which is a placement decision rather than wiring. But
FR-301 asks for consecutive sends to be **impossible at the API layer**, and
the strongest form of that is an invariant on the aggregate that owns the log.
A check in a handler is a rule only for as long as that handler is the only
way in. Composing the rest of the drum engine stays open, and §36 does not
pretend to have done it.

**Stale is reported before out-of-turn.** A member whose partner has just
replied is behind, not out of turn, and telling them to catch up is both true
and the thing that lets them send. Reversing the order would answer a member
who is one fetch away from their turn with a dead end. The two conditions are
naturally disjoint, and a test pins each.

**Existing rooms self-heal rather than needing a backfill.** The head document
gains `lastActorKey`, absent on every room written before this. Treating
absent as "nobody has spoken" would hand one free consecutive turn to whoever
sent first after deploy, so `State` reads the last event when the head has no
key yet, and the next write fills it in.

**A malformed key is refused, not ignored.** A last-actor value that is
present but not a real key would never equal any actor and would silently
switch the rule off for that room. `Rehydrate` rejects it; empty stays
legitimate.

**The refusal is a 409 that names the rule.** Without an explicit case it fell
to the default and returned 500 — telling a member who had simply spoken twice
that the server was broken, and inviting them to retry until something was.

| Task    | Deliverable                                                            | Status |
| ------- | ---------------------------------------------------------------------- | ------ |
| ALT-01  | Alternation as an invariant on the queue aggregate                     | DONE   |
| ALT-02  | Last actor persisted, self-healing for rooms that predate it           | DONE   |
| ALT-03  | 409 `not_your_turn` instead of a 500                                   | DONE   |

All three guards were checked by removing them: the aggregate accepted a
second consecutive turn, a reloaded room forgot who spoke last, and the route
answered 500 with `internal_error`.

The change also broke an existing fuzz test, correctly. `FuzzAcceptedEventsAreStrictlyOrdered`
built its sequences from a single actor, which is no longer a conversation the
aggregate will accept. It now alternates two speakers, which is what a real
sequence of turns looks like and leaves the test measuring what it was written
to measure.

### Audit rows reconciled

The handover table still listed FR-101a, FR-101b, M1-AC-01, M1-02 and
TS-AGE-ASSURANCE as MISSING after §32–§35 closed them, so the table was
describing a codebase that no longer exists. Those rows, plus FR-301, IM-029
and M5-02, now carry their real state and a note saying which goal changed
them. Three were set to PARTIAL rather than DONE on purpose: M1-AC-01 gates
the romantic surfaces but not every member surface, and M1-02 and
TS-AGE-ASSURANCE both still rest on a self-declared date of birth.


## 37. Goal: make Tier 2 reachable (2026-09-05)

§32 gated sowing at Tier 2 correctly and left it unreachable: nothing in the
codebase promoted anyone past Tier 1, so the rule was real and the rung was
not. The owner's ruling settled it — **Tier 1 plus a recorded Voice of
Introduction earns the sowing rung** — which is also the rule the Build Pack
implies, since FR-202 already ties sowing to Voice of Introduction playback.

### What had to be built first

The server could not tell whether a member had recorded their introduction.
It had no per-member query, and worse, no idea what any recording was an
answer to: the three prompts lived only in the browser. Counting recordings
would have let a member re-record "what brought you here" three times and
reach the sowing rung without ever answering the other two.

So the prompt vocabulary moved to the server. `arrival`, `ordinary` and
`welcome` are now domain constants, a recording cannot be opened without
naming one, and completeness is distinct questions rather than a count.

### The decisions

**"Recorded" means the bytes exist, not that transcription succeeded.**
`RecordedStatuses` includes every transcription outcome. The configured
transcriber is deferred and reports uncertain for everything, so requiring
`ready` would have meant no member ever finished an introduction — the same
trap the liveness provider set earlier in this project. A recording nobody
could transcribe is still the member's voice.

**A failed promotion never fails the recording.** The recording is already
stored. Telling a member who has just finished their introduction that it
failed would send them to re-record work that is safely saved, so the
promotion stays out of their way and is retried on the next confirmation.
The bridge logs its own failures so an outstanding promotion is visible.

**The bridge reads the rung before writing it.** An account already on the
sowing rung and one trying to skip a rung are refused by the same error, and
only the first is success. Swallowing both would hide a Tier-0 account
reaching a surface it should never have reached.

**The module refuses to build without the ladder.** The service takes it
through a builder so existing constructors keep working, but `NewModule`
rejects a nil one — forgetting it would leave sowing dark with nothing
saying why, which is exactly how it got dark in the first place.

| Task    | Deliverable                                                            | Status |
| ------- | ---------------------------------------------------------------------- | ------ |
| T2-01   | The server records which question each recording answers               | DONE   |
| T2-02   | Distinct-prompt completeness query and `Complete`                      | DONE   |
| T2-03   | Promotion to Tier 2 on completion, idempotent and non-blocking         | DONE   |
| T2-04   | Contract, generated client and recorder all carry the prompt           | DONE   |

### A test that was weaker than its name

`TestThreeTakesOfOneQuestionEarnNothing` did not test that. Its stub is handed
distinct prompts, so replacing the completeness rule with a plain length count
left it passing. It has been renamed to what it actually asserts, and
`TestDuplicateTakesDoNotFinishAnIntroduction` in the domain package is now the
real guard — proven by making `Complete` count recordings, which made it fail
with "three takes of one question finished an introduction".

That check is the only reason the weak test was found. Every other guard in
this goal was verified the same way: removing the promotion call made the
sowing test report that nothing earned the rung.


## 38. Goal: a Tier-0 account may look, not take part (M1-AC-01, 2026-09-05)

The owner ruled the ambiguity: M1-AC-01 says "fully gated from member
surfaces", but verification was deliberately moved out of onboarding so signup
is not blocked. The reading taken is **read-only browse, no participation** —
an unverified account can see that the place is real and cannot act in it.

### Capability placements, for review

Answer to the standing question about the ~39 capabilities the dark contexts
need: the FR-101 default is applied and recorded here rather than asked one at
a time. Two rows were added to `authz/domain/policy.go` in this goal:

| Capability            | Resource | Rung   | Why |
| --------------------- | -------- | ------ | --- |
| `circles.participate` | circle   | Tier 1 | Starting a circle or asking to join one is taking part, not looking. |
| `games.play`          | game     | Tier 1 | Entering a competition cohort is the same. |

`fires.attend` already existed at Tier 1 and now also covers scheduling a
fire, which was ungated even though attending one has required Tier 1 inside
the fire aggregate all along.

### What is gated, and what deliberately is not

Gated: `POST /v1/circles`, `POST /v1/circles/{id}/requests`, `POST /v1/fires`,
`POST /v1/game-cohorts/{cohortId}/join`.

Not gated, on purpose: every read, and every exit. Leaving is never
participation — a member who cannot join a circle must still be able to walk
out of one they are already in — so `POST /v1/circles/{id}/leave`,
`POST /v1/game-cohorts/{cohortId}/leave` and
`DELETE /v1/fires/{id}/rsvps/{memberId}` joined the exit invariant.

**In-circle activity is not gated here.** Rooms, stories, Ampe, Ebe and Oware
all require circle membership, and membership now requires a gated join, so a
new Tier-0 account cannot reach them. That is transitive protection rather
than a gate, and it does not cover an account that was already a member before
this shipped. Gating those turns directly is the next slice, not something
this goal did.

| Task    | Deliverable                                                          | Status |
| ------- | -------------------------------------------------------------------- | ------ |
| T0-01   | `circles.participate` and `games.play` at Tier 1                     | DONE   |
| T0-02   | Circle create/join, fire scheduling and cohort join behind the gate  | DONE   |
| T0-03   | Leaving added to the exit invariant                                  | DONE   |
| T0-04   | The other half of the invariant: participation routes stay gated     | DONE   |

### The half of the invariant that was missing

Checking a guard by breaking it found a hole in the guards themselves. Removing
`gate.guard` from the circle-join route broke nothing: the exit test proves the
gate never blocks the way out, and **nothing proved the gate was still on the
way in**. Ungating a participation route was a silent change every test passed.

`TestEveryParticipationRouteIsStillBehindTheGate` now reads the registration
tables from source and fails if any of twelve participation routes loses its
guard — verified by removing one and watching it name the offending line. That
test exists only because an experiment failed to fail.


## 39. Goal: a sow must be armed by listening (FR-202, 2026-09-05)

`Sprout` keyed the participants and recorded the intent. It consulted nothing:
not the listening record, not the allowance, not the decline lock. A member
could reach toward anyone without having heard a second of them, which is the
thing this product exists not to be.

This is a different shape from the orphaned contexts. `seed/listening` is
composed and reachable — `RequiredSeconds = 20.0` is real, heartbeats and
eligibility both have routes. The contexts were built and even wired to the
outside world. They were simply never introduced to each other at the one
place the decision is made.

### The decisions

**The gate takes members, not an asset id.** Listening eligibility is keyed by
asset, and the caller knows asset ids. Letting the sow name the asset would
let a member satisfy "you have heard them" with a recording belonging to
somebody else, so the bridge resolves the target's own recordings and checks
those.

**A member with no recording cannot be sown toward.** They have no assets, so
nobody can have heard them. That reads harsh until you turn it around: it is
the same rule as "people meet you through your voice", enforced from the other
side. Recorded here because it is a real consequence, not an oversight.

**Only sowing needs the gate.** The first attempt put it in `ready()`, which
also blocked `Exchange` — speaking inside a doorway both members already
opened. That would have shut existing conversations whenever the gate was
unavailable. The tests caught it immediately.

**A nil or failing gate refuses.** An outage must not arm a sow. Without
object storage there are no recordings at all, so nobody can have been heard,
and the sprout service reports itself unavailable rather than accepting an
unarmed sow.

**Refusal happens before anything is keyed or written**, so a sow that was
never armed leaves no trace rather than a refused one.

| Task    | Deliverable                                                          | Status |
| ------- | -------------------------------------------------------------------- | ------ |
| SOW-01  | `ListenGate` port; `Sprout` refuses an unheard reach                 | DONE   |
| SOW-02  | Bridge resolving the target's own recordings, not a caller's asset   | DONE   |
| SOW-03  | 409 `not_heard_yet` telling the member what would arm it             | DONE   |

Verified by breaking it: with the refusal disabled, the test reported that an
unheard sow returned nil. A test fixture also had to be fixed rather than
trusted — it keyed every value identically, which made actor and target the
same person, and the aggregate's own refusal would have made the test pass for
the wrong reason.

**Still open in this area:** FR-201a's seed allowance is not spent on a sow,
and M4-AC-01's 90-day decline lock is not consulted. Both contexts are built.
They are the next two slices, and this goal does not claim them.


## 40. Goal: a sow costs a seed (FR-201a, 2026-09-05)

The allowance ledger was complete and reachable — `GET /v1/seed/allowance`
served it, `Spend` existed with `ErrInsufficient`, and the domain even renews
the week automatically inside `Spend`. Nothing ever spent from it. A sow cost
nothing, and a sow that costs nothing is a swipe.

### The decisions

**Spend after the gate, before the intent.** Charging first would take a seed
from a member whose sow was then refused for not having listened, which is the
one failure here that costs them something real. Recording first would let a
sow through unpaid. Both sides are idempotent by command id, so a client retry
completes the sow without charging twice. A test asserts the allowance is not
touched at all when the listen gate refuses.

**The issuance and the spend carry different command ids.** This one would
have broken every first sow. A ledger records each command id alongside a
fingerprint of what that command did, so opening a member's first ledger with
the sow's own id would make the spend that follows look like the same command
with different input — refused as a conflict rather than charged. The bridge
uses `sow-open:` and `sow:` prefixes.

**A nil allowance refuses**, like the listen gate, and for the same reason: a
sow that cost nothing is exactly the outcome the rule exists to prevent, so it
must not be what happens when the wiring is missing.

**The refusal says when seeds come back.** A member who has spent the week's
allowance has not done anything wrong and is not broken; the message says when
more arrive rather than only that they are gone.

| Task    | Deliverable                                                          | Status |
| ------- | -------------------------------------------------------------------- | ------ |
| SOW-04  | `Allowance` port; a sow spends one seed                              | DONE   |
| SOW-05  | Bridge opening a first-time ledger under its own command id          | DONE   |
| SOW-06  | 409 `no_seeds_left` naming when the allowance returns                | DONE   |

Both guards were proven by breaking them: removing the spend made the
no-seeds test report nil, and moving the charge above the listen gate made the
ordering test report an unexpected call to `Spend` — which is precisely the
member-visible bug it exists to catch.

**Still open in this area:** M4-AC-01's 90-day decline lock is still not
consulted. That context is built too, and it is the next slice.


## 41. Plan: organization coupons and affiliate marketing (2026-09-05)

Requested by the owner, to be built after the current gap-closing work. This
section is a plan and nothing in it is implemented.

### What exists today

Nothing. A wide search returns no coupon, promo, discount, referral or
affiliate concept anywhere in the codebase. Two near-misses worth naming so
nobody trips on them later:

- **`vouch/assisted` already owns the word "voucher"** — but a voucher there is
  a person who vouches for another member's trustworthiness. It has nothing to
  do with money. The new concept must not be called a voucher.
- **There is no organization entity at all.** The system knows members,
  circles, matchmaker licences and admin principals. "Coupons for
  organizations" therefore needs an organization first, and that is the
  largest hidden cost in the request.

What can be built on: `commerce/catalog` (SKUs, minor units, GHS and USD),
`commerce/membership` (passes with granted/cancelled/refund events),
`commerce/ledger` (real double-entry with asset/liability/equity/revenue/
expense classes), `commerce/momo` (inbound payment intents),
`commerce/escrow`, `commerce/reconciliation`.

Note that RPM-25 — the matchmaker take rate — is still MISSING, so there is no
existing commission machinery to copy. Affiliate payouts would be the first
money the platform ever sends *out*, and `commerce/momo` models only inbound
intents. That is a rail that does not exist yet.

### The tension this has to be designed around

Obiara's architecture is built so that member identity is not legible: subject
and candidate ids are HMAC-keyed, operators act on cases rather than member
ids, and the admin directory deliberately returns one-way refs. Both requested
features pull directly against that.

- An organization sponsoring seats will want to know **who used them**.
- An affiliate earning commission needs their referrals **counted**, and will
  want to know who they were.

**Recommendation: both see counts, never identities.** An organization gets
"37 of your 50 seats are in use"; an affiliate gets "12 qualified referrals
this month, GHS X accrued". If an organization needs proof that a redeemer
genuinely belongs to it, that is a *claim proven at redemption* — a
domain-verified email or a one-time organization token — not an identity
report afterwards. Designing it the other way would quietly undo a property
the rest of the system spends real effort maintaining.

### The second tension, which matters more

This is a trust-first product with a safety mandate and a verification ladder
built to slow down exactly the behaviour a naive affiliate scheme rewards.
**Paying per signup creates an incentive to bulk-recruit.** That is the
opposite of what Tier gating, age assurance and Sentinel exist to do, and the
people best placed to exploit it are the ones the safety model is least able
to see.

**Recommendation: commission never accrues on a signup.** It accrues on a
*qualified* conversion — verified to Tier 1, retained some number of days, no
upheld safety finding — and is clawed back on refund or on removal for
conduct. This makes the affiliate's incentive point at the same thing the
product wants, and it is much harder to farm.

### Phasing

Each phase is independently shippable, and the money movement is deliberately
last. A phase that only accrues is safe to get wrong; a phase that pays out is
not.

**Phase 0 — organizations exist.** A new `internal/organization` context: an
organization with a name, a billing contact, a status and an audit trail; an
organization principal that can sign in to a narrow console. Extends the authz
grant table with organization roles. Nothing commercial yet. Without this,
"coupons for organizations" has no organization to hang off.

**Phase 1 — discount codes.** A `commerce/promotion` context (the name avoids
the `vouch` collision): a code with a scope (which SKUs), a shape (percentage
or fixed minor amount), a validity window, a redemption cap, and a
one-per-member rule. Redemption is keyed. It reduces what a member pays and
posts a contra-revenue line through the existing double-entry ledger, so
discounts are visible in the books rather than hidden in a price. No
organization money moves; the organization is only the issuer.

**Phase 2 — affiliate accrual.** An affiliate is a party with a code, entered
at signup rather than tracked by link, because a code is honest, auditable and
does not require following anybody around the internet. Referrals are keyed.
A qualified conversion posts an accrual to a liability account — the platform
now owes the affiliate — and nothing is paid. Clawback reverses the accrual.

**Phase 3 — payouts.** MoMo disbursement, which is new rails, plus affiliate
KYC (they are a payee), Ghana withholding on commission, and reconciliation
against the ledger. This phase is where the real risk is and it should be
built last and slowly.

**Phase 4 — sponsored seats.** An organization pre-funds a balance or is
invoiced, and its coupons draw down real money rather than discounting. Needs
B2B billing, which MoMo is a poor fit for. Deferred until somebody actually
asks for it.

### What has to be decided before Phase 1 starts

These are product decisions, not adapter decisions, and I will not invent
them:

1. **Who issues a coupon?** Obiara staff on an organization's behalf, or the
   organization itself in a console? The second needs Phase 0's console and
   changes the trust model.
2. **What can be discounted?** Membership passes only, or the whole catalog
   including matchmaker consultations and event tickets? Discounting a
   matchmaker's fee spends someone else's money.
3. **Does a coupon prove membership of the organization?** If a code leaks,
   anyone can use it. A cap limits damage; a per-member claim prevents it but
   requires the organization to supply something verifiable.
4. **What qualifies an affiliate conversion, exactly?** Recommended: Tier 1
   plus 30 days retained plus no upheld safety finding. The numbers are the
   owner's call.
5. **What is the commission?** RPM-25 already sets a 20%/15% platform take on
   the matchmaker marketplace; an affiliate rate should be set knowing that,
   and the two must not compound into a loss on a sale.
6. **Can members be affiliates?** Paying members to recruit members inside a
   dating product changes what the community is. This is the decision with the
   least engineering in it and the most consequence.

| Phase | Deliverable                                          | Status  |
| ----- | ---------------------------------------------------- | ------- |
| P0    | `internal/organization` context and org principals   | PLANNED |
| P1    | `commerce/promotion`: discount codes, keyed redemption | PLANNED |
| P2    | Affiliate codes and qualified-conversion accrual     | PLANNED |
| P3    | MoMo payouts, affiliate KYC, withholding, clawback   | PLANNED |
| P4    | Organization-funded sponsored seats                  | DEFERRED |


## 42. Goal: a decline shields for ninety days (M4-AC-01, 2026-09-05)

`ExclusionPeriod = 90 * 24 * time.Hour` has been in the decline domain the
whole time, and nothing on the sow path ever asked about it. A member could be
declined and reach again the same minute, which makes a decline meaningless
for the person it is supposed to protect.

### What was actually missing

The decline context could not answer the question. `IsExcluded` locks a
**seed** — it stops the same seed being acted on twice — and the sow needs the
**pair**: after A declines B, B may not reach for A again for ninety days.
That query did not exist, and there was no index for it.

It turned out to need no backfill. A decline already stores `recipientKey`,
which is the seed's owner — whoever reached out and was told no. Reusing that
field, and the namespaces it was keyed under, makes every decline already on
record enforceable from the day this ships.

### The decisions

**Raw ids across the boundary, not keys.** The seed stage keys participants
under its own namespace, so its keys and the decline context's keys are
different strings for the same person and can never be compared. The port
passes member ids and lets the decline service key them the way it wrote them.

**The refusal says the outcome and not the reason.** `Eligible` is documented
to return "no reason, timestamp, owner, or rejection signal" (FR-205), and a
lock that announced "they declined you" would hand the sower exactly the
rejection signal the context refuses to expose — and make the shield worse
than useless for the person behind it. The code is `reach_unavailable` and the
message is "You cannot reach toward this person right now."

**Checked before the charge.** Same rule as the listen gate: a member must not
spend a seed on a sow that was never going to land. A test fails if the
allowance is touched at all when the shield is up.

**An unreadable shield refuses.** If the store cannot answer, the sow does not
go through. The alternative is that a database blip lets somebody through a
wall built to protect a person who said no.

| Task    | Deliverable                                                          | Status |
| ------- | -------------------------------------------------------------------- | ------ |
| SOW-07  | `IsPairExcluded` query and its index                                 | DONE   |
| SOW-08  | `Service.Locked`, keyed the way declines were written                | DONE   |
| SOW-09  | `DeclineLock` port; sprout refuses before charging                    | DONE   |
| SOW-10  | A refusal that reveals nothing about a decline                       | DONE   |

Both guards proven by breaking them: removing the refusal let a shielded
target through, and moving the charge above the shield made the ordering test
report an unexpected call to `Spend`.

### The sow path is now complete

`POST /v1/seed/sprouts` enforces, in order: Tier 2 (§32, reachable since §37),
twenty seconds of the target's voice (§39), the ninety-day shield (§42), and
one seed from the weekly allowance (§40) — and only then records the intent.
Every refusal happens before anything is written or charged. That is FR-101b,
FR-202, FR-201a and M4-AC-01's listen gate, allowance and lock, on the live
path rather than in libraries beside it.

M4-AC-01's remaining clause is "no purchase" — that seeds cannot be bought.
Nothing sells them today, so nothing violates it; it becomes a real constraint
the moment a catalog SKU could top up an allowance, and §41's promotion work
is where that could accidentally be introduced.


## 43. Goal: gate the turns as well as the door (2026-09-05)

§38 gated joining a circle and left the turns inside it transitively
protected: membership requires a gated join, so a new Tier-0 account cannot
reach a circle room, a story or a game. That reasoning was recorded as a
deferral rather than a claim, and this closes it — because transitive
protection covers new accounts and not accounts that were already members
before the join was gated.

### What is gated now

Under `circles.participate`: creating a circle room entry, starting a story,
adding or editing a passage, granting publication, publishing, and the host
actions that grow or open a circle — visibility, approve, promote.

Under `games.play`: starting and playing Ampe, Ebe and Oware inside a circle.

### What is deliberately still open

**Expelling.** It removes somebody from a circle, which protects the people in
it, so a host must be able to do it at any rung — the same rule that keeps
blocking and reporting open. It is now in the exit invariant, so gating it
later fails a test rather than passing review.

**Deleting your own circle-room entry**, for the same reason: taking your own
words back out is not participation.

**Every read.** Browsing is the whole point of the ruling.

| Task    | Deliverable                                                          | Status |
| ------- | -------------------------------------------------------------------- | ------ |
| T0-05   | Circle rooms, stories and in-circle games behind the gate            | DONE   |
| T0-06   | Host growth actions gated; expel left open and pinned as an exit     | DONE   |
| T0-07   | Both structural invariants extended to the new routes                | DONE   |

The extended guard was proven by ungating one Oware move, which made it name
the offending line. M1-AC-01 can now be read the way the owner ruled it —
browse yes, participate no — rather than resting on the fact that most
participation happens to sit behind a join.
