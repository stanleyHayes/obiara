// Package application scores scam-arc signals per room and drives the
// action ladder (E11-S11): education at the education rung, friction
// recording at friction, and a T&S case at the top rung. Monitoring is
// consent-map governed (Doc 08 §8: scam-arc room monitoring is opt-out).
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/domain"
)

var ErrMonitoringOptedOut = errors.New("room opted out of scam-arc monitoring")

// SignalStore is the append-only signal ledger with per-room state.
type SignalStore interface {
	Append(context.Context, domain.Signal) error
	KindsForRoom(context.Context, string) ([]domain.SignalKind, error)
	SaveState(context.Context, domain.RoomState) error
	StateForRoom(context.Context, string) (domain.RoomState, error)
}

// MonitoringConsent answers the scam-arc consent-map row (opt-out,
// default on; Doc 08 §8). A missing record means monitoring is on.
type MonitoringConsent interface {
	MonitoringAllowed(ctx context.Context, roomID string) (bool, error)
}

// CaseOpener creates a T&S case at the top rung (safety context port).
type CaseOpener interface {
	OpenScamCase(ctx context.Context, roomID, actorID string, score float64) error
}

// EducationCard is the in-context education surface for the at-risk party.
type EducationCard struct {
	RoomID     string
	ActorID    string
	ContentKey string
}

const educationContentKey = "education.scam_awareness"

// ScamArcService observes signals and recomputes ladder states.
type ScamArcService struct {
	signals SignalStore
	consent MonitoringConsent
	cases   CaseOpener
	now     func() time.Time
	newID   func() string
}

func NewScamArcService(signals SignalStore, consent MonitoringConsent, cases CaseOpener, now func() time.Time, newID func() string) ScamArcService {
	if consent == nil {
		consent = allowAllConsent{}
	}
	return ScamArcService{signals: signals, consent: consent, cases: cases, now: now, newID: newID}
}

// allowAllConsent is the default (Doc 08 §8: scam-arc monitoring is on
// unless a room opts out). The consent registry bridge replaces it.
type allowAllConsent struct{}

func (allowAllConsent) MonitoringAllowed(context.Context, string) (bool, error) { return true, nil }

// Observe records a signal and returns the room's new ladder state. When
// the education rung is crossed, an education card is produced; at the
// case rung a T&S case opens (idempotent by state: a room already at case
// produces nothing new).
func (service ScamArcService) Observe(ctx context.Context, roomID, actorID string, kind domain.SignalKind) (domain.RoomState, *EducationCard, error) {
	allowed, err := service.consent.MonitoringAllowed(ctx, roomID)
	if err != nil {
		return domain.RoomState{}, nil, err
	}
	if !allowed {
		return domain.RoomState{}, nil, ErrMonitoringOptedOut
	}

	signal := domain.Signal{ID: service.newID(), RoomID: roomID, ActorID: actorID, Kind: kind, ObservedAt: service.now().UTC()}
	if !signal.Valid() {
		return domain.RoomState{}, nil, errors.New("invalid scam-arc signal")
	}
	if err := service.signals.Append(ctx, signal); err != nil {
		return domain.RoomState{}, nil, err
	}

	kinds, err := service.signals.KindsForRoom(ctx, roomID)
	if err != nil {
		return domain.RoomState{}, nil, err
	}
	previous, _ := service.signals.StateForRoom(ctx, roomID)
	state := domain.Recompute(roomID, kinds, service.now())
	if err := service.signals.SaveState(ctx, state); err != nil {
		return domain.RoomState{}, nil, err
	}

	var card *EducationCard
	if state.Ladder == domain.LadderEducation && previous.Ladder != domain.LadderEducation {
		card = &EducationCard{RoomID: roomID, ActorID: actorID, ContentKey: educationContentKey}
	}
	if state.Ladder == domain.LadderCase && previous.Ladder != domain.LadderCase && service.cases != nil {
		if err := service.cases.OpenScamCase(ctx, roomID, actorID, state.Score); err != nil {
			return domain.RoomState{}, nil, err
		}
	}
	return state, card, nil
}

// StateForRoom reads the current ladder state.
func (service ScamArcService) StateForRoom(ctx context.Context, roomID string) (domain.RoomState, error) {
	return service.signals.StateForRoom(ctx, roomID)
}
