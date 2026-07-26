// Package anomaly implements a stateless, offline-only anomaly review gate.
//
// It accepts bounded graph aggregates and opaque evidence references only.
// Rules, consent evidence, authority and human cases remain in their owning
// systems behind ports, so package-local Mongo persistence is inapplicable.
// Raw graph paths, device identifiers, member scores and enforcement decisions
// are intentionally absent from this boundary.
package anomaly
