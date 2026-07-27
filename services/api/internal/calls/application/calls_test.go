package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/calls/domain"
	livekitapp "github.com/stanleyHayes/obiara/services/api/internal/realtime/livekit/application"
)

var callsNow = time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

func newService(t *testing.T) (CallService, *MockCallRepository, *MockRoomMembership, *MockTokenIssuer) {
	t.Helper()
	ctrl := gomock.NewController(t)
	calls := NewMockCallRepository(ctrl)
	rooms := NewMockRoomMembership(ctrl)
	tokens := NewMockTokenIssuer(ctrl)
	return NewCallService(calls, rooms, tokens, func() time.Time { return callsNow }, func() string { return "call_test" }), calls, rooms, tokens
}

func TestInitiateRequiresMembership(t *testing.T) {
	service, _, rooms, _ := newService(t)
	rooms.EXPECT().IsMember(gomock.Any(), "room_1", "m-1").Return(true, nil)
	rooms.EXPECT().IsMember(gomock.Any(), "room_1", "m-2").Return(false, nil)

	if _, err := service.Initiate(context.Background(), "room_1", "m-1", "m-2"); err != domain.ErrNotRoomMember {
		t.Fatalf("non-member = %v, want ErrNotRoomMember", err)
	}
}

func TestInitiateIssuesSpeakerTokenPerParticipant(t *testing.T) {
	service, calls, rooms, tokens := newService(t)
	rooms.EXPECT().IsMember(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
	calls.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	issued := map[string]bool{}
	tokens.EXPECT().Issue(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, request livekitapp.JoinRequest) (livekitapp.JoinToken, error) {
			if request.Role != livekitapp.RoleSpeaker || len(request.RoomKey) != 64 {
				t.Fatalf("request = %#v", request)
			}
			if request.TTL != domain.CallTTL {
				t.Fatalf("ttl = %v", request.TTL)
			}
			issued[request.ParticipantKey] = true
			if len(request.ParticipantKey) != 64 {
				t.Fatalf("participant key must be 64-hex opaque: %q", request.ParticipantKey)
			}
			return livekitapp.JoinToken{Signed: "tok_" + request.ParticipantKey, ExpiresAt: callsNow.Add(domain.CallTTL)}, nil
		}).Times(2)

	result, err := service.Initiate(context.Background(), "room_1", "m-1", "m-2")
	if err != nil {
		t.Fatal(err)
	}
	if !issued[domain.OpaqueKey("participant", "m-1")] || !issued[domain.OpaqueKey("participant", "m-2")] {
		t.Fatalf("tokens = %#v", result.Tokens)
	}
	if result.Tokens["m-1"].Signed == result.Tokens["m-2"].Signed {
		t.Fatal("each participant gets their own token")
	}
}

func TestEndParticipantOnly(t *testing.T) {
	service, calls, _, _ := newService(t)
	call := domain.ReconstituteCall("call_1", "room_1", [2]string{"m-1", "m-2"}, domain.StatusRinging, 1, callsNow, nil)
	calls.EXPECT().FindByID(gomock.Any(), "call_1").Return(call, nil)

	if err := service.End(context.Background(), "call_1", "m-3"); err != ErrNotParticipant {
		t.Fatalf("outsider end = %v, want ErrNotParticipant", err)
	}

	calls.EXPECT().FindByID(gomock.Any(), "call_1").Return(call, nil)
	calls.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, updated domain.Call) error {
			if updated.Status() != domain.StatusEnded {
				t.Fatalf("call = %#v", updated)
			}
			return nil
		})
	if err := service.End(context.Background(), "call_1", "m-2"); err != nil {
		t.Fatal(err)
	}
}
