# Obiara Architecture Decision Records

Architecture Decision Records (ADRs) capture consequential technical choices,
their constraints, and the evidence required to revisit them. They supplement
the product requirements; they do not override product laws, consent rules,
safety controls, or founder decisions in `agent_plan.md`.

## Status vocabulary

- **Proposed**: under review and not yet an implementation constraint.
- **Accepted**: the default for new implementation.
- **Superseded**: replaced by a later ADR, which must link back to it.
- **Deprecated**: retained for historical context but no longer recommended.

## Index

| ADR | Decision | Status |
|---|---|---|
| [0001](0001-monorepo.md) | One polyglot monorepo | Accepted |
| [0002](0002-go-hexagonal-modular-monolith.md) | Go hexagonal modular monolith | Accepted |
| [0003](0003-mongodb-operational-store.md) | MongoDB as initial operational store | Accepted |
| [0004](0004-rest-openapi-client-contract.md) | REST and OpenAPI 3.1 client contract | Accepted |
| [0005](0005-expo-cross-platform-mobile.md) | Expo cross-platform mobile client | Accepted |
| [0006](0006-render-blueprint-deployment.md) | Render Blueprint deployment | Accepted |

## Maintenance

ADRs are immutable after acceptance except for typo/link fixes and status
changes. A material change requires a new ADR that supersedes the old one.
Every implementation should reference the applicable ADR in its task evidence.
