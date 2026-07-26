# CI, security, and supply-chain baseline

- Status: Accepted Sprint 0 baseline
- Task: S0-006
- Owner: `/root/platform_scaffold`
- Established: 2026-07-26

## Required checks

Every push and pull request targeting `main` runs:

1. An exact pnpm 11.17.0 frozen-lockfile install with build-script allowlisting.
2. Peer dependency validation.
3. Workspace lint, type-check, unit-test, and Go test commands.
4. `go vet ./...`.
5. Production builds for member web, admin, and the Expo web export.
6. A clean-tree check to detect uncommitted generated drift.

GitHub Actions run on pinned immutable commit SHAs. Version comments make
automated upgrades reviewable. Workflow permissions default to read-only and
are elevated only for the CodeQL job's security result upload.

## Security checks

CodeQL scans Go and JavaScript/TypeScript on pushes, pull requests, a weekly
schedule, and manual dispatch. The dependency job runs pnpm's high-severity
audit gate and Go's official vulnerability scanner.

High or critical findings fail the check. The pnpm gate evaluates the JSON
severity counts directly because package-manager CLI exit behavior has varied.
A temporary exception requires an owner, rationale, compensating controls,
expiry date, and a tracked remediation task. Secrets, production data, and
credentials must never be committed or printed in workflow logs.

## Dependency updates

Dependabot proposes weekly GitHub Actions, npm/pnpm, and Go module updates.
Updates are not auto-merged. The dependency matrix compatibility rules still
apply, especially the Expo-aligned React boundary and TypeScript/ESLint pins.
Lockfile changes require the same install, peer, check, build, and security
evidence as application changes.

## Local reproduction

Run from repository root:

```sh
npx --yes pnpm@11.17.0 install --frozen-lockfile
npx --yes pnpm@11.17.0 peers check
npx --yes pnpm@11.17.0 run check
go vet ./...
npx --yes pnpm@11.17.0 run build
go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 .github/workflows/*.yml
```

Workflow syntax is checked with `actionlint`. Network-backed audit results are
time-sensitive and must be rerun in CI even when local validation passes.

## Bootstrap findings

The 2026-07-26 local audit detected four high advisories in transitive packages:
Next.js paths to `sharp` and `postcss`, plus ESLint paths to
`brace-expansion`. The baseline deliberately fails on these findings. They
require a separately coordinated lockfile remediation because client tasks were
actively changing the shared dependency graph during S0-006; they are not
silently waived.
