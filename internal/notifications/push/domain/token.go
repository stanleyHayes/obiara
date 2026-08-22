// Package domain models device push registration and the bounded message
// shapes the push channel may send (E13-S03 channel ladder).
//
// FR-701 applies here exactly as it does to WhatsApp: a push notification
// leaves the device's lock screen visible to anyone holding the phone, so it
// carries a template and an opaque reference, never room content, never a
// counterpart's name.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrTokenInvalid    = errors.New("push token must be an Expo push token")
	ErrMemberRequired  = errors.New("push registration requires a member")
	ErrPlatformInvalid = errors.New("push platform must be ios, android or web")
)

// Platform is the device family a token belongs to.
type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
	PlatformWeb     Platform = "web"
)

// expoToken matches the two documented Expo push token shapes. Validating at
// the boundary keeps a malformed token from reaching the provider, where a
// single bad entry can reject a whole batch.
var expoToken = regexp.MustCompile(`^Expo(nent)?PushToken\[[A-Za-z0-9._%+-]{1,128}\]$`)

// Registration is one device's push registration.
type Registration struct {
	memberID   string
	token      string
	platform   Platform
	registered time.Time
}

// NewRegistration validates a device registration.
func NewRegistration(memberID, token string, platform Platform, now time.Time) (Registration, error) {
	memberID = strings.TrimSpace(memberID)
	token = strings.TrimSpace(token)
	if memberID == "" {
		return Registration{}, ErrMemberRequired
	}
	if !expoToken.MatchString(token) {
		return Registration{}, ErrTokenInvalid
	}
	switch platform {
	case PlatformIOS, PlatformAndroid, PlatformWeb:
	default:
		return Registration{}, ErrPlatformInvalid
	}
	return Registration{
		memberID: memberID, token: token, platform: platform, registered: now.UTC(),
	}, nil
}

// Reconstitute rebuilds a stored registration without policy checks.
func Reconstitute(memberID, token string, platform Platform, registered time.Time) Registration {
	return Registration{memberID: memberID, token: token, platform: platform, registered: registered}
}

func (r Registration) MemberID() string      { return r.memberID }
func (r Registration) Token() string         { return r.token }
func (r Registration) Platform() Platform    { return r.platform }
func (r Registration) Registered() time.Time { return r.registered }

// ValidToken reports whether a token is well formed, for callers pruning a
// stored set before a send.
func ValidToken(token string) bool {
	return expoToken.MatchString(strings.TrimSpace(token))
}
