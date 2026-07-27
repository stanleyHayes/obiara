# Quarterly fairness, regret and safety projection

This boundary implements E15-S07 from Doc 08 section 5 and the P0 truth gates
in Doc 04. It consumes only privacy-thresholded quarterly aggregates. Cohorts
are opaque governance keys; protected-trait labels, member identifiers, source
events, content and free text are outside the port contract.

The versioned reviewed definition publishes its maximum exposure-rate gap in
permille. Reports preserve every numerator and denominator, treat colorism
drift as an audited failure, require regret rate to be strictly lower, and
require zero unresolved Tier-A cases. Missing evidence produces `incomplete`,
never `pass`. Reports are append-only and cannot change matching, rollout,
ranking, enforcement, safety cases, or source analytics.
