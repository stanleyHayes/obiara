# platform/mongo — MongoDB conventions

Platform-level MongoDB helpers for the api service (`agent_plan.md` §7.4,
§8). Module adapters build repositories on top of this package; driver types
never cross into domain code.

## Client setup

`Connect(ctx, uri)` is the single client construction path. Callers set the
deadline via context (`MONGO_CONNECT_TIMEOUT` in `platform/config`). The
worker service reuses this package when it lands (S1-009).

## Repositories

- One repository per aggregate, owned by the module's
  `adapters/outbound/mongodb` package, implementing an application-layer
  port (see the `member` module for the template).
- Bounded aggregates; high-volume append-only events get their own
  collections (`agent_plan.md` §8).
- Translate driver errors with `IsDuplicateKey` into domain-meaningful
  results; never leak driver error types past the adapter.
- Optimistic concurrency: documents that carry state machines include a
  version field and update with `findOneAndUpdate` on
  `{_id, version}` filters.

## Transactions

`WithTransaction(ctx, client, fn)` guards multi-document invariants (ledger
pairs, mutual-water room creation, outbox-committed-with-domain-change).
Single-document writes must not use transactions.

## Migrations and indexes

- Schema changes and index builds ship as `Migration` values applied by the
  `Runner` at service start (pre-deploy job in Render; see `agent_plan.md`
  §12 expand/migrate/contract).
- Migration IDs are append-only (`NNNN_snake_case`); applied migrations are
  never edited. A recorded migration is a no-op on re-run, and concurrent
  starters are safe via the unique `_id` on `schema_migrations`.
- Module `EnsureIndexes` helpers (e.g. the member repository's) are wrapped
  as migrations rather than run ad hoc.

## Testing

- Unit tests cover validation and error translation without a database.
- Integration tests use the `integration` build tag and Testcontainers
  (`mongo:8.0.13`); they must fail with an actionable message when no
  container runtime is available and must never silently skip
  (`agent_plan.md` §13).
