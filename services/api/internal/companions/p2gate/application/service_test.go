package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/companions/p2gate/domain"
	"go.uber.org/mock/gomock"
)

type fixedClock struct{ at time.Time }

func (c fixedClock) Now() time.Time { return c.at }

func TestProposeGateLinkRevalidatesConsentAndOnlyPersistsProposal(t *testing.T) {
	ctrl := gomock.NewController(t)
	consent, facts := NewMockConsentSource(ctrl), NewMockCompanionSource(ctrl)
	auth, repo, ids := NewMockSessionAuthenticator(ctrl), NewMockRepository(ctrl), NewMockIDSource(ctrl)
	now := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	cmd := ProposeCommand{
		CommandID: "command-0001", ActorRef: "member-00001", CourtshipRef: "courtship-001",
		ReviewerRef: "reviewer-001", PackVersion: 4, Items: []domain.PackItem{domain.IdentityCard},
	}
	gomock.InOrder(
		auth.EXPECT().Authenticate(gomock.Any(), cmd.ActorRef, cmd.CourtshipRef).Return(nil),
		consent.EXPECT().CurrentGateConsent(gomock.Any(), cmd.CourtshipRef).Return(domain.GateConsent{
			CourtshipRef: cmd.CourtshipRef, PackVersion: 4, ConsentedItems: cmd.Items,
			PartyAApproved: true, PartyBApproved: true, Current: true,
		}, nil),
		ids.EXPECT().NewID().Return("proposal-001"),
		ids.EXPECT().NewTokenRef().Return("tokenref-001"),
		ids.EXPECT().NewWatermarkRef().Return("watermark-001"),
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, p domain.Proposal) error {
			if p.ReviewerRef != cmd.ReviewerRef || !p.OTPRequired || !p.NoForward {
				t.Fatalf("unsafe proposal: %+v", p)
			}
			return nil
		}),
	)
	got, err := New(consent, facts, auth, repo, ids, fixedClock{now}).ProposeGateLink(context.Background(), cmd)
	if err != nil || !got.DeliveryProposed {
		t.Fatalf("proposal=%+v err=%v", got, err)
	}
}

func TestProposeGateLinkFailsBeforeIDsAndPersistenceWithoutAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	auth := NewMockSessionAuthenticator(ctrl)
	auth.EXPECT().Authenticate(gomock.Any(), "member-00001", "courtship-001").Return(errors.New("denied"))
	service := New(NewMockConsentSource(ctrl), NewMockCompanionSource(ctrl), auth, NewMockRepository(ctrl), NewMockIDSource(ctrl), fixedClock{time.Now()})
	_, err := service.ProposeGateLink(context.Background(), ProposeCommand{ActorRef: "member-00001", CourtshipRef: "courtship-001"})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
}

func TestViewUSSDIsAuthenticatedBoundedAndReadOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	facts, auth := NewMockCompanionSource(ctrl), NewMockSessionAuthenticator(ctrl)
	now := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	auth.EXPECT().Authenticate(gomock.Any(), "session-0001", "member-00001").Return(nil)
	facts.EXPECT().CurrentCompanionFacts(gomock.Any(), "member-00001").Return(domain.CompanionFacts{
		MemberRef: "member-00001", PodCount: 3, DrumWaiting: true,
		UpcomingFire: []domain.FireSlot{{ScheduleRef: "fire-slot-001", StartsAt: now.Add(time.Hour)}},
		HelpRefs:     []string{"help-safety-001"},
	}, nil)
	view, err := New(NewMockConsentSource(ctrl), facts, auth, NewMockRepository(ctrl), NewMockIDSource(ctrl), fixedClock{now}).
		ViewUSSD(context.Background(), "session-0001", "member-00001")
	if err != nil || view.PodCount != 3 || !view.DrumWaiting {
		t.Fatalf("view=%+v err=%v", view, err)
	}
}
