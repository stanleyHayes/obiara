# MongoDB Atlas environment contract

Status: configuration specification only; no Atlas or cloud resource has been
created.

This directory defines E17-S02 controls for the authoritative MongoDB
operational store. The files are inputs to a future reviewed provisioning
workflow, not Atlas API credentials or Terraform state.

## Environments

- `staging.yaml` permits only synthetic data and references a synthetic restore
  rehearsal record.
- `production.yaml` is an Africa-region candidate in AWS Cape Town. It remains
  explicitly `blocked`; its empty restore evidence reference is intentional.
  It cannot become an approved topology until a signed founder/DPO/legal
  residency decision, DPIA, Atlas tier/DPA/support/procurement review, latency
  evidence, production topology ADR, and production restore rehearsal exist.

Both documents require three electable nodes across three availability zones,
termination protection, TLS 1.2 or newer, no public access, exact external
allowlist references, customer-managed at-rest encryption, application-level
C4 field encryption, continuous localized backup and PITR. RPO is at most five
minutes and RTO at most sixty minutes.

## Identity boundaries

API, worker, C4 identity, C4 safety, backup and restore credentials are unique.
General application roles cannot read C4 collection families. Backup can only
read backup material; restore can only create an isolated target. References
name secret-manager slots—their values must never enter Git.

The Render/Atlas split cannot use same-region Private Link. Before approval,
procurement must resolve the named egress CIDR references and validate public
internet TLS/allowlist behavior. If compute becomes colocated, a private
endpoint is mandatory.

## Validation

Run:

```sh
go run ./internal/quality/atlasconfig/cmd .
go test ./internal/quality/atlasconfig/...
```

The validator rejects wildcard ingress, public access, shared credentials,
broad roles, missing C4 isolation, weak encryption, cross-region backup,
missing PITR, RPO/RTO regression, destructive restore, fake production
readiness, secret-shaped content and invalid restore evidence.

`restore-evidence.schema.json` is the handoff contract for E17-S07. Restore
evidence must identify an isolated target, distinct data-owner and security
approvals, PITR use, validation digest, and measured RPO/RTO. The committed
staging record is explicitly synthetic and is not evidence of a production
restore.
