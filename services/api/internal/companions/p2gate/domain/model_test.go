package domain

import (
	"errors"
	"testing"
	"testing/quick"
	"time"
)

func validConsent() GateConsent {
	return GateConsent{
		CourtshipRef: "courtship-001", PackVersion: 4,
		ConsentedItems: []PackItem{IdentityCard, VoiceIntroduction},
		PartyAApproved: true, PartyBApproved: true, Current: true,
	}
}

func TestProposalIsBoundedOTPProtectedAndExpires(t *testing.T) {
	now := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	p, err := Propose("proposal-001", "command-0001", "courtship-001", "reviewer-001",
		"tokenref-001", "watermark-001", 4, []PackItem{VoiceIntroduction, IdentityCard}, validConsent(), now)
	if err != nil {
		t.Fatal(err)
	}
	if !p.OTPRequired || !p.NoForward || !p.DeliveryProposed || !p.ExpiresAt.Equal(now.Add(ReviewWindow)) {
		t.Fatalf("unsafe proposal: %+v", p)
	}
	if p.Items[0] != IdentityCard || p.Items[1] != VoiceIntroduction {
		t.Fatalf("items not canonical: %v", p.Items)
	}
}

func TestProposalRequiresExactCurrentBilateralConsent(t *testing.T) {
	now := time.Now().UTC()
	cases := []GateConsent{
		func() GateConsent { c := validConsent(); c.Current = false; return c }(),
		func() GateConsent { c := validConsent(); c.PartyAApproved = false; return c }(),
		func() GateConsent { c := validConsent(); c.PartyBApproved = false; return c }(),
		func() GateConsent { c := validConsent(); c.PackVersion++; return c }(),
	}
	for _, consent := range cases {
		_, err := Propose("proposal-001", "command-0001", "courtship-001", "reviewer-001",
			"tokenref-001", "watermark-001", 4, []PackItem{IdentityCard}, consent, now)
		if !errors.Is(err, ErrConsentRequired) {
			t.Fatalf("expected consent failure for %+v, got %v", consent, err)
		}
	}
}

func TestUSSDViewCopiesAndBoundsFacts(t *testing.T) {
	now := time.Now().UTC()
	facts := CompanionFacts{
		MemberRef: "member-0001", PodCount: 2, DrumWaiting: true,
		UpcomingFire: []FireSlot{{ScheduleRef: "schedule-002", StartsAt: now.Add(2 * time.Hour)}, {ScheduleRef: "schedule-001", StartsAt: now.Add(time.Hour)}},
		HelpRefs:     []string{"help-safety-001", "help-support-001"},
	}
	view, err := NewUSSDView(facts, now)
	if err != nil {
		t.Fatal(err)
	}
	facts.HelpRefs[0] = "mutated"
	if view.HelpRefs[0] != "help-safety-001" || view.UpcomingFire[0].ScheduleRef != "schedule-001" {
		t.Fatalf("view is mutable or unsorted: %+v", view)
	}
}

func TestUnconsentedItemsAreNeverAcceptedProperty(t *testing.T) {
	property := func(raw uint8) bool {
		item := PackItem(raw)
		if allowedPack[item] {
			return true
		}
		_, err := Propose("proposal-001", "command-0001", "courtship-001", "reviewer-001",
			"tokenref-001", "watermark-001", 4, []PackItem{item}, validConsent(), time.Now().UTC())
		return err != nil
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 1000}); err != nil {
		t.Fatal(err)
	}
}

func FuzzProposalPackConsent(f *testing.F) {
	f.Add("identity-card")
	f.Add("room-content")
	f.Fuzz(func(t *testing.T, item string) {
		_, err := Propose("proposal-001", "command-0001", "courtship-001", "reviewer-001",
			"tokenref-001", "watermark-001", 4, []PackItem{PackItem(item)}, validConsent(), time.Unix(1, 0))
		if err == nil && PackItem(item) != IdentityCard && PackItem(item) != VoiceIntroduction {
			t.Fatalf("accepted item outside consent: %q", item)
		}
	})
}
