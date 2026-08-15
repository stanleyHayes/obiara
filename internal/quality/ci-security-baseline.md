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

## Bootstrap findings and remediation

The 2026-07-26 local audit initially detected four high advisories in transitive
packages: Next.js paths to `sharp` and `postcss`, plus ESLint paths to
`brace-expansion`. S0-015 resolved them with exact root overrides for
`sharp@0.35.3`, `postcss@8.5.23`, and the compatible
`minimatch@10.2.5`/`brace-expansion@5.0.8` pair. The frozen install, peer check,
full lint/type/test/Go test suite, Go vet, and all client builds passed after
remediation. The final audit reported zero high and zero critical findings; no
waiver was used.

## 2026-08-15 advisory wave and image-size exception

A post-completion audit found five new high advisories in transitive packages.
Exact root overrides remediated three of them (`brace-expansion@5.0.9`,
`js-yaml@4.3.1`, `nanoid@3.3.18` in `pnpm-workspace.yaml`).

The remaining two have no patched upstream release (vulnerable `<=2.0.2`;
latest published is 2.0.2), so they are held as a temporary exception under
the policy above:

- Advisories: GHSA-w3rx-r6r6-pgpr and GHSA-5p2g-fcmc-qvqq (`image-size` ICNS
  and JXL/HEIF parser infinite-loop DoS).
- Owner: `/root/platform_scaffold`.
- Rationale: no fix exists to upgrade to; both consumers are dev/CI tooling
  only (`metro` bundler and `vite-plugin-storybook-nextjs`). The production
  runtime (Next.js apps, Go services) never loads `image-size`.
- Compensating controls: reachable only when a developer or CI builds with a
  deliberately crafted image asset; all assets are repository-reviewed; CI
  runners are ephemeral.
- Mechanism: `auditConfig.ignoreGhsas` in `pnpm-workspace.yaml`; the CI gate
  counts severities from the `advisories` map (which honors the ignore list)
  rather than the stale `metadata.vulnerabilities` totals.
- Expiry: 2026-11-15, or immediately when `image-size` 2.0.3+ publishes.
- Remediation task: replace the exception with an exact override to
  `image-size@>=2.0.3` once released; verify metro and Storybook still build.
