package telemetry

import "io"

// LoggerOutput keeps runtime construction testable without coupling callers
// to a concrete file or network destination.
type LoggerOutput = io.Writer
