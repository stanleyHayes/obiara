# API telemetry kernel

This package is the transport-neutral observability boundary for the API and
worker. It deliberately does not configure an exporter or mutate the global
OpenTelemetry provider. Composition roots own exporter lifecycle and bind the
`Metrics` port when the deployment topology is approved.

Privacy rules:

- Only bounded request, correlation, operation, trace and span identifiers
  belong in telemetry context.
- Never log raw member content, voice, identity artifacts, contact data,
  credentials, tokens, cookies or provider secrets.
- Attribute keys naming sensitive fields are deterministically replaced with
  `[REDACTED]`, including nested `slog.Group` attributes and attributes attached
  through `Logger.With`.
- Errors are not logged by `InstrumentHealth`; callers receive the original
  error while logs contain only dependency and status.
- Metric dimensions must be low-cardinality static vocabulary. Member,
  request, session and correlation identifiers are forbidden dimensions.

The HTTP layer can bridge its effective `X-Correlation-ID` into
`WithContextFields`; no dependency from telemetry to the transport package is
required.

## Runtime composition

The API and worker composition roots use vendor-neutral OTLP over HTTP. Export
is deliberately disabled when `OTEL_EXPORTER_OTLP_ENDPOINT` is empty. When
enabled, the endpoint must be an absolute credential-free `https` URL; local
plain HTTP additionally requires `OTEL_EXPORTER_OTLP_INSECURE=true`.

Deployment metadata is supplied through `SERVICE_VERSION` and `APP_ENV`.
Authentication belongs in the collector/network boundary; never place tokens
or credentials in the endpoint URL. Both processes flush trace and metric
providers during graceful shutdown.
