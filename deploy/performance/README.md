# Performance, load and cost evidence

These profiles are bounded local engineering evidence for E17-S08. They never
target Render, Atlas, a paid provider, production traffic or member data.
`budgets.yaml` fixes sample counts, concurrency, percentile and error budgets so
the result cannot hide its denominator. The HTTP profile exercises the real
`/live` handler in memory. The Mongo profile uses a temporary MongoDB 8.0.13
Testcontainer with a bounded connection pool and synthetic documents.

Run the deterministic gates:

```sh
go test ./internal/quality/performance ./services/api/internal/platform/health
go test -tags=integration ./internal/quality/performance -run TestMongoBoundedLoad
go test -race ./internal/quality/performance ./services/api/internal/platform/health
go test -bench=. -benchmem ./internal/quality/performance
```

`cost-assumptions.yaml` is a transparent sensitivity model in USD cents. Values
are committed planning assumptions, not current provider quotes, invoices,
procurement approval or a capacity promise. The 100k-MAU scenario is calibrated
against Doc 07's order-of-magnitude envelope, while the pilot and 1m scenarios
make scaling sensitivity visible. Before a release decision, replace
assumptions with dated quotes and measured usage through a separately reviewed
evidence change.

Local wall-clock results are regression signals, not SLO proof. NFR-100 still
requires budget-Android/3G device-lab evidence; NFR-201/NFR-202 still require
staging peak and live-audio capacity work. A green local profile cannot approve
production scale or spend.
