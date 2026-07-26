# Scam-arc sequence boundary

This package evaluates bounded, opaque event summaries against reviewed,
versioned rules and routes neutral signals to a human. It contains no raw
message, voice, payment, member identity, free text, accusation, hidden
member score, model/vendor client, or enforcement port.

Recommendations form a least-harm ladder (`observe_only`, `education`,
`friction`, `human_case`). They are recommendations for human review, never
automatic block, payment, account, trust, or matching actions. A single event
cannot produce a signal.

MongoDB and Testcontainers are intentionally inapplicable here: the evaluator
is stateless and reads a current bounded sequence through `EventSource`.
Upstream event-summary ownership and downstream human-case persistence remain
outside this slice.
