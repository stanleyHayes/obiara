# Obiara Admin Redesign Execution Ledger

Status: **COMPLETE**  
Plan version: 1.0  
Started: 2026-08-22  
Goal owner: `/root`  
Current slice: **All route batches and global verification — DONE**

This is the route-by-route coordination ledger for the authorized redesign of
the Obiara admin application. It is subordinate to `agent_plan.md`: product
laws, safety rules, access controls, data contracts, and release rules there
remain authoritative.

## Goal and completion condition

Redesign all 22 admin routes into one coherent, responsive operations product
while preserving their existing content, server-authoritative behavior, access
controls, and API contracts. Replace every textual or circular indeterminate
loading treatment with layout-faithful skeletons. Replace every true no-data
state with a borderless empty-state composition containing an animated icon, a
title, and a useful description. Reproduce the useful right-side account and
notification interaction pattern from `ashesischeduleflow`, adapted to Obiara's
roles, routes, tokens, and language.

The goal is complete only when every route below is `DONE`, shared primitives
are used consistently, and the repository-level verification and rendered QA
gates pass.

## Non-goals and immutable rules

- Do not invent, remove, rename, summarize away, or reorder away required
  business content. Visual hierarchy and placement may change; meaning may not.
- Do not change API shapes, authentication, authorization, MFA/step-up,
  redaction, audit, privacy, safety, or idempotency behavior as part of visual
  work.
- Do not fabricate metrics, records, alerts, identities, permissions, actions,
  or success states to make a page look populated.
- Do not make a destructive or privileged action easier to trigger. Existing
  confirmations, disabled states, actor authority, and server checks remain.
- Do not turn error states into empty states. Errors retain explicit recovery
  guidance and retry only where retry is already safe.
- Do not add a new design dependency unless the foundation owner records and
  verifies the reason. Prefer the current MUI, CSS, and icon stack.
- Do not use side-by-side list/detail compositions anywhere in the admin app.
  Lists occupy the full available content width. Complex records open on a
  dedicated route/page; only small, bounded details may open in an accessible
  dialog.
- Do not place asymmetric peer cards side by side. Cards with meaningfully
  different content density, hierarchy, or purpose stack vertically.
- All admin forms use one consistent structure for section hierarchy, labels,
  help text, field spacing, validation/error placement, busy/loading behavior,
  and action footers. Domain content may differ; interaction grammar may not.
- Every admin route uses the shared premium card system. Each card archetype has
  a meaningful page- and purpose-specific low-contrast watermark while retaining
  one coherent hierarchy, surface language, brand palette, dark mode, responsive
  behavior, reduced-motion behavior, and interactive-state grammar.
- `/login` and `/signed-out` retain the approved split-screen auth language;
  this program may harmonize tokens and shared states without regressing it.
- Content preservation is validated against the pre-redesign source, not by
  visual memory or screenshots alone.

## Status and claim rules

Allowed statuses are `BACKLOG`, `READY`, `IN PROGRESS`, `IN REVIEW`, `BLOCKED`,
and `DONE`. Only one owner may hold an `IN PROGRESS` row. Before editing, claim
the row and its paths. Shared paths require a reservation below or a recorded
handoff. A route becomes `DONE` only after its acceptance evidence is recorded.

## Shared-file reservations

| Reservation                              | Owner                             | Status      | Paths                                                                                                                                                                                                                                                   | Boundary / handoff rule                                                                                                                           |
| ---------------------------------------- | --------------------------------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| ADM-R00 Foundation, shell, shared states | `/root/foundation_implementation` | DONE        | `apps/admin/app/admin-rail.tsx`; `apps/admin/app/(ops)/layout.tsx`; `apps/admin/app/empty-state.tsx`; `apps/admin/app/styles.css`; shared admin presentation components and tests                                                                       | Completed and independently accepted 2026-08-22. Shared primitives are the required baseline for route work.                                      |
| ADM-R01/R04/R05/R06 Command and triage   | `/root/command_triage_redesign`   | DONE | `apps/admin/app/(ops)/page.tsx`; `apps/admin/app/(ops)/verification/**`; `apps/admin/app/(ops)/safety/**`; `apps/admin/app/(ops)/care/**`; route-specific extracted components/tests; `apps/admin/app/styles.css` only for coordinated shared additions | Accepted with dedicated details, vertical cards, truthful async states, segmented MFA, and exact route safety. |
| ADM-R02 Auth harmony                     | `/root/triage_routes_code`        | DONE | `apps/admin/app/auth-shell.tsx`; `apps/admin/app/login/**`; `apps/admin/app/signed-out/**`                                                                                                                                                              | Split screen, solid type, password visibility, support recovery, segmented OTP, exact auth bodies, and responsive QA accepted. |
| ADM-R03 Account-menu destinations        | `/root/triage_routes_code`        | DONE | Account/security/notification destinations or dialogs introduced under `apps/admin/app/`                                                                                                                                                                | Profile, security, appearance, replay-tour, sign-out, theme, and truthful notification states accepted. |

Reservation changes must be logged here before a second owner touches a shared
path. Page-specific desk files are exclusively owned by the matching route row.

## Foundation contracts

### Shell, sidebar, and top navigation

- The desktop sidebar is collapsible. Expanded mode shows icon and label;
  collapsed mode is a deliberate icon-only rail, not clipped expanded content.
- Every collapsed navigation item has an accessible name and a keyboard- and
  pointer-accessible tooltip. Tooltip text matches the visible expanded label
  and never becomes the only source of accessible naming.
- Active navigation state remains unmistakable in expanded and collapsed modes
  through more than color alone, and each active link exposes
  `aria-current="page"` where applicable.
- Sidebar scrolling is independent from main-content scrolling. Its brand,
  collapse control, and essential session/navigation affordances remain usable
  when either region contains long content.
- Desktop collapse preference persists per browser when storage is available,
  without hydration mismatch. Storage failure degrades safely to the expanded
  state and never blocks navigation.
- At mobile widths the persistent rail becomes a drawer opened from the top
  navbar. The drawer has a labelled trigger, focus containment, Escape and
  backdrop close behavior, scrollable navigation, focus return, and closes
  after route selection. Desktop persistence must not force the mobile drawer
  open or closed.
- The top navbar is split into two purposeful regions: the left side contains
  useful current-page context, the sidebar/drawer toggle, and page-level actions
  that genuinely apply; the right side contains action icons, notifications,
  theme control, and the operator/account cluster.
- Page context is route-aware and concise. Header actions remain permission
  aware, do not duplicate primary content actions without reason, and collapse
  intentionally on narrow screens instead of overflowing.
- The shell header remains available while the workspace scrolls; header,
  sidebar, drawer, popovers, dialogs, and toasts have an explicit, tested layer
  order.

### Page composition, records, and forms

- List pages are full-width working surfaces. Selecting a record never shrinks
  the list into a permanent side-by-side detail pane.
- Complex records with multiple sections, workflows, audit history, privileged
  actions, or substantial reading use a dedicated route/page with a clear back
  path and preserved list context where appropriate.
- Small and bounded details may use an accessible dialog only when the operator
  can understand and complete the task without page-level navigation. Dialogs
  have a programmatic name/description, focus containment and return, Escape
  behavior appropriate to mutation safety, and a mobile-safe layout.
- Asymmetric cards stack vertically. Equal-grid layouts are reserved for truly
  comparable peer summaries with equivalent hierarchy and density.
- Forms share a consistent page/dialog anatomy: purpose and context, logically
  grouped sections, persistent labels, optional help text, uniform field gaps,
  reserved local validation/error placement, and a predictable action footer.
- Validation appears adjacent to the responsible field without shifting the
  action area unexpectedly. Form-level/server errors appear in one consistent
  summary location and focus/announce appropriately.
- Busy forms preserve field and footer geometry, prevent duplicate submission,
  identify the action still in progress without a circular loader, and retain
  entered values and actionable server errors. Primary/cancel/destructive
  actions keep consistent order and responsive wrapping across the admin app.

### Premium cards and semantic watermarks

- Metric, queue, evidence, policy, form, warning, and timeline cards use shared
  premium foundations for radius, border, elevation, spacing, typography,
  selected/hover/focus/pressed states, and content layering. Variants express
  purpose without becoming unrelated mini-design systems.
- Every applicable card includes a meaningful watermark chosen for that page and
  card purpose. Metric watermarks reflect the measured subject; queue watermarks
  reflect workflow/triage; evidence watermarks reflect review or audit; policy
  watermarks reflect governance; form watermarks reflect the task; warning
  watermarks reflect risk; timeline watermarks reflect sequence/history.
- Watermarks are decorative only: `aria-hidden="true"`,
  `pointer-events: none`, excluded from keyboard order, and never relied on to
  convey state, identity, severity, label, or action meaning.
- Watermarks remain low contrast in light and dark mode, sit behind content, do
  not reduce text/control contrast, never collide with actions or status marks,
  and cannot be mistaken for a real control, badge, chart, or data value.
- Card hierarchy makes title, state, primary value/content, supporting context,
  and actions immediately distinguishable. Stronger surface treatment must not
  add decorative noise or hide dense operational information.
- Card layouts adapt without clipping or horizontal page overflow. Interactive
  cards expose visible keyboard focus and a non-color cue; non-interactive cards
  do not acquire misleading hover/cursor behavior.
- Any watermark or card transition is subtle and optional. Under
  `prefers-reduced-motion: reduce`, motion is removed while hierarchy, purpose,
  and interactive state remain clear.

### Loading

- No visible `Loading…`, `Loading`, spinner, circular progress indicator, or
  text placeholder remains in the admin dashboard or detailed routes.
- Skeleton geometry mirrors the final region: page header, metric cards,
  toolbar/filter controls, table header and rows, detail panels, charts, and
  action groups as applicable.
- Skeletons use a single accessible loading region (`aria-busy="true"` and a
  screen-reader-only status), never announce every skeleton cell, and do not
  imply real values.
- Skeleton transitions avoid layout shift and respect `prefers-reduced-motion`.
- Button-local mutation progress may use a stable disabled label/skeleton only
  when the action's meaning stays clear; never swap to an ambiguous spinner.

### Empty states

- Every true zero-result state uses the shared empty-state primitive with an
  animated, semantically relevant icon; title; explanatory description; and a
  distinctive cool-toned border treatment.
- Optional action is shown only when the current role has authority and the
  action is real. Filtered-zero states offer a safe clear-filter action.
- Animation is subtle, non-blocking, and decorative; reduced motion shows a
  stable icon with no loss of information.
- Empty state is distinct from first-load, permission denied, partial data,
  unavailable integration, and request failure.

### Errors and partial data

- Errors use explicit error treatment with what failed, operational impact, and
  a retry/recovery action when valid. Never show an animated empty state for an
  API failure.
- Pages with independently loaded regions preserve successful regions while a
  failed region reports its own error.
- Existing redaction and fail-closed language remains visible.

## Ashesi shell and right-side feature checklist

Reference implementation inspected in
`../ashesischeduleflow/apps/web/src/components/account-menu.tsx`,
`notification-bell.tsx`, and the admin shell. Reproduce the interaction model,
not Ashesi branding or route names.

- [ ] Desktop sidebar expands and collapses between labelled and intentional
      icon-only modes; the active route remains obvious in both.
- [ ] Icon-only navigation has correct accessible names and hover/focus
      tooltips; collapsed controls retain full keyboard and touch usability.
- [ ] Desktop collapse preference persists safely when appropriate, without a
      server/client hydration mismatch or dependency on local storage availability.
- [ ] Sidebar and main workspace scroll independently; long navigation and long
      page content do not move or trap one another.
- [ ] Mobile uses an accessible navigation drawer with labelled open/close
      controls, focus management, Escape/backdrop close, route-change close, and
      focus return.
- [ ] Top navbar left side presents the responsive sidebar/drawer toggle,
      meaningful current-page context, and only useful permission-backed page
      actions.
- [ ] Right-aligned header cluster containing theme control, notifications,
      other useful action icons, separator where appropriate, and account trigger.
- [ ] Account trigger shows avatar initials, operator name/email fallback, role,
      chevron, open state, and compact responsive behavior.
- [ ] Dropdown summary contains large avatar, operator identity, email, and a
      role/status pill.
- [ ] Dropdown items include **My profile** (account details), **Security**
      (password/account protection), **Notifications** (system/operations alerts),
      and **Replay tour** (walk through the workspace again), each with icon, title,
      and supporting description.
- [ ] **Sign out** is visually separated and describes that it ends the session
      on this device; it uses Obiara's existing sign-out behavior.
- [ ] Notification trigger exposes an unread badge and accessible unread count.
- [ ] Notification panel includes heading, unread summary, **Mark all read**,
      bounded recent-item list, per-item mark/read navigation, timestamps, and
      **View all notifications**.
- [ ] Notification zero state uses the new animated empty-state primitive, not
      plain “No notifications” text.
- [ ] Menus close on outside interaction, Escape, selection, route change, and
      sign-out; focus returns to the trigger and keyboard traversal is complete.
- [ ] Popovers remain within the viewport at mobile widths and have correct
      menu/dialog semantics, focus visibility, layering, and touch targets.
- [ ] Only backed capabilities appear. Missing notification/tour/account
      contracts are implemented truthfully or recorded `BLOCKED`; no decorative
      dead controls.

Notification inbox status (2026-08-22): **BLOCKED** by the absence of an admin
notification API/contract. Unread notification records, mark-all-read, item
timestamps, item navigation, and notification history must not be fabricated.
Until a server-authoritative contract exists, ADM-R00 exposes only a truthful
unavailable/zero-state treatment in the shell. This dependency does not change
the foundation from completing. ADM-R00 is `DONE`; the live inbox capability
remains separately `BLOCKED` until a server-authoritative contract exists.

## Execution batches and route inventory

All 22 discovered admin routes are represented exactly once.

| ID      | Batch | Route / surface                   | Status      | Owner                             | Exclusive paths                                                              | Required acceptance evidence                                                                                                                                                                                                                                                                  |
| ------- | ----: | --------------------------------- | ----------- | --------------------------------- | ---------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ADM-R00 |     0 | Shared shell and state foundation | DONE        | `/root/foundation_implementation` | Shared-file reservation above                                                | Strict independent acceptance PASS; anti-pattern PASS; review found no blocking/high/medium issues; admin 30/30 tests, typecheck, lint, 42-route production build, and diff-check PASS. Collapsible rail, top navbar, right action cluster, skeletons, and animated empty states implemented. |
| ADM-R01 |     1 | `/` operations dashboard          | DONE        | `/root/command_triage_redesign`   | `(ops)/page.tsx`; dashboard-only components/tests; coordinated shared styles | Content parity, skeleton-first loading, partial-error/animated-empty states, stacked command panels, responsive behavior, and stale-request protection independently accepted.                                                                                                                |
| ADM-R02 |     1 | `/login`                          | DONE | `/root/triage_routes_code`        | `login/**`; auth shared files by handoff                                     | Split screen, solid typography, password toggle, support recovery, segmented OTP, exact request/response validation, skeleton busy state, lifecycle, keyboard, and 320–1440 responsive proof accepted. |
| ADM-R03 |     1 | `/signed-out`                     | DONE | `/root/triage_routes_code`        | `signed-out/**`; auth shared files by handoff                                | Exact current-device truth, direct `/login` action, semantic control, and 320–1440 responsive proof accepted. |
| ADM-R04 |     2 | `/verification`                   | DONE        | `/root/command_triage_redesign`   | `(ops)/verification/**`; coordinated shared styles                           | Full-width queue plus dedicated detail route; MFA/redaction preserved; skeleton/zero/error/responsive and safe-return-context proof accepted.                                                                                                                                                 |
| ADM-R05 |     2 | `/safety`                         | DONE        | `/root/command_triage_redesign`   | `(ops)/safety/**`; coordinated shared styles                                 | Full-width queue plus dedicated detail route; assignment/evidence safeguards, async generations, skeleton/zero/error states accepted.                                                                                                                                                         |
| ADM-R06 |     2 | `/care`                           | DONE        | `/root/command_triage_redesign`   | `(ops)/care/**`; coordinated shared styles                                   | Full-width queue plus dedicated detail route; safety language, async generations, skeleton/zero/error/responsive states accepted.                                                                                                                                                             |
| ADM-R07 |     2 | `/incidents`                      | DONE        | `/root`                           | `(ops)/incidents/**`                                                         | Runbook content and authority caveats preserved; semantic watermarked cards, compact priority treatment, and full vertical responsive stack independently accepted.                                                                                                                           |
| ADM-R08 |     3 | `/members`                        | DONE        | `/root/command_triage_redesign`   | `(ops)/members/**`                                                           | Full-width redacted roster, truthful skeleton/error/empty states, bounded accessible detail dialog, cancellation/generation safety, and privacy boundaries independently accepted.                                                                                                            |
| ADM-R09 |     3 | `/operators`                      | DONE        | `/root/command_triage_redesign`   | `(ops)/operators/**`                                                         | Dedicated exact keyed detail route, truthful independent queues, immutable confirmation terms, segmented MFA, four-eyes and server-authority protections independently accepted.                                                                                                              |
| ADM-R10 |     3 | `/workforce`                      | DONE        | `/root/command_triage_redesign`   | `(ops)/workforce/**`                                                         | Six safeguards and authority boundaries preserved in a full-width semantic vertical card stack; responsive and accessibility evidence accepted.                                                                                                                                               |
| ADM-R11 |     3 | `/account`                        | DONE        | `/root/command_triage_redesign`   | `(ops)/account/**`                                                           | Account/security/appearance destinations, shaped states, truthful local preferences, safe sign-out, mobile tabs, forms, and accessibility independently accepted.                                                                                                                             |
| ADM-R12 |     4 | `/community`                      | DONE        | `/root/command_triage_redesign`   | `(ops)/community/**`                                                         | Static fail-closed evidence chain, review boundary, and real links preserved in a full-width semantic vertical stack.                                                                                                                                                                         |
| ADM-R13 |     4 | `/mpanyimfo`                      | DONE        | `/root/command_triage_redesign`   | `(ops)/mpanyimfo/**`                                                         | Outcomes, policy, five dimensions, six gaps, and authority boundaries preserved without invented runtime; final CSS cascade verified vertical.                                                                                                                                                |
| ADM-R14 |     4 | `/matchmakers`                    | DONE        | `/root/triage_routes_code`        | `(ops)/matchmakers/**`                                                       | Full-width register plus dedicated create/exact-renew/escrow routes; contract validation, immutable confirmation/MFA replay, stable idempotency, truthful states, and exact response identity independently accepted.                                                                         |
| ADM-R15 |     4 | `/waitlist`                       | DONE        | `/root/command_triage_redesign`   | `(ops)/waitlist/**`                                                          | GET-only roster, consent copy, truthful skeleton/error/empty states, safe refresh cancellation, and mobile email wrapping independently accepted.                                                                                                                                             |
| ADM-R16 |     5 | `/finance`                        | DONE        | `/root/triage_routes_code`        | `(ops)/finance/**`                                                           | Bounded reconciliation validation, independently available guarded settlement, exact escrow/statement truth, vertical semantic evidence cards, and truthful skeleton/error/empty states independently accepted.                                                                               |
| ADM-R17 |     5 | `/analytics`                      | DONE | `/root/triage_routes_code`        | `(ops)/analytics/**`                                                         | Exact bounded 30-day evidence validation, freshness, skeleton/error states, accessible thresholds/counts, vertical cards, and permanently fail-closed release truth accepted. |
| ADM-R18 |     5 | `/governance`                     | DONE        | `/root/triage_routes_code`        | `(ops)/governance/**`                                                        | Validated vertical register, truthful states, immutable draft/publish/retire confirmations, four-eyes and exact MFA replay independently accepted.                                                                                                                                            |
| ADM-R19 |     5 | `/controls`                       | DONE        | `/root/command_triage_redesign`   | `(ops)/controls/**`                                                          | Exact retained proposal terms, stable idempotency, confirmation/MFA flow, truthful load states, vertical dialogs/list, and responsive behavior independently accepted.                                                                                                                        |
| ADM-R20 |     6 | `/game-content`                   | DONE        | `/root/triage_routes_code`        | `(ops)/game-content/**`                                                      | Contract-bound vertical approval form, private answer-count confirmation, exact public response proof, field validation/focus, and guarded mutation independently accepted.                                                                                                                   |
| ADM-R21 |     6 | `/tournaments`                    | DONE        | `/root/triage_routes_code`        | `(ops)/tournaments/**`                                                       | Landing/cohort/competition routes, full privacy-safe projections, stable revision commands, review-only MFA, truthful states, and semantic vertical layout independently accepted.                                                                                                            |
| ADM-R22 |     6 | `/launch`                         | DONE | `/root/triage_routes_code`        | `(ops)/launch/**`                                                            | Exact static 4-item repository handoff and 13 external gates preserved as a vertical, semantic, production-blocked inventory without invented live authority. |

Batch dependency order is 0 → 1 → 2 → 3 → 4 → 5 → 6. A later batch may be
promoted to `READY` when the foundation interfaces are stable and its paths do
not overlap an active slice. Within a batch, disjoint route rows may run in
parallel after they are claimed.

### Batch layout and form acceptance overlay

These requirements apply in addition to each route row's evidence column.

| Batch / routes                                                                             | Required layout and form evidence                                                                                                                                                                                                                                                                           |
| ------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Active command and triage: `/`, `/verification`, `/safety`, `/care`                        | Full-width queues/lists; no split list/detail panes; complex cases/reviews use dedicated detail routes and bounded previews/actions use accessible dialogs; unequal command cards stack; all filters, decisions, assignments, and care forms follow the shared form anatomy and busy/error/footer contract. |
| Auth: `/login`, `/signed-out`                                                              | Approved split-screen outer auth presentation may remain, but the form itself follows the shared field spacing, validation/error, busy, and action-footer contract; this exception is not permission for an in-app list/detail split.                                                                       |
| Operations and incidents: `/incidents`, `/members`, `/operators`, `/workforce`, `/account` | Full-width rosters, queues, and histories; complex incidents, members, principals, and audits use dedicated routes; bounded edits/confirmations use accessible dialogs; asymmetric summaries stack; every account/role/safeguard form uses the shared contract.                                             |
| Community: `/community`, `/mpanyimfo`, `/matchmakers`, `/waitlist`                         | Full-width list and docket surfaces; complex community/licence/candidate records use dedicated routes; bounded approvals/edits use dialogs; asymmetric cards stack; all workflow forms use consistent validation, busy behavior, and footers.                                                               |
| Governance and evidence: `/finance`, `/analytics`, `/governance`, `/controls`              | Full-width ledgers, evidence tables, and control lists; complex financial/governance records use dedicated pages; bounded confirmations use accessible dialogs; non-peer evidence cards stack; all privileged forms retain consistent structure and action placement.                                       |
| Content and launch: `/game-content`, `/tournaments`, `/launch`                             | Full-width catalogs/readiness lists; complex content, cohort, tournament, and readiness records use dedicated pages; bounded edits use accessible dialogs; asymmetric cards stack; configuration and action forms follow the shared contract.                                                               |

Every batch must also inventory its metric, queue, evidence, policy, form,
warning, and timeline card archetypes; apply the shared premium foundation;
choose a meaningful page-specific watermark for each applicable purpose; and
prove light/dark, responsive, interaction, accessibility, and reduced-motion
behavior.

## Per-route redesign checklist

Each route owner records evidence for every applicable item before `IN REVIEW`:

- [ ] Pre-change content/action/API inventory captured; post-change parity
      checked line by line.
- [ ] Page title, purpose, priority action, context, filters, status, and detail
      hierarchy are clear without changing meaning.
- [ ] Every initial, refetch, and region-level load uses a geometry-faithful
      skeleton; no text loader or circular loader remains.
- [ ] True empty and filtered-empty states use animated icon + title +
      description + cool border; authorized actions only.
- [ ] Error, permission, unavailable-provider, partial-data, and empty states are
      visually and semantically distinct.
- [ ] Tables, cards, charts, forms, dialogs, menus, and destructive actions retain
      their content, validation, confirmations, and server authority.
- [ ] Lists remain full-width with no side-by-side detail pane; complex records
      have dedicated routes/pages and only small bounded details use accessible
      dialogs.
- [ ] Asymmetric cards stack vertically; grids contain only comparable peers.
- [ ] Every form matches the shared structure, spacing, validation/error,
      skeleton/busy, duplicate-submit prevention, and action-footer contract.
- [ ] Every card uses the shared premium hierarchy and surface treatment with a
      meaningful page/purpose-specific watermark for its metric, queue, evidence,
      policy, form, warning, or timeline role.
- [ ] Each watermark is low contrast, behind content, `aria-hidden`, ignores
      pointer input, cannot obscure or read as content/control, and remains safe in
      light/dark, mobile, interactive, and reduced-motion states.
- [ ] Keyboard order, visible focus, labels, landmarks, live-region behavior,
      contrast, and 44px touch targets pass.
- [ ] Layout is verified at 390px, 768px, 1024px, and wide desktop with no
      horizontal page overflow; dense tables have an intentional narrow strategy.
- [ ] Motion communicates state only, does not block interaction, and becomes
      static under `prefers-reduced-motion: reduce`.
- [ ] Focused tests, admin typecheck, lint, relevant production build, and
      `git diff --check` pass; rendered screenshots or browser notes are linked in
      the row/evidence log.

## Program acceptance gates

1. Source scan finds no user-visible loading prose or circular progress in
   `apps/admin/app`, excluding screen-reader-only loading announcements.
2. Every asynchronous data region has tested loading, success, empty, error, and
   partial states where its contract supports partial data.
3. All true empty states meet the shared animated icon/title/description/cool
   border contract and reduced-motion fallback.
4. Ashesi-derived right-side dropdown and notification capabilities are present,
   accessible, responsive, and wired to real Obiara behavior.
5. The shell proves accessible expanded/collapsed desktop navigation, tooltips,
   active state, independent scrolling, safe preference persistence, responsive
   mobile-drawer behavior, and the required left-context/right-actions navbar
   composition.
6. Every route proves full-width lists with no side-by-side list/detail layout,
   dedicated pages for complex records, accessible dialogs for bounded detail,
   vertically stacked asymmetric cards, and consistent forms.
7. Every applicable card archetype uses the shared premium card system and a
   meaningful page-specific decorative watermark; rendered QA proves hierarchy,
   brand palette, dark mode, responsive/interactive states, reduced motion, and
   that watermarks neither obscure nor read as content or controls.
8. All route content and permitted actions match the pre-redesign inventory;
   server/API/security behavior is unchanged.
9. Admin tests, typecheck, lint, production build, and repository-required checks
   pass. No unrelated worktree changes are absorbed.
10. Authenticated rendered QA covers all 22 routes at desktop and mobile widths,
    including keyboard-only operation, empty/error fixtures, no console errors,
    and no unintended horizontal overflow.

## Acceptance evidence log

| Date       | Slice / route                                     | Owner                             | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                       | Result |
| ---------- | ------------------------------------------------- | --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| 2026-08-22 | Inventory and execution ledger                    | `/root/admin_redesign_ledger`     | 22 route entries reconciled against `apps/admin/app/**/page.tsx`; Ashesi account/notification source inspected; parent link added without changing existing task states                                                                                                                                                                                                                                                                                        | DONE   |
| 2026-08-22 | ADM-R00 foundation                                | `/root/foundation_implementation` | Strict independent acceptance and anti-pattern checks PASS; code review found no blocking/high/medium issues; admin 30/30 tests, typecheck, lint, 42-route production build, and diff-check PASS. Implemented collapsible rail, contextual top navbar, right action cluster, shared skeleton loading, and animated empty states. Live notification inbox remains BLOCKED by missing admin API/contract and retains a truthful unavailable/zero state. | DONE   |
| 2026-08-22 | ADM-R01/R04/R05/R06 command and triage            | `/root/command_triage_redesign`   | Independent acceptance, code review, and anti-pattern audit PASS after remediation. Lists and details are separated; unequal panels stack; skeleton/animated-empty states, segmented MFA OTP, trusted return notices, Strict Mode-safe generations, and stale-request guards verified. Admin 59/59 tests, UI 14/14 tests, typecheck, lint, production build, and diff-check PASS.                                                                              | DONE   |
| 2026-08-22 | ADM-R08/R10/R12/R13/R15 people and community      | `/root/command_triage_redesign`   | Independent acceptance, code review, and anti-pattern audit PASS after truthfulness, cancellation, semantic markup, final-cascade stacking, and dialog-state remediations. Admin 68/68 tests, typecheck, lint, production build, and diff-check PASS.                                                                                                                                                                                                          | DONE   |
| 2026-08-22 | ADM-R09/R11/R19 privileged administration         | `/root/command_triage_redesign`   | Independent acceptance, code review, and anti-pattern audit PASS after truth-state, exact-confirmation, MFA-generation, idempotency, language, and mobile remediations. Admin 71/71 tests, UI 14/14 tests, typecheck, lint, production build, and diff-check PASS.                                                                                                                                                                                             | DONE   |
| 2026-08-22 | ADM-R14/R16 commercial operations                 | `/root/triage_routes_code`        | Independent acceptance, code/security review, and anti-pattern audit PASS after typed state-machine, endpoint identity, wire-format instant, validation, truth-state, and finance lifecycle remediations. Admin 80/80 tests, UI 14/14 tests, typecheck, lint, 44-page production build, and diff-check PASS.                                                                                                                                                   | DONE   |
| 2026-08-22 | ADM-R07 incidents and shared empty-state revision | `/root`                           | Borderless animated EmptyState and compact semantic incident-priority card accepted; every incident card group stacks vertically and wrapper-aware packet layout is regression-tested. Admin 72/72 tests, typecheck, lint, production build, and diff-check PASS.                                                                                                                                                                                              | DONE   |
| 2026-08-22 | ADM-R18/R20/R21 content and programmes            | `/root/triage_routes_code`        | Independent acceptance, code/security review, and anti-pattern audit PASS after transition-state, route-key, modal-error, prompt-boundary, and stale-load remediations. Admin 87/87 tests, UI 14/14 tests, typecheck, lint, 44-page production build, and diff-check PASS.                                                                                                                                                                                     | DONE   |
| 2026-08-22 | ADM-R02/R03/R17/R22 auth, evidence and launch      | `/root/triage_routes_code`        | Independent acceptance, code/security review, and full-admin anti-pattern audit PASS after exact auth-body, password-minimization, analytics-freshness, triage-error, metric-stack, watermark, and live-region remediations. Admin 94/94 tests, UI 14/14 tests, typecheck, lint, 44-page production build, and diff-check PASS. Browser QA confirmed solid split-screen auth at 1440px and stacked layouts at 768/390/320px with no horizontal overflow, 48px password toggle, direct signed-out action, and no application console errors. | DONE   |

## Decision log

| Date       | Decision                                                                           | Reason / constraint                                                                                                                                                                                                               | Owner                         |
| ---------- | ---------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- |
| 2026-08-22 | Treat this as a full admin program, not a one-page polish task                     | User explicitly authorized page-by-page redesign and rearrangement while preserving content                                                                                                                                       | `/root`                       |
| 2026-08-22 | Shared loading and empty-state primitives land before route batches                | Prevent one-off spinners, loaders, borders, and motion patterns from diverging                                                                                                                                                    | `/root/admin_redesign_ledger` |
| 2026-08-22 | Adapt Ashesi interactions rather than copy its visuals/routes                      | Obiara must retain its brand, permissions, and truthful capabilities                                                                                                                                                              | `/root/admin_redesign_ledger` |
| 2026-08-22 | Keep detailed routes unclaimed until foundation stabilizes                         | Prevent shared CSS/shell collisions and dishonest parallel status                                                                                                                                                                 | `/root/admin_redesign_ledger` |
| 2026-08-22 | Make the collapsible responsive shell and two-sided top navbar part of ADM-R00     | User explicitly requires the sidebar, drawer, page context/actions, and right-side control cluster before page-by-page redesign proceeds                                                                                          | `/root/admin_redesign_ledger` |
| 2026-08-22 | Block live notification inbox features until an admin notification contract exists | Unread records, mark-all-read, timestamps, navigation, and history require server authority; the shell must show a truthful unavailable/zero state instead of invented data                                                       | `/root/admin_redesign_ledger` |
| 2026-08-22 | Accept ADM-R00 and promote command-and-triage routes                               | Independent acceptance, anti-pattern review, code review, tests, typecheck, lint, 42-route build, and diff-check passed; dashboard and highest-priority triage desks can now consume the stable foundation                        | `/root/admin_redesign_ledger` |
| 2026-08-22 | Ban side-by-side list/detail layouts and standardize admin forms                   | User explicitly requires full-width lists, dedicated pages for complex records, accessible dialogs for bounded detail, vertical asymmetric cards, and one consistent form grammar across all active and remaining batches         | `/root/admin_redesign_ledger` |
| 2026-08-22 | Require premium cards with semantic decorative watermarks on every route           | User explicitly requires shared high-quality hierarchy and surfaces with page/purpose-specific low-contrast watermarks that remain brand-consistent, accessible, dark-mode safe, responsive, interactive, and reduced-motion safe | `/root/admin_redesign_ledger` |

## Review and handoff log

| Date       | Slice                                                  | Reviewer / handoff                                      | Findings, corrections, next owner                                                                                                                                                                                                                      |
| ---------- | ------------------------------------------------------ | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 2026-08-22 | ADM-R00 foundation                                     | Independent acceptance                                  | PASS; anti-pattern PASS; no blocking/high/medium review findings. Foundation accepted with live notification inbox capability still contract-blocked.                                                                                                  |
| 2026-08-22 | ADM-R01/R04/R05/R06 command and triage                 | Independent acceptance, code review, anti-pattern audit | PASS after sanitizer, trusted-notice, lifecycle-generation, vertical-stack, and dashboard abort-guard remediations. Admin 59/59 tests, UI 14/14 tests, typecheck, lint, production build, and diff-check passed.                                       |
| 2026-08-22 | Shared semantic card foundation on ADM-R01/R04/R05/R06 | Independent acceptance, code review, anti-pattern audit | PASS after state-aware watermark suppression, wrapper-selector, keyboard-focus, and filtered-zero remediations. Admin 63/63 tests, typecheck, lint, production build, and diff-check passed.                                                           |
| 2026-08-22 | ADM-R08/R10/R12/R13/R15 people and community           | Independent acceptance, code review, anti-pattern audit | PASS after error-vs-empty truthfulness, abort/generation, semantic skeleton markup, vertical CSS cascade, refresh, dialog-state, and accessible-description remediations. Admin 68/68 tests, typecheck, lint, production build, and diff-check passed. |
| 2026-08-22 | Complete admin programme and global audit               | Independent acceptance, code/security review, anti-pattern audit, rendered browser QA | PASS. All 22 route rows and foundation are DONE. Global scan found no circular/raw/text loaders, raw route Cards, continuous OTP, nested Link/Button, active side-by-side card grids, or list/detail co-rendering. Live notification history remains the only capability-level BLOCKED item because no authoritative API exists. |

## Current next actions

1. All route, shell, state, card, form, auth, and responsive redesign rows are
   `DONE` and independently accepted.
2. Live notification history/mark-all-read remains `BLOCKED` until an
   authoritative admin notification contract exists; the current UI is
   intentionally truthful and unavailable rather than simulated.
