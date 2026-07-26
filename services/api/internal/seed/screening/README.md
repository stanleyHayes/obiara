# Sow screening boundary

This package is a stateless adapter behind
`seed/sow/application.Screening`. It normalizes transient text, resolves a
reviewed locale, inspects bounded media metadata, collects provider-neutral
advisory evidence, and requires a separate human adjudicator for final
approval or rejection.

Unsupported locales and media, uncertain evidence, invalid adjudication, and
provider failures are routed to human review and fail closed to the caller.
The adapter owns no delivery, allowance, acceptance, persistence, raw-content
store, translation, model, or vendor client.

MongoDB and Testcontainers are intentionally inapplicable: this slice has no
material state. The existing sow acceptance transaction remains the sole
owner of persistence and allowance spend after an approved screening result.

