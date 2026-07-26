# ADR-0001: Use one polyglot monorepo

- Status: Accepted
- Date: 2026-07-26
- Owners: Engineering
- Decision source: founder override in `agent_plan.md`

## Context

Obiara has three clients (member web, admin web, and mobile), Go API and worker
processes, shared design and localization assets, generated API clients, and
cross-cutting quality controls. Contract changes frequently span more than one
of these surfaces. Independent repositories would increase version skew and
make atomic contract changes harder during the initial team and product stage.

## Decision

Keep applications, Go services, shared packages, deployment configuration,
architecture records, and engineering automation in one Git repository.

- Use pnpm workspaces and Turborepo for TypeScript task orchestration.
- Use a Go workspace for Go services and internal Go tooling.
- Generate the TypeScript API client from the versioned OpenAPI contract.
- Expose root commands for formatting, linting, type checking, tests, contract
  drift checks, and builds.
- Enforce path ownership through the coordination ledger and later through
  repository ownership rules; repository proximity does not permit arbitrary
  cross-lane edits.
- Do not share domain implementations between clients and services. Share
  contracts, tokens, localization registries, and intentionally portable
  utilities only.

## Consequences

Cross-surface changes can be reviewed and verified atomically. One checkout and
one CI graph simplify onboarding. CI must use affected-project caching and path
filters to avoid running every expensive check for every change. Access control
for highly sensitive deployment operations must remain external to source-code
layout.

## Revisit triggers

Reconsider repository boundaries only if regulatory access separation, build
graph performance, independently governed release trains, or team ownership
creates measured friction that tooling and path controls cannot resolve.
