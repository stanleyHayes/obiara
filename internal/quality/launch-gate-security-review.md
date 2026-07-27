# Launch-gate security and privacy cross-review

Review owner: `/root/security_closure`  
Boundary reviewed: executable production-gate registry and evidence evaluator  
Decision: production must remain fail-closed until every repository and external
gate has current, environment-bound, kind-correct evidence.

## Trust boundary

Repository evidence can prove only repository-controlled facts such as an exact
commit's tests, build, policy validation, and immutable configuration. It cannot
prove a regulator decision, provider account state, credential provisioning,
store approval, residency outcome, independent review, or production action.

All evidence is metadata-only. A `production_action` record can prove that a
separately controlled authorization was reviewed; evaluating it must not deploy,
activate, provision, submit, notify, call a provider, or mutate an account.

## Required adversarial checks

| Attack or error | Required outcome |
|---|---|
| Synthetic fixture presented as real evidence | Reject or retain explicit blocker |
| Evidence completed after its expiry or evaluated after expiry | Reject as stale |
| Evidence issued for preview, local, test, or staging | Reject for a production gate |
| Repository evidence used for an external, provider, credential, store, or production-action gate | Reject kind mismatch |
| Duplicate evidence ID or duplicate gate ID | Reject the entire registry/evidence set |
| Required dependency is missing or blocked | Keep dependent gate blocked |
| Evidence owner also acts as required approver/verifier | Reject self-approval |
| Approval roles contain duplicate actors | Reject insufficient separation |
| Evidence ref contains a URL, credential/token shape, email, phone, member ID, request/response body, or free text | Reject as unsafe metadata |
| Unknown field, evidence kind, environment, status, severity, or gate | Reject rather than ignore |
| External result merely declared `approved` without bounded provenance and current evidence | Keep the gate blocked |
| Production authorization evidence is valid | Return a decision only; perform no production action |
| One gate remains unresolved | Overall production eligibility is false |

## Evidence minimization

Allowed references are opaque, bounded identifiers for a gate, assessment,
artifact, immutable commit, provider decision, role, and verifier. Evidence must
not embed or link to member data, content, credentials, secret digests, raw
provider payloads, scan bodies, request/response bodies, or exploit material.
Logs and error messages must identify only the gate/evidence reference and a
finite reason code.

## External approval law

External evidence needs a finite external kind, exact `production` environment,
issued and expiry times, bounded opaque provenance, an owner, and every distinct
approval role required by the gate. Repository CI must not manufacture any of
these facts. Missing, synthetic, stale, wrong-environment, wrong-kind, duplicate,
unknown, self-approved, or dependency-blocked evidence is a blocker.

The evaluator is a pure decision boundary. It produces explicit blockers and
never changes deployment, feature flags, stores, credentials, providers,
residency records, launch state, or production infrastructure.

## Findings and bounded corrections

1. **Corrected — evidence replay across gates.** The first review found that
   duplicate gate IDs were rejected but the same immutable `evidenceRef` could
   be reused for two different gates. L1-002 now enforces global evidence
   reference uniqueness and both owner and cross-review suites cover the case.
2. **Corrected — incomplete order-invariance proof.** The initial property test
   asserted an incidental blocker ordering property rather than comparing the
   shuffled decision with its baseline. It now compares the complete decision;
   the cross-review also independently reverses every evidence item and requires
   an identical decision.
3. **Verified — adversarial boundary.** Independent tests cover synthetic and
   stale evidence, stale dependency propagation, wrong environment,
   self-approval, secret/PII/URL-shaped actor references, repository evidence
   substituted for an external gate, and activation evidence presented while
   an external dependency remains pending.

No external approval is marked satisfied by this review. The committed fixture
remains synthetic and the production decision remains blocked.
