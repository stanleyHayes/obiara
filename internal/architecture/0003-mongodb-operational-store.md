# ADR-0003: Use MongoDB as the initial operational store

- Status: Accepted
- Date: 2026-07-26
- Owners: Backend and platform engineering
- Supersedes: PostgreSQL and Neo4j defaults in source architecture

## Context

Obiara requires evolving profile documents, strict state transitions, immutable
audit and ledger history, and bounded trust-path queries. The founder selected
MongoDB as the initial system of record and asked that a graph database be
introduced only when needed. This choice must not weaken transactional
invariants, idempotency, retention, auditability, or recovery objectives.

## Decision

Use MongoDB as the authoritative operational datastore for the initial
platform.

- Design bounded aggregates and keep unbounded messages, events, ledger entries,
  and audit records in separate append-oriented collections.
- Use JSON Schema validation, explicit collection ownership, documented indexes,
  and query-plan evidence for production queries.
- Use unique indexes, atomic conditional updates, aggregate versions, and
  multi-document transactions where invariants cross document boundaries.
- Commit durable outbox entries in the same transaction as domain changes.
- Model trust relationships as scoped edge collections. Use bounded traversal
  and materialized trust-path projections rather than speculative graph
  infrastructure.
- Store media in object storage. MongoDB keeps metadata, ownership, checksums,
  retention state, and signed-access policy.
- Isolate identity and biometric data behind narrower credentials and
  application-level envelope encryption. TTL deletion is allowed only where it
  is legally correct and compatible with legal holds.
- Immutable double-entry records, not mutable balances, are the source for
  commerce and allowance accounting.
- Integration tests run against a real ephemeral MongoDB using Testcontainers;
  repository mocks do not count as persistence verification.
- Production configuration must demonstrate the target RPO of at most five
  minutes and RTO of at most sixty minutes before launch.

## Consequences

The team operates one primary database technology while retaining transactional
and audit requirements. Data modelling and index discipline become mandatory;
MongoDB flexibility is not permission for schemaless production data. Graph
projections may add write amplification and must be rebuildable.

## Revisit triggers

Evaluate a graph store behind an application port only after representative
trust-path workloads exceed agreed latency/cost targets despite indexes,
bounded traversal, caching, and materialized projections. Evaluate another
specialized store if measured transaction, search, analytics, or residency
requirements cannot be met safely. Any migration requires dual-write/replay,
reconciliation, rollback, and ownership plans in a new ADR.
