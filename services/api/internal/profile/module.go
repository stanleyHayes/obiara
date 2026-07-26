// Package profile contains the transport-neutral profile bounded context.
//
// The package intentionally does not expose HTTP handlers or import identity,
// consent, privacy, or authorization implementations. Composition supplies
// outbound ports after those boundaries have made their own access decisions.
package profile
