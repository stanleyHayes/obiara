package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxProfileLanguages   = 8
	MaxProfileSpecialties = 8
)

// LicensedProfile is the bounded public projection of a professional
// matchmaker. It intentionally contains no phone, email, member identity or
// private rating input.
type LicensedProfile struct {
	License              License
	DisplayName          string
	Languages            []string
	Specialties          []string
	CompletedEngagements uint64
	RatingBasisPoints    uint16
}

func (profile LicensedProfile) Valid(at time.Time) bool {
	if !profile.License.Current(at) ||
		!boundedLabel(profile.DisplayName, 2, 80) ||
		len(profile.Languages) == 0 || len(profile.Languages) > MaxProfileLanguages ||
		len(profile.Specialties) == 0 || len(profile.Specialties) > MaxProfileSpecialties ||
		profile.RatingBasisPoints > 500 {
		return false
	}
	for _, value := range append(append([]string(nil), profile.Languages...), profile.Specialties...) {
		if !boundedLabel(value, 2, 48) {
			return false
		}
	}
	return true
}

func (profile LicensedProfile) Clone() LicensedProfile {
	profile.Languages = append([]string(nil), profile.Languages...)
	profile.Specialties = append([]string(nil), profile.Specialties...)
	return profile
}

func boundedLabel(value string, minimum, maximum int) bool {
	value = strings.TrimSpace(value)
	size := utf8.RuneCountInString(value)
	return size >= minimum && size <= maximum
}
