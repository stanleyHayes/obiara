# ADR-0006: Deploy the initial platform with a Render Blueprint

- Status: Accepted
- Date: 2026-07-26
- Owners: Platform engineering
- Supersedes: Kubernetes and Terraform default in source architecture

## Context

The initial platform needs repeatable dev, staging, and production deployments
without the operational cost of a Kubernetes platform. API and worker processes
are container-friendly, while MongoDB, object storage, communications, identity,
and live audio are managed boundaries. Data location and provider suitability
must be verified before production commitments.

## Decision

Describe Render-hosted infrastructure in a repository-root `render.yaml` and
keep supporting environment documentation under `deploy/render/`.

- Deploy the stateless API and worker as separate services with independent
  scaling, commands, health checks, and least-privilege credentials.
- Define `/live` and `/ready` semantics; readiness must distinguish critical
  dependency failure from explicitly supported degraded operation.
- Use dev, staging, and production environments. Staging uses synthetic data;
  production member data must not be copied into lower environments.
- Keep secrets out of Git and inject them through managed environment groups or
  an approved secret store.
- Use immutable build inputs, pinned toolchains/dependencies, migration
  compatibility checks, staged rollout procedures, and documented rollback.
- Keep production state in approved managed services; do not rely on ephemeral
  service disks for durable truth.
- Add observability, backup verification, restore exercises, resource limits,
  scaling thresholds, and cost alerts before production.
- Verify current Render regions and every managed provider's processing,
  storage, backup, and support location during Sprint 0. This ADR does not itself
  approve a residency exception or paid provider contract.

## Consequences

The team can reproduce platform topology with lower operational overhead.
Render feature and region constraints become explicit dependencies. Managed
service outages and portability require tested degradation, export, and restore
paths rather than assumptions.

## Revisit triggers

Choose another platform only when verified residency/legal requirements,
network topology, workload shape, security isolation, reliability targets, or
cost cannot be met on Render. A replacement ADR must include provider evidence,
migration and rollback plans, recovery objectives, cost, staffing impact, and
the founder/legal approvals required by the execution plan.
