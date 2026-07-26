# ADR-0002: Build a Go hexagonal modular monolith

- Status: Accepted
- Date: 2026-07-26
- Owners: Backend engineering
- Supersedes: Kotlin/Spring or NestJS choice in source architecture

## Context

The product contains strong state-machine and ledger invariants, numerous
provider integrations, and asynchronous transitions. The initial organization
is better served by transactional cohesion and simple operations than by
premature distributed-service boundaries. Provider choices will change, so
domain rules must not depend on SDKs or infrastructure representations.

## Decision

Implement a stateless Go HTTP API as a modular monolith and a separate Go worker
process that consumes durable jobs and performs scheduled transitions. Organize
each bounded context as:

```text
internal/<context>/
├── domain/
├── application/
├── adapters/
│   ├── inbound/
│   └── outbound/
└── module.go
```

- Domain code imports only the standard library and domain-owned code.
- Application code orchestrates domain behavior through explicit ports.
- Inbound and outbound adapters depend inward and translate external types.
- Composition roots wire concrete adapters at process startup.
- Provider SDK and transport types never cross into domain models.
- Commands carry actor context and idempotency keys; state-changing aggregates
  use explicit versions or equivalent concurrency guards.
- Domain changes and outbox records commit atomically. Consumers use durable
  inbox/deduplication records and are safe under at-least-once delivery.
- GoMock is limited to application-port unit tests. Domain logic uses direct
  tests, while persistence and provider boundaries use integration or contract
  tests, including Testcontainers for MongoDB.
- API and worker may share modules, but each has a separate composition root and
  least-privilege runtime credentials.

## Consequences

Product invariants remain testable without infrastructure, and provider
replacement is localized to adapters. The module boundary discipline must be
enforced through dependency tests and review because Go packages alone do not
guarantee architectural direction. Long-running or load-distinct workloads
remain outside request handlers.

## Revisit triggers

Extract a module only when production evidence demonstrates a distinct scaling
profile, security boundary, failure-isolation requirement, data-governance
boundary, or independently owned release cadence. Extraction must preserve its
ports and contract tests and receive a new ADR.
