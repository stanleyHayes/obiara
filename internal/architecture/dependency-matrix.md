# Toolchain and dependency compatibility matrix

- Status: Accepted bootstrap baseline
- Resolved: 2026-07-26
- Owner: `codex-root`
- Task: S0-004

This matrix records the stable versions discovered at bootstrap. Application
manifests must pin exact versions. Expo-governed packages use Expo's published
compatible versions even when a newer standalone release exists.

## Resolution rules

1. Use the latest stable non-prerelease release that satisfies the selected
   platform's engine and peer constraints.
2. Prefer one version across workspaces unless a platform compatibility
   boundary requires otherwise.
3. Expo's `bundledNativeModules.json` is authoritative for Expo/RN packages.
4. Exact versions belong in manifests and the pnpm lockfile. Go dependencies
   belong in `go.mod`/`go.sum`; the Go toolchain belongs in `go.mod`.
5. A newer incompatible release is not silently forced with peer-dependency
   overrides. Record the compatibility exception here.
6. Re-resolve versions during a deliberate upgrade task; do not float versions
   during ordinary installs.

## Toolchains

| Tool | Bootstrap version | Rationale / constraint |
|---|---:|---|
| Go | `1.26.5` | Current stable from `go.dev/VERSION` |
| Node.js | `26.5.0` | Current installed stable; satisfies Next and Expo (`>=22.13`) |
| npm | `11.17.0` | Current stable discovery/install client |
| pnpm | `11.17.0` | Current stable workspace package manager; pin with `packageManager` |
| Turborepo | `2.10.7` | Current stable task orchestrator |
| TypeScript | `5.9.3` | Latest stable accepted by `typescript-eslint@8.52.0`, whose peer range is `>=4.8.4 <6.0.0`; TypeScript 7 cannot be used with the current lint stack |

`corepack` is not available in the current Node installation. Bootstrap must
either install the exact pnpm version explicitly or use pnpm's supported
self-management path; scripts must not assume `corepack` exists.

## Web and admin

| Package | Bootstrap version | Compatibility evidence |
|---|---:|---|
| `next` | `16.2.12` | Current stable; Node engine `>=20.9.0`; React 19 supported |
| `react` | `19.2.3` | Expo 57 compatible version; shared across clients to avoid duplicate React |
| `react-dom` | `19.2.3` | Expo 57 compatible version; valid for Next 16 |
| `@mui/material` | `9.2.0` | Current stable; peer range includes React 19 |
| `@mui/icons-material` | `9.2.0` | Aligned with MUI core |
| `@tanstack/react-query` | `5.101.4` | Current stable |
| `zod` | `4.4.3` | Current stable candidate for client-safe validation |
| `resend` | `6.18.0` | Current stable; server-only use |
| `openapi-typescript` | `7.13.0` | Current stable contract type generator candidate |
| `@hey-api/openapi-ts` | `0.99.0` | Current stable SDK generator candidate; choose one generator in S0-005 |

React `19.2.8` is the newest standalone release discovered, but Expo 57 pins
React `19.2.3`. Obiara uses `19.2.3` across the monorepo until Expo publishes a
compatible update. This is a compatibility exception, not an accidental stale
pin.

## Mobile

| Package | Bootstrap version | Compatibility evidence |
|---|---:|---|
| `expo` | `57.0.8` | Current stable SDK; Node engine `>=22.13` |
| `expo-router` | `57.0.8` | Expo 57 bundled compatible version |
| `expo-status-bar` | `57.0.1` | Latest published stable; `57.0.8` does not exist in the registry |
| `react-native` | `0.86.0` | Expo 57 bundled compatible version |
| `react` | `19.2.3` | Expo 57 bundled compatible version |
| `react-dom` | `19.2.3` | Expo 57 bundled compatible version |
| `react-native-web` | `0.21.0` | Expo 57 compatible range; standalone latest is `0.21.2` |
| `react-native-gesture-handler` | `2.32.0` | Expo 57 bundled compatible version |
| `react-native-reanimated` | `4.5.0` | Expo 57 bundled compatible version |
| `react-native-screens` | `4.26.x` | Expo 57 bundled compatible range |
| `react-native-safe-area-context` | `5.7.x` | Expo 57 bundled compatible range |
| `react-native-worklets` | `0.10.3` | Latest release satisfying Expo 57 module peer range `^0.7.4` through `^0.10.0` |

All Expo libraries must be installed through `expo install` or checked against
the pinned SDK's bundled module map. Android 8 support remains unconfirmed until
S0-010 produces a native build and reference-device evidence.

## Go backend and integration testing

| Module | Bootstrap version | Purpose |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.8.0` | MongoDB adapter |
| `github.com/go-chi/chi/v5` | `v5.3.1` | Minimal HTTP routing candidate |
| `github.com/testcontainers/testcontainers-go` | `v0.43.0` | Real dependency integration tests |
| `github.com/testcontainers/testcontainers-go/modules/mongodb` | `v0.43.0` | MongoDB integration-test module |
| `go.uber.org/mock` | `v0.6.0` | Application-port mocks |
| `go.opentelemetry.io/otel` | `v1.44.0` | Tracing/metrics API and SDK baseline |
| `github.com/stretchr/testify` | `v1.11.1` | Optional assertion helpers; standard library remains sufficient by default |

Testcontainers is mandatory for MongoDB repository integration tests. Such
tests must not silently skip when Docker is unavailable.

## Quality tooling

| Package | Bootstrap version | Purpose |
|---|---:|---|
| `eslint` | `9.39.2` | Latest stable accepted by `typescript-eslint@8.52.0`, whose peer range is `^8.57.0 || ^9.0.0`; ESLint 10 is not compatible |
| `prettier` | `3.9.6` | Formatting |
| `vitest` | `4.1.10` | TypeScript unit/component test candidate |
| `@testing-library/react` | `16.3.2` | Accessible web component tests |
| `@playwright/test` | `1.62.0` | Browser end-to-end and rendered UI checks |

ESLint and TypeScript are ecosystem-sensitive major versions. S0-005 must prove
that the chosen Next, Expo, React Native and MUI lint/build plugins support these
versions. If not, pin the newest compatible stable versions and append the
exact peer constraint and evidence here.

## Verification record

Discovery commands:

```text
npm view <package> version --json
npm view <package> engines --json
npm view <package> peerDependencies --json
npm pack expo@57.0.8
go list -m -versions <module>
curl -fsSL https://go.dev/VERSION?m=text
```

Observed compatibility:

- Next 16.2.12 accepts React 19 and Node >=20.9.
- MUI 9.2.0 accepts React/React DOM 19.
- Expo 57.0.8 requires Node >=22.13 and bundles React 19.2.3 with React Native
  0.86.0.
- Current Node 26.5.0 satisfies the documented Next, Expo and React Native
  engine ranges.
- Current Go 1.26.5 successfully resolved the listed module versions.
- Exact pnpm 11.17.0 frozen install completed with no peer-dependency issues.
- TypeScript 5.9.3 and ESLint 9.39.2 are compatibility pins. Registry peer
  metadata for `typescript-eslint@8.52.0` rejects TypeScript 6/7 and ESLint 10.
- `pnpm run check` passed lint, type-check and empty-skeleton unit-test commands
  for member web, admin and mobile, plus `go test ./...` for API and worker.
- `pnpm run build` passed production Next builds for web/admin and an Expo web
  export for mobile.
