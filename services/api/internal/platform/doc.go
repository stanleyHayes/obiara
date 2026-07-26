// Package platform anchors the api service's service-specific platform
// packages (config, health, http, telemetry). It carries no code itself.
// Shared platform packages used by both api and worker (mongo, outbox,
// inbox, idempotency) live at the module root in internal/platform/.
package platform
