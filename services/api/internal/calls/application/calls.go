// Package application initiates and ends in-app calls (E09-S09).
// Membership in the courtship room is verified before any token is issued
// (FR-304: the call path never exposes phone numbers — join tokens carry
// opaque participant keys only).
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/calls/domain"
	livekitapp "github.com/stanleyHayes/obiara/services/api/internal/realtime/livekit/application"
)

var (
	ErrCallNotFound   = errors.New("call not found")
	ErrNotParticipant = errors.New("caller is not a call participant")
)

// CallRepository persists call sessions (metadata only, per FR-304
// retention posture).
type CallRepository interface {
	Create(context.Context, domain.Call) error
	FindByID(context.Context, string) (domain.Call, error)
	Update(context.Context, domain.Call) error
}

// RoomMembership verifies room membership (courtship context port).
type RoomMembership interface {
	IsMember(ctx context.Context, roomID, memberID string) (bool, error)
}

// TokenIssuer is the LiveKit token port (S7-001 adapter).
type TokenIssuer interface {
	Issue(context.Context, livekitapp.JoinRequest) (livekitapp.JoinToken, error)
}

// IssuedCall is the initiation result: the call plus one join token per
// participant (each gets only their own token).
type IssuedCall struct {
	Call   domain.Call
	Tokens map[string]livekitapp.JoinToken
}

// CallService initiates and ends calls.
type CallService struct {
	calls  CallRepository
	rooms  RoomMembership
	tokens TokenIssuer
	now    func() time.Time
	newID  func() string
}

func NewCallService(calls CallRepository, rooms RoomMembership, tokens TokenIssuer, now func() time.Time, newID func() string) CallService {
	return CallService{calls: calls, rooms: rooms, tokens: tokens, now: now, newID: newID}
}

// Initiate opens a call between two room members and issues a speaker
// token to each. Both must be room members (FR-304; E09-S09).
func (service CallService) Initiate(ctx context.Context, roomID, initiatorID, otherID string) (IssuedCall, error) {
	for _, memberID := range []string{initiatorID, otherID} {
		member, err := service.rooms.IsMember(ctx, roomID, memberID)
		if err != nil {
			return IssuedCall{}, err
		}
		if !member {
			return IssuedCall{}, domain.ErrNotRoomMember
		}
	}

	call, err := domain.NewCall(service.newID(), roomID, [2]string{initiatorID, otherID}, service.now())
	if err != nil {
		return IssuedCall{}, err
	}
	if err := service.calls.Create(ctx, call); err != nil {
		return IssuedCall{}, err
	}

	tokens := make(map[string]livekitapp.JoinToken, 2)
	for _, participantID := range call.Participants() {
		token, err := service.tokens.Issue(ctx, livekitapp.JoinRequest{
			RoomKey:        call.LiveKitRoomKey(),
			ParticipantKey: domain.OpaqueKey("participant", participantID),
			Role:           livekitapp.RoleSpeaker,
			ServerTime:     service.now().UTC(),
			TTL:            domain.CallTTL,
		})
		if err != nil {
			return IssuedCall{}, err
		}
		tokens[participantID] = token
	}
	return IssuedCall{Call: call, Tokens: tokens}, nil
}

// End closes a call; only a participant may end it.
func (service CallService) End(ctx context.Context, callID, actorID string) error {
	call, err := service.calls.FindByID(ctx, callID)
	if err != nil {
		return err
	}
	participants := call.Participants()
	if actorID != participants[0] && actorID != participants[1] {
		return ErrNotParticipant
	}
	if err := call.End(service.now()); err != nil {
		return err
	}
	return service.calls.Update(ctx, call)
}
