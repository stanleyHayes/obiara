package domain

import (
	"testing"
	"time"
)

func TestAClaimedLengthHasToFitTheBytes(t *testing.T) {
	// Ninety seconds of Opus is not forty kilobytes. The listen gate counts
	// against this number, so a member who could inflate it could arm a sow
	// having heard almost nothing.
	if PlausibleDuration("audio/ogg", 40_000, 90*time.Second) {
		t.Fatal("ninety seconds accepted from forty kilobytes")
	}
	// A minute of Opus at a normal voice bitrate is a few hundred kilobytes.
	if !PlausibleDuration("audio/ogg", 240_000, 60*time.Second) {
		t.Fatal("an ordinary recording was refused")
	}
}

func TestAClaimTooShortForTheBytesIsRefusedToo(t *testing.T) {
	// The other direction matters as well: claiming one second for a large
	// file would make a whole recording count as heard in one second.
	if PlausibleDuration("audio/ogg", 5_000_000, time.Second) {
		t.Fatal("one second accepted from five megabytes")
	}
}

func TestAnUnrecognisedTypeIsRefusedNotWaved(t *testing.T) {
	// Admitting one would let a caller invent a content type to escape the
	// bound — the same wildcard mistake the access policy refuses to make
	// about purposes.
	if PlausibleDuration("audio/invented", 240_000, 60*time.Second) {
		t.Fatal("an unknown type escaped the bound")
	}
	if PlausibleDuration("not a media type", 240_000, 60*time.Second) {
		t.Fatal("an unparseable type escaped the bound")
	}
}

func TestNothingIsPlausibleWithoutBytesOrTime(t *testing.T) {
	if PlausibleDuration("audio/ogg", 0, 60*time.Second) {
		t.Fatal("a zero-byte recording had a length")
	}
	if PlausibleDuration("audio/ogg", 240_000, 0) {
		t.Fatal("a zero-length recording was accepted")
	}
}

func TestParametersOnTheTypeDoNotDefeatTheBound(t *testing.T) {
	// "audio/ogg; codecs=opus" is what a browser actually sends.
	if !PlausibleDuration("audio/ogg; codecs=opus", 240_000, 60*time.Second) {
		t.Fatal("a browser's own content type was refused")
	}
}
