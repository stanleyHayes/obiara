// Package domain defines the provider-neutral P2 companion contract.
package domain

import (
	"errors"
	"regexp"
	"slices"
	"time"
)

const (
	ReviewWindow = 30 * 24 * time.Hour
	MaxPodCount  = 99
	MaxFires     = 3
	MaxHelpRefs  = 3
)

var (
	ErrInvalidProposal = errors.New("invalid Gate link proposal")
	ErrConsentRequired = errors.New("exact bilateral Gate consent required")
	ErrInvalidUSSDView = errors.New("invalid USSD companion view")
	opaque             = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:-]{7,127}$`)
)

type PackItem string

const (
	IdentityCard      PackItem = "identity-card"
	SubanMarks        PackItem = "suban-marks"
	VoiceIntroduction PackItem = "voice-introduction"
	CourtshipSummary  PackItem = "courtship-summary"
)

var allowedPack = map[PackItem]bool{
	IdentityCard: true, SubanMarks: true, VoiceIntroduction: true, CourtshipSummary: true,
}

type GateConsent struct {
	CourtshipRef   string
	PackVersion    uint64
	ConsentedItems []PackItem
	PartyAApproved bool
	PartyBApproved bool
	Current        bool
}

type Proposal struct {
	ID               string     `bson:"id"`
	CommandID        string     `bson:"commandId"`
	CourtshipRef     string     `bson:"courtshipRef"`
	ReviewerRef      string     `bson:"reviewerRef"`
	PackVersion      uint64     `bson:"packVersion"`
	Items            []PackItem `bson:"items"`
	TokenRef         string     `bson:"tokenRef"`
	WatermarkRef     string     `bson:"watermarkRef"`
	OTPRequired      bool       `bson:"otpRequired"`
	NoForward        bool       `bson:"noForward"`
	CreatedAt        time.Time  `bson:"createdAt"`
	ExpiresAt        time.Time  `bson:"expiresAt"`
	DeliveryProposed bool       `bson:"deliveryProposed"`
}

func Propose(id, commandID, courtshipRef, reviewerRef, tokenRef, watermarkRef string, packVersion uint64, requested []PackItem, consent GateConsent, now time.Time) (Proposal, error) {
	if !opaque.MatchString(id) || !opaque.MatchString(commandID) || !opaque.MatchString(courtshipRef) ||
		!opaque.MatchString(reviewerRef) || !opaque.MatchString(tokenRef) || !opaque.MatchString(watermarkRef) ||
		packVersion == 0 || now.IsZero() || len(requested) == 0 || len(requested) > len(allowedPack) {
		return Proposal{}, ErrInvalidProposal
	}
	if !consent.Current || !consent.PartyAApproved || !consent.PartyBApproved ||
		consent.CourtshipRef != courtshipRef || consent.PackVersion != packVersion {
		return Proposal{}, ErrConsentRequired
	}
	seen := make(map[PackItem]bool, len(requested))
	items := append([]PackItem(nil), requested...)
	for _, item := range items {
		if !allowedPack[item] || seen[item] || !slices.Contains(consent.ConsentedItems, item) {
			return Proposal{}, ErrConsentRequired
		}
		seen[item] = true
	}
	slices.Sort(items)
	return Proposal{
		ID: id, CommandID: commandID, CourtshipRef: courtshipRef, ReviewerRef: reviewerRef,
		PackVersion: packVersion, Items: items, TokenRef: tokenRef, WatermarkRef: watermarkRef,
		OTPRequired: true, NoForward: true, CreatedAt: now.UTC(), ExpiresAt: now.UTC().Add(ReviewWindow),
		DeliveryProposed: true,
	}, nil
}

type FireSlot struct {
	ScheduleRef string
	StartsAt    time.Time
}

type CompanionFacts struct {
	MemberRef    string
	PodCount     uint8
	DrumWaiting  bool
	UpcomingFire []FireSlot
	HelpRefs     []string
}

type USSDView struct {
	PodCount     uint8
	DrumWaiting  bool
	UpcomingFire []FireSlot
	HelpRefs     []string
}

func NewUSSDView(f CompanionFacts, now time.Time) (USSDView, error) {
	if !opaque.MatchString(f.MemberRef) || f.PodCount > MaxPodCount ||
		len(f.UpcomingFire) > MaxFires || len(f.HelpRefs) == 0 || len(f.HelpRefs) > MaxHelpRefs {
		return USSDView{}, ErrInvalidUSSDView
	}
	fires := append([]FireSlot(nil), f.UpcomingFire...)
	for _, fire := range fires {
		if !opaque.MatchString(fire.ScheduleRef) || !fire.StartsAt.After(now) {
			return USSDView{}, ErrInvalidUSSDView
		}
	}
	help := append([]string(nil), f.HelpRefs...)
	for _, ref := range help {
		if !opaque.MatchString(ref) {
			return USSDView{}, ErrInvalidUSSDView
		}
	}
	slices.SortFunc(fires, func(a, b FireSlot) int { return a.StartsAt.Compare(b.StartsAt) })
	slices.Sort(help)
	return USSDView{PodCount: f.PodCount, DrumWaiting: f.DrumWaiting, UpcomingFire: fires, HelpRefs: help}, nil
}
