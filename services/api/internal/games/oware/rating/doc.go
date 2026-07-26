// Package rating implements the pure E10-S03 Oware rating and notation
// kernel. Ratings use the Glicko-2 rating-period algorithm. Notation replays
// every recorded ply through the existing Oware legality engine.
//
// The package has no persistence, transport, member identity, matching,
// discovery, or popularity concerns.
package rating
