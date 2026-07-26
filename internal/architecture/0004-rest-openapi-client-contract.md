# ADR-0004: Use REST with an OpenAPI 3.1 client contract

- Status: Accepted
- Date: 2026-07-26
- Owners: API and client engineering
- Supersedes: GraphQL client gateway in source architecture

## Context

Member web, admin, and mobile need one stable, typed contract. Most operations
are commands, resource queries, provider webhooks, uploads, or explicit
state-machine transitions. GraphQL would add an execution and authorization
surface before the product has demonstrated a need for client-selected query
graphs.

## Decision

Expose versioned JSON REST APIs described by a repository-owned OpenAPI 3.1
document.

- Treat the OpenAPI document as the external client contract and validate it in
  CI.
- Generate and exactly pin a TypeScript client for web, admin, and mobile.
- Keep domain types distinct from transport schemas and map at inbound adapters.
- Use explicit command endpoints for state transitions rather than generic
  document mutation.
- Require idempotency keys on retriable writes and return a consistent error
  envelope with stable machine codes and correlation identifiers.
- Use cursor pagination and bounded field selection designed by the server.
- Use signed, replay-protected REST webhooks for partners.
- Use WebSocket or Server-Sent Events only for realtime state where polling
  fails a measured product or performance requirement; the REST contract
  remains authoritative for durable state.
- Contract evolution is additive within a supported version. Breaking changes
  require a new version, migration window, generated-client compatibility
  evidence, and an ADR or recorded API decision.

## Consequences

Clients gain compile-time types and deterministic cache and authorization
boundaries. Some screens may require purpose-built aggregation endpoints.
OpenAPI generation drift and breaking-change checks become required quality
gates.

## Revisit triggers

Consider GraphQL only if measured client use cases repeatedly require
independent graph-shaped composition that creates harmful endpoint
proliferation or over-fetching after aggregation endpoints are evaluated. A
proposal must include field-level authorization, query cost controls, persisted
operations, observability, caching, migration, and mobile bandwidth evidence.
