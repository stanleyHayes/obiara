package domain

import (
	"errors"
	"testing"
	"time"
)

func TestProfileRejectsSensitiveAndUnconsentedFields(t *testing.T) {
	for _, value := range []string{
		"write me at ama@example.com",
		"call +233 55 000 0123",
		"visit https://example.com/me",
	} {
		if _, err := NewField(value, VisibilityPrivate, "", 280, false); !errors.Is(err, ErrUnsafeProfile) {
			t.Fatalf("NewField(%q) error = %v, want %v", value, err, ErrUnsafeProfile)
		}
	}
	if _, err := NewField("Ama", VisibilityCommunity, "", 80, true); !errors.Is(err, ErrConsentRequired) {
		t.Fatalf("community field error = %v, want %v", err, ErrConsentRequired)
	}
	if _, err := NewField("Ama", VisibilityPrivate, "consent-1", 80, true); !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("private consent ref error = %v, want %v", err, ErrInvalidProfile)
	}
}

func TestProfileVisibilityIsFieldScoped(t *testing.T) {
	cases := []struct {
		visibility Visibility
		audience   Audience
		want       bool
	}{
		{VisibilityPrivate, AudienceSelf, true},
		{VisibilityPrivate, AudienceCircle, false},
		{VisibilityCircles, AudienceCircle, true},
		{VisibilityCircles, AudienceCommunity, false},
		{VisibilityCommunity, AudienceCommunity, true},
	}
	for _, test := range cases {
		field, err := NewField("Ama", test.visibility, consentFor(test.visibility), 80, true)
		if err != nil {
			t.Fatal(err)
		}
		got, err := field.VisibleTo(test.audience)
		if err != nil || got != test.want {
			t.Fatalf("%s visible to %s = %v, %v; want %v", test.visibility, test.audience, got, err, test.want)
		}
	}
}

func TestProfileOptimisticRevisionAndIdempotency(t *testing.T) {
	now := time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)
	display, _ := NewField("Ama", VisibilityCircles, "", 80, true)
	intro, _ := NewField("Here to build community.", VisibilityPrivate, "", 280, false)
	change := Change{CommandID: "cmd-1", DisplayName: display, Introduction: intro, RecordedAt: now}
	profile, err := Create("member-1", change)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := profile.Update(change)
	if err != nil || replay.Revision() != 1 {
		t.Fatalf("replay revision = %d, error = %v", replay.Revision(), err)
	}
	changedDisplay, _ := NewField("Akua", VisibilityCircles, "", 80, true)
	change.DisplayName = changedDisplay
	if _, err := profile.Update(change); !errors.Is(err, ErrCommandMismatch) {
		t.Fatalf("mismatched replay error = %v, want %v", err, ErrCommandMismatch)
	}
	change.CommandID = "cmd-2"
	change.ExpectedRevision = 0
	if _, err := profile.Update(change); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("stale update error = %v, want %v", err, ErrStaleRevision)
	}
}

func consentFor(visibility Visibility) string {
	if visibility == VisibilityCommunity {
		return "consent-community-1"
	}
	return ""
}
