# P0 phase-exit projection

This projection implements the founder-approved Doc 04 gates using the Doc 08
event taxonomy: pods heard 65%, seed-to-sprout 25%, sprout-to-room 35%, weekly
fire attendance 40%, D30 retention 45%, regret trending down, and zero
unresolved Tier-A incidents.

Every denominator and completeness claim is explicit. Missing coverage or a
zero denominator produces `incomplete`, never a phase pass. Inputs are bounded
cohort aggregates with opaque window/snapshot/watermark references; there are
no source events, content, member identities, rankings, rollout controls, or
admin mutation.
