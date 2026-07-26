// Package platform anchors the shared Go platform packages (mongo, outbox,
// inbox, idempotency) used by every service in the module (api, worker).
// It carries no code itself; see the subpackages and their READMEs for
// conventions. Service-specific platform packages (config, health, http,
// telemetry) remain under services/api/internal/platform/.
package platform
