package domain

import (
	"testing"
	"time"
)

func TestLicensedProfileRequiresCurrentBoundedPublicData(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	profile := LicensedProfile{
		License: License{
			ID: "license.ghana", MatchmakerKey: key(2), Jurisdiction: "ghana",
			Version: 1, ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour),
			MinimumFeePesewas: 8000, MaximumFeePesewas: 25000,
		},
		DisplayName: "Akosua Mensah", Languages: []string{"Twi", "English"},
		Specialties: []string{"Consultation"}, RatingBasisPoints: 475,
	}
	if !profile.Valid(now) {
		t.Fatal("current bounded profile must be valid")
	}
	profile.License.ValidUntil = now
	if profile.Valid(now) {
		t.Fatal("expired profile must fail closed")
	}
}
