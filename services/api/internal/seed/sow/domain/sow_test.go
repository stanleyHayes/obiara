package domain

import (
	"testing"
	"time"
)

func TestAcceptCopiesMediaAndRejectsInvalidUnits(t *testing.T) {
	media := []Media{{Key: "media-key", ScreeningKey: "screen-key"}}
	sow, err := Accept("id", "actor", "body", media, "command", "fingerprint", 1, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	media[0].Key = "mutated"
	if sow.Media[0].Key != "media-key" {
		t.Fatal("media aliases caller memory")
	}
	if _, err = Accept("id", "actor", "body", nil, "command", "fingerprint", 0, time.Now()); err == nil {
		t.Fatal("zero-unit sow accepted")
	}
}
