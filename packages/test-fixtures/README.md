# @obiara/test-fixtures — Synthetic Personas and Fixture Policy

This package is the single source of synthetic test actors and shared
fixtures for unit, integration (Testcontainers), e2e, golden-path monitoring
and staging environments.

## Fixture policy (binding)

1. **Synthetic only.** No production data, no real member data, and no real
   provider credential may ever appear in this package, in test code, or in
   committed fixtures (`agent_plan.md` §14).
2. **Obviously-fake identifiers.** Phone numbers use the reserved fictional
   pattern `+23355000xxxx`. Ghana Card numbers use the clearly invalid
   pattern `GHA-000000000-X`. Email addresses use `*.test`. Names are common
   Ghanaian first names only — never full names of real people.
3. **Marker field.** Every persona carries `synthetic: true`; backend seed
   paths must refuse to insert personas without it outside test/staging
   configuration. Golden-path synthetic actors (`agent_plan.md` §13) are
   isolated from real members by this marker plus dedicated circle/fixture
   scoping.
4. **Deterministic.** Personas are stable constants with fixed IDs so tests
   are reproducible and golden-path monitors can assert on known state.
5. **Privacy-shaped even in tests.** Personas model consent states, tiers
   and verification states explicitly so tests exercise the consent map
   (Doc 08 §8) rather than bypassing it.
6. **No secrets.** Fixtures never contain OTPs, tokens, or keys that would
   work against a real provider; provider adapters are exercised through
   contract tests and simulators (`agent_plan.md` §23 risk mitigation).
7. **Moderation-welfare aware.** Abuse/fraud test content uses the mildest
   fixture that exercises the classifier path; graphic content is never
   committed (Doc 09 §7 workforce welfare).

## Persona registry

See `src/personas.ts`. Coverage targets:

- Tier ladder: unverified (Tier 0), verified Tier 1, sowing-capable Tier 2.
- Age gate: an under-18 attempt persona for the FR-104 purge-proof tests.
- Golden path: a sow/sprout/room pair (`ama.sow` ↔ `kwame.sprout`) for the
  register→verify→sow→sprout→room synthetic monitor (NFR-602).
- Safety: a scam-arc actor for Sentinel/Sika Shield tests (E11-S10/S11) and
  a distressed-member persona for care-queue routing tests (E12-S05).
- Operations: a circle host and a T&S admin for authorization-matrix tests
  (FR-801).
- Device class: a low-end Android/3G profile marker for client budget tests
  (NFR-101–106), used by mobile/web test harnesses.

Adding a persona: append to `PERSONAS`, keep IDs kebab-cased under the
`obiara.test.` namespace, and state which requirement/test class it serves.
