# `@obiara/quality-client`

Deterministic localization, copy-policy, and source-quality checks shared by
Obiara clients. The package contains testable library functions and a CLI.

## Enforcement rollout

Existing client prototypes are intentionally in **staged** mode: CI executes
the package tests, while application ESLint configurations can adopt
`productQualityConfig({ enforcement: "warn" })` without turning inherited copy
debt into blanket failures. New or migrated surfaces should use `"error"`.
Set `OBIARA_QUALITY_ENFORCEMENT=strict` when running the CLI to fail on findings.

## Narrow escape hatch

Only literal non-product data may bypass the hardcoded-copy source check:

```tsx
// quality-ignore-next-line non-product-data: upstream provider label
const value = <span>Provider status</span>;
```

The classification must be exactly `non-product-data`, include a meaningful
reason, and applies only to the immediately following line. It must not be used
for product instructions, actions, errors, labels, marketing, or consent copy.
