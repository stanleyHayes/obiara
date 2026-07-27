# Launch-readiness aggregate projection

This read-side hexagon creates immutable, append-only review snapshots from
consented family counts and density, current host training/certification
coverage, and current matchmaker licensing for an exact jurisdiction.

Missing, incomplete, expired, insufficient or jurisdiction-mismatched evidence
fails closed. The persisted schema contains aggregate counts and versions only:
no family/member/contact list, outreach state, CRM record or source identifier.
There are no ports for recruitment, training, certification, licensing,
notification, waitlist or market activation mutation.
