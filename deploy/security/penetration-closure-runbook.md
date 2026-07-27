# Security assessment and finding closure

This control separates a CI security exercise from an independent penetration
test. The committed evidence is deliberately marked `synthetic-ci-exercise`;
it validates the evidence contract but can never approve production.

## DAST boundary

- CI starts the repository-owned fixture on exactly `http://127.0.0.1:18080`.
- ZAP 2.17.0 is pinned by immutable Linux/amd64 image digest and runs the
  bounded passive baseline only.
- The workflow has no credential, browser session, production hostname, webhook,
  issue-write permission, active-scan option, or arbitrary target input.
- Required response protections are failures. TLS is the sole ignored rule
  because the loopback fixture deliberately has no public transport boundary.

## Independent assessment intake

Before a production decision, replace the synthetic record with a separately
reviewed `independent-penetration-test` record. Use only opaque assessment,
service, finding, remediation, retest and role references. Never attach request
or response bodies, member identifiers, credentials, exploit payloads or
provider exports to Git.

The assessment must cover `service:api`, `service:web`, `service:admin` and
`service:mobile`; be no older than 90 days; and identify an assessor. Every
finding must be closed, with an owner, immutable remediation reference, retest
reference, closure time, and a verifier distinct from both assessor and owner.
High and critical findings cannot be accepted or waived. In this repository,
production remains blocked unless all findings are closed.

## Triage and closure

1. Assign severity using the assessor's finite critical/high/medium/low scale.
2. Record only opaque references and set the finding to `open` or `in-progress`.
3. Remediate on a reviewed commit and record the immutable commit reference.
4. Have a distinct security verifier retest the exact affected surface.
5. Mark `closed` only after the retest evidence exists.
6. Re-run the deterministic validator and the security workflow.

Any missing, malformed, stale, synthetic, partially scoped, unresolved,
self-verified, or accepted finding keeps production fail-closed. A human release
review must additionally satisfy the deployment, residency and environment
approval gates; this evidence never deploys or changes an external account.
