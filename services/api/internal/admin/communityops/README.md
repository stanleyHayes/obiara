# Community operations proposals

This bounded admin hexagon stores only opaque circle, fire, host, actor and evidence
keys plus density counts. It revalidates current density, host verification,
training/certification and the exact participant-notice preview before an
acknowledgement can make a proposal ready for human review.

There is deliberately no circle, fire, host or notification mutation port. A
proposal cannot apply an operation, assign a host, bypass capacity, send a
notice, or approve itself. Every accepted command appends an immutable audit
entry and persistence uses command uniqueness plus optimistic concurrency.
