package domain

import "regexp"

// identifierPatterns mask the identifier shapes that appear in member
// free text. Patterns are deliberately conservative — over-redaction is
// safe, under-redaction is a breach (Doc 09 §3).
var identifierPatterns = []*regexp.Regexp{
	// E.164 and local Ghana phone shapes.
	regexp.MustCompile(`(?:\+?\d[\d ()-]{7,}\d)`),
	// Email addresses.
	regexp.MustCompile(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`),
	// Social handles.
	regexp.MustCompile(`@[A-Za-z0-9_.]{2,}`),
}
