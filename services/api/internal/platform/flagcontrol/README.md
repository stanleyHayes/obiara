# Feature-flag control plane

This E16-S07 boundary wraps the existing fail-closed flags kernel with durable
administrative governance. A proposal names exactly one environment
(`staging` or `production`), market (`GH`), and canonical capability. There is
no global, member, or cohort scope.

Proposal terms are immutable and expire within two hours. The authority port
must resolve an active, least-privilege stepped-up admin session to an opaque
principal key. Approval requires a different stepped-up principal. Proposal
creation and every CAS transition commit with an append-only audit record in
one MongoDB transaction.

The runtime adapter is bound to one exact environment/market registry.
Expiration always publishes `enabled=false, killed=true`; kill-switch
precedence remains owned by the existing flags kernel. This package cannot
deploy software, target members, schedule rollout percentages, or bypass the
kernel.
