# `@obiara/observability`

Cross-client privacy boundary for web, admin and Expo telemetry. It provides
context normalization, recursive sensitive-field redaction, event construction,
and exporter-neutral event/metric sink interfaces.

Rules:

- Events are an allowlisted operational vocabulary, never prose or member
  content.
- Do not include raw conversation text, voice, contact information, identity
  artifacts, credentials, cookies, tokens, or provider secrets.
- Use low-cardinality metric dimensions only. Member, session, request and
  correlation identifiers are forbidden metric dimensions.
- Product analytics consent must be checked by the calling feature before
  emitting analytics. Service-reliability telemetry remains operational and
  must still obey this package's data minimization rules.
- Exporters belong in app composition roots so browser, React Native and server
  runtimes can select appropriate transports without this package importing a
  vendor SDK.
