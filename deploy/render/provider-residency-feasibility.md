# Provider and data-residency feasibility

- Task: S0-011
- Evidence checked: 2026-07-26
- Status: Technical spike complete; production residency decision blocked on
  founder and Ghana privacy/legal approval
- Scope: Evidence and architecture options only. No provider account, resource,
  contract, or deployment was created.

## Requirement being tested

The Obiara plan requires low latency for members in Ghana and says member
content must remain "in-region." It also requires a DPIA and explicit approval
before production if the selected providers cannot meet that posture.

This spike uses two interpretations so the decision is not hidden:

1. **Ghana-only:** personal data, member content, backups, and relevant
   processing must remain physically in Ghana.
2. **Africa-region:** those workloads may remain within Africa, subject to an
   approved cross-border-transfer assessment and contracts.

Technical availability does not establish legal compliance. Ghana's current
Data Protection Act, 2012 (Act 843) is the applicable primary legal source
identified in this spike. The Data Protection Commission also publishes a 2025
draft bill, but a draft is not treated as enacted law here. Counsel or the
organization's qualified privacy owner must document the applicable
cross-border-transfer basis, processor obligations, sensitive-data controls,
and DPIA outcome.

## Verified provider matrix

| Boundary                                         | Current primary-source fact                                                                                                                                                                                                                                                                 | Ghana-only                                            | Africa-region                                                          | Constraints and evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- | ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Render compute: Go API, worker, Next.js services | Render lists Oregon, Ohio, Virginia, Frankfurt, and Singapore. It does not list Ghana or any African region. A service's region cannot be changed in place.                                                                                                                                 | **No**                                                | **No**                                                                 | Render [Regions](https://render.com/docs/regions) and [Blueprint YAML reference](https://render.com/docs/blueprint-spec), checked 2026-07-26. Frankfurt is the nearest listed Render geography, but is outside Africa.                                                                                                                                                                                                                                                                                          |
| Render service networking                        | Services share a private network only when they are in the same Render region. Render Private Link requires Pro or higher, connects to AWS-hosted providers, and is same-region only.                                                                                                       | **No**                                                | **No** for African state; partial for Frankfurt-aligned external state | Render [Private Network](https://render.com/docs/private-network) and [Private Link Connections](https://render.com/docs/private-link), checked 2026-07-26. A Render Frankfurt service cannot use Render's same-region Private Link to an Atlas cluster in Cape Town.                                                                                                                                                                                                                                           |
| MongoDB Atlas operational store                  | Atlas supports AWS `af-south-1` in Cape Town and Azure `southafricanorth` in Johannesburg / `southafricawest` in Cape Town. Atlas supports regionalization and localized cloud backups; continuous cloud backup can provide an RPO as low as one minute.                                    | **No documented Ghana region**                        | **Technically feasible in South Africa**                               | MongoDB [AWS regions](https://www.mongodb.com/docs/atlas/reference/amazon-aws/), [Azure regions](https://www.mongodb.com/docs/atlas/reference/microsoft-azure/), [compliance architecture](https://www.mongodb.com/docs/atlas/architecture/current/compliance/), and [disaster recovery](https://www.mongodb.com/docs/atlas/architecture/current/disaster-recovery/), checked 2026-07-26. Exact tier, backup-copy regions, logs, support access, and contract terms still require confirmation before purchase. |
| Object storage and CDN                           | No object-storage vendor is selected. AWS documents `af-south-1` (Cape Town) as a three-AZ region, and its launch material includes S3 and states customer content can be stored in South Africa.                                                                                           | **No verified Ghana option in the current shortlist** | **Candidate exists in South Africa**                                   | AWS [available regions](https://docs.aws.amazon.com/global-infrastructure/latest/regions/aws-regions.html) and [Cape Town launch](https://aws.amazon.com/blogs/aws/now-open-aws-africa-cape-town-region/), checked 2026-07-26. A production selection still needs service-specific S3 replication, backup, CDN/cache, log, deletion, key-management, DPA, and price verification.                                                                                                                               |
| LiveKit realtime audio/video                     | LiveKit Cloud supports protocol-based pinning to an `africa` region located in South Africa. Pinning is available on Scale or higher, must be enabled by LiveKit, disables out-of-region automatic failover, and the Africa region currently has one location without in-region redundancy. | **No**                                                | **Conditionally feasible**                                             | LiveKit [Regions](https://docs.livekit.io/deploy/admin/regions/) and [Region pinning](https://docs.livekit.io/deploy/admin/regions/region-pinning/), last-updated by LiveKit 2026-05-14 and checked 2026-07-26. Pinning covers LiveKit Cloud network traffic; recording, egress, support/telemetry, and any agent workloads require separate confirmation.                                                                                                                                                      |
| Resend transactional email                       | Resend states that customer data is stored in the United States and that its primary processing operations occur in the United States. Email region selection concerns sending infrastructure and is not evidence of customer-data residency.                                               | **No**                                                | **No**                                                                 | Resend [GDPR statement](https://resend.com/security/gdpr), [DPA](https://resend.com/legal/dpa), and [subprocessors](https://resend.com/legal/subprocessors), checked 2026-07-26. Keep sensitive member content out of email; legal/privacy must approve the transfer posture or select a different provider.                                                                                                                                                                                                    |

## Feasible topology options

| Option                           | Topology                                                                                                                                                 | Benefits                                                                                                                       | Blocking issues                                                                                                                                                                                                                            | Disposition                                                    |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------- |
| A. Render-aligned Europe         | Render Frankfurt; Atlas and object storage in compatible European regions; LiveKit EU pinning                                                            | Simpler regional alignment and potential same-region private connectivity to an AWS-hosted dependency on qualifying paid plans | Neither Ghana-only nor Africa-region; all relevant cross-border processing and transfers need approval; Ghana latency must be measured                                                                                                     | **Not approved**                                               |
| B. Split Render and South Africa | Render Frankfurt compute; Atlas Cape Town/Johannesburg; object storage Cape Town; LiveKit Africa pinning                                                 | Keeps primary durable member state and realtime traffic in Africa                                                              | API processing occurs in Europe; database/storage traffic crosses the public internet with TLS and allow-list controls because same-region Render Private Link is unavailable; added latency, egress, failure modes, and transfer analysis | **Technically possible, not production-approved**              |
| C. Africa-cohesive platform      | Replace Render compute with an approved platform in AWS Cape Town or another verified African location; colocate Atlas/storage and pin LiveKit to Africa | Can satisfy the technical Africa-region interpretation and reduce cross-region dependency traffic                              | Requires a replacement deployment ADR, new cost/operations analysis, Blueprint replacement, availability design, and legal approval; still not Ghana-only                                                                                  | **Preferred technical fallback if Africa-region is mandatory** |
| D. Ghana-only platform           | Select Ghana-hosted compute, database, storage, realtime, email/communications, backups, logs, and support boundaries                                    | Can target the strict interpretation                                                                                           | No complete provider set was verified in this spike; requires vendor diligence, resilience/DR design, supportability, and likely a platform change                                                                                         | **Blocked pending provider and legal research**                |

## Decision and guardrails

S0-011 does **not** approve Render for production member workloads. Render
remains feasible for local-adjacent development, synthetic preview, and
synthetic staging while the production residency posture is unresolved.

The production gate is:

1. Founder chooses whether "in-region" means Ghana-only or Africa-region.
2. Privacy/legal records the Act 843 transfer analysis and DPIA decision,
   including compute processing, databases, backups, object replicas/CDN
   caches, logs/observability, LiveKit media/egress, Resend email content, and
   provider support access.
3. Engineering validates latency from Ghana over representative 3G and fixed
   networks, including Render-to-Atlas round trips for Option B.
4. Procurement verifies provider DPAs, subprocessors, breach terms, deletion,
   retention, encryption/key ownership, backup locations, support access,
   service tiers, costs, and exit/export paths.
5. Architecture records the selected topology in a superseding or follow-up
   ADR before any production resource is created.

Until those gates pass:

- use synthetic data only on Render;
- do not send identity, biometric, private-circle, payment, safeguarding, or
  health-like content through Resend;
- do not enable LiveKit recording/egress for member sessions;
- do not claim Ghana-only or Africa-region compliance in product or operational
  material; and
- keep provider ports replaceable and do not bind domain code to provider SDKs.

## Required follow-up evidence

- Legal interpretation and signed DPIA decision for Ghana-only versus
  Africa-region.
- Measured Ghana client-to-compute and compute-to-database latency for Options A
  and B.
- Written confirmations for Atlas primary, backup, logs, support, and deletion
  locations on the intended tier.
- Storage/CDN vendor selection and full replica/cache/log/key location map.
- LiveKit confirmation covering network traffic, recordings, egress, analytics,
  support data, subprocessors, and outage behavior for Africa pinning.
- Resend approval or replacement, with a transactional-email data-minimization
  template and retention decision.
- Cost and reliability comparison for Render Frankfurt versus Africa-cohesive
  compute, followed by the production topology ADR.

## Primary legal source

- Ghana Data Protection Commission,
  [Data Protection Act, 2012 (Act 843)](https://dataprotection.org.gh/wp-content/uploads/2025/05/Data-Protection-Act-2012-Act-843.pdf),
  checked 2026-07-26.
