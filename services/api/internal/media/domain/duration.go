package domain

import (
	"mime"
	"strings"
	"time"
)

// bitrateBounds is the range of bits per second a codec is actually used at.
//
// Wide on purpose. This is not a quality judgement; it is the difference
// between a number that could have come from these bytes and one that could
// not.
var bitrateBounds = map[string][2]int64{
	"audio/ogg":  {6_000, 512_000},
	"audio/opus": {6_000, 512_000},
	"audio/webm": {6_000, 512_000},
	"audio/mpeg": {32_000, 320_000},
	"audio/mp4":  {8_000, 320_000},
	"audio/aac":  {8_000, 320_000},
	"audio/wav":  {64_000, 3_000_000},
}

// PlausibleDuration reports whether a claimed length could have come from this
// many bytes of this kind of audio.
//
// Nothing here can decode audio, so a recording's length is the one thing
// about it the client says rather than the server establishes. That claim is
// load-bearing — the twenty seconds that arm a sow are counted against it — so
// it is bounded rather than believed: a member cannot claim ninety seconds
// from forty kilobytes, because ninety seconds of Opus is not forty kilobytes.
//
// An unrecognised content type is refused. Admitting one would let a caller
// invent a type to escape the bound, which is the same wildcard mistake the
// access policy refuses to make about purposes.
func PlausibleDuration(contentType string, size int64, duration time.Duration) bool {
	if size <= 0 || duration <= 0 {
		return false
	}
	parsed, _, err := mime.ParseMediaType(strings.ToLower(strings.TrimSpace(contentType)))
	if err != nil {
		return false
	}
	bounds, known := bitrateBounds[parsed]
	if !known {
		return false
	}
	seconds := duration.Seconds()
	implied := int64(float64(size) * 8 / seconds)
	return implied >= bounds[0] && implied <= bounds[1]
}
