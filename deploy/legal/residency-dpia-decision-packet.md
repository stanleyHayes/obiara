# Residency and DPIA external decision packet

This packet requests metadata for a decision by the founder, qualified DPO/legal
authority, and an independent reviewer. It is not legal advice, a DPIA, a
signature, an approval, or evidence that any production topology is lawful.

## Materials for the external reviewers

- `internal/architecture/data-classification.md`
- `internal/architecture/dpia-inputs.md`
- `internal/architecture/threat-model-v0.md`
- `deploy/render/provider-residency-feasibility.md`
- `deploy/render/environment-matrix.md`
- `deploy/atlas/production.yaml`
- `deploy/secrets/inventory.yaml`
- `deploy/observability/slo.yaml`
- `deploy/security/penetration-closure-runbook.md`

## Decision questions

1. Does the binding interpretation require every processing and storage
   boundary to be Ghana-only, or does it permit an exact Africa-region topology
   with separately assessed transfers?
2. Does the assessment cover identity/authentication, biometric liveness,
   profiles/matching, private voice/media, community/trust, safety/care
   evidence, AI/language processing, commerce/financial records, and
   analytics/audit/operations?
3. Are exact locations recorded for compute, operational data, backups, object
   storage, CDN caches, logs, live-media network and egress/recording,
   transactional email, AI processing, identity verification, and provider
   support access?
4. Does the transfer assessment cover every non-Ghana processing or access
   location, processor/subprocessor duties, retention/deletion, breach terms,
   key ownership, support access, and exit/export?
5. Does the DPIA record necessity, proportionality, risks, mitigations,
   consultation needs and residual risk for the complete processing scope?

## Metadata handoff

External reviewers create a fresh JSON record conforming to
`residency-dpia.schema.json`. Use only opaque role and SHA-256 evidence
references. Do not commit names, emails, phone numbers, signatures, legal
opinions, provider payloads, credentials, member data, request/response bodies
or secret digests.

The three authority slots are mandatory and distinct:

- `founder`
- `dpoLegal`
- `independentReviewer`

Each authority provides an opaque actor reference, an opaque evidence reference
for the separately retained signature/approval record, and its signing time.
The repository stores no signature material.

`ghana-only` means every required boundary has country code `GH`.
`africa-region` still requires an exact country and non-ambiguous provider
region for every boundary, plus a complete transfer assessment reference.
Values such as `Africa`, `Ghana`, `global`, `automatic`, `nearest`, `unknown`
or `undecided` never satisfy the validator.

The decision expires within 90 days and is bound to the exact production
candidate SHA. Any unsigned, self-approved, stale, synthetic, incomplete,
repository-issued, wrong-environment, ambiguous-location, pending, blocked or
declined record keeps production blocked. Passing this contract produces only
decision metadata; it never provisions, deploys, activates or changes a
provider.
