// Package sikashield defines a stateless, offline-only safety evaluation gate.
//
// Pattern definitions, aggregate evaluation metrics, consent evidence, and human
// cases remain owned by their source-of-truth ports. This package deliberately
// persists no raw text, raw voice, derived member profile, score, accusation, or
// enforcement decision; consequently a package-local Mongo adapter is not
// applicable. Only opaque evidence references cross the boundary.
package sikashield
