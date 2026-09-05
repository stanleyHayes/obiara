package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
	"go.uber.org/mock/gomock"
)

func TestSendScreensBeforeAtomicAcceptance(t *testing.T) {
	ctrl := gomock.NewController(t)
	screening := NewMockScreening(ctrl)
	acceptance := NewMockAcceptance(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	screening.EXPECT().Screen(gomock.Any(), "hello", []string{"raw-media"}).Return(ScreeningDecision{Approved: true, Reference: "raw-screen"}, nil)
	keyer.EXPECT().Key("allowance-subject", "raw-actor").Return("actor-key", nil)
	keyer.EXPECT().Key("screening", "raw-screen").Return("screen-key", nil)
	keyer.EXPECT().Key("media", "raw-media").Return("media-key", nil)
	ids.EXPECT().NewID().Return("sow-1")
	acceptance.EXPECT().Accept(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, s domain.Sow) (domain.Sow, bool, error) {
		if s.ActorKey != "actor-key" || s.Media[0].Key != "media-key" || s.AllowanceUnits != 1 {
			t.Fatalf("unsafe candidate %#v", s)
		}
		return s, false, nil
	})
	service := New(screening, acceptance, keyer, ids, func() time.Time { return now }, 1)
	result, err := service.Send(context.Background(), Command{ID: "command-1", ActorID: "raw-actor", Body: " hello ", MediaRefs: []string{"raw-media"}, Confirmed: true})
	if err != nil || result.Sow.ID != "sow-1" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestSendRejectsWithoutConfirmationOrScreening(t *testing.T) {
	ctrl := gomock.NewController(t)
	screening := NewMockScreening(ctrl)
	service := New(screening, NewMockAcceptance(ctrl), NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1)
	if _, err := service.Send(context.Background(), Command{ID: "c", ActorID: "a", Body: "body"}); !errors.Is(err, domain.ErrNotConfirmed) {
		t.Fatalf("got %v", err)
	}
	screening.EXPECT().Screen(gomock.Any(), "body", gomock.Any()).Return(ScreeningDecision{Approved: false}, nil)
	if _, err := service.Send(context.Background(), Command{ID: "c", ActorID: "a", Body: "body", Confirmed: true}); !errors.Is(err, domain.ErrScreeningRejected) {
		t.Fatalf("got %v", err)
	}
}

func FuzzFingerprintIsDeterministicAndInputBound(f *testing.F) {
	f.Add("c", "a", "body", "media")
	f.Fuzz(func(t *testing.T, c, a, b, m string) {
		one := fingerprint(c, a, b, []string{m}, 1)
		two := fingerprint(c, a, b, []string{m}, 1)
		if one != two || len(one) != 64 {
			t.Fatal("unstable fingerprint")
		}
		if m != "x" && one == fingerprint(c, a, b, []string{"x"}, 1) {
			t.Fatal("media not bound")
		}
	})
}

func TestASowSentToAPersonIsHeldRatherThanFailed(t *testing.T) {
	// Before this, a sow routed to a human came back as an error and the
	// member was told the service was unavailable — neither true nor
	// something they could act on. The seed is spent on the way in, because
	// a sow anyone could send for free is the point of the allowance.
	ctrl := gomock.NewController(t)
	screening := NewMockScreening(ctrl)
	acceptance := NewMockAcceptance(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)

	keyer.EXPECT().Key(gomock.Any(), gomock.Any()).DoAndReturn(
		func(namespace, value string) (string, error) { return namespace + ":" + value, nil }).AnyTimes()
	ids.EXPECT().NewID().Return("sow-1")
	screening.EXPECT().Screen(gomock.Any(), "body", gomock.Any()).
		Return(ScreeningDecision{Approved: false, Reference: "review-1"}, ErrHumanReviewRequired)
	acceptance.EXPECT().Accept(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s domain.Sow) (domain.Sow, bool, error) {
			if s.Status != domain.StatusPendingReview {
				t.Fatalf("status = %q, want pending review", s.Status)
			}
			if s.ScreeningRef != "review-1" {
				t.Fatalf("screening ref = %q, want the review's reference", s.ScreeningRef)
			}
			if s.AllowanceUnits <= 0 {
				t.Fatal("a held sow spent no seed")
			}
			return s, false, nil
		})

	service := New(screening, acceptance, keyer, ids, time.Now, 1)
	result, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", Body: "body", Confirmed: true,
	})
	if err != nil {
		t.Fatalf("a held sow returned an error: %v", err)
	}
	if result.Sow.Status != domain.StatusPendingReview {
		t.Fatalf("result status = %q", result.Sow.Status)
	}
}

func TestAReviewWithNoReferenceIsNotAHold(t *testing.T) {
	// The reference is how the held sow is found again. Without one it could
	// never be released or refused, so holding it would strand both the sow
	// and the member's seed.
	ctrl := gomock.NewController(t)
	screening := NewMockScreening(ctrl)
	acceptance := NewMockAcceptance(ctrl)
	screening.EXPECT().Screen(gomock.Any(), "body", gomock.Any()).
		Return(ScreeningDecision{Approved: false}, ErrHumanReviewRequired)
	// No Accept expectation: nothing may be stored.

	service := New(screening, acceptance, NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1)
	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", Body: "body", Confirmed: true,
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

// heldSow builds a sow that is waiting on a person.
func heldSow(t *testing.T) domain.Sow {
	t.Helper()
	sow, err := domain.Accept("sow-1", "actor-key", "body",
		[]domain.Media{{Key: "media-key", ScreeningKey: "screen-key"}},
		"command-1", "fingerprint", 1, domain.StatusPendingReview, "review-1", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return sow
}

func TestApprovingAReviewDeliversTheSowAndKeepsTheSeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	acceptance := NewMockAcceptance(ctrl)
	acceptance.EXPECT().FindByScreening(gomock.Any(), "review-1").Return(heldSow(t), nil)
	acceptance.EXPECT().Settle(gomock.Any(), gomock.Any(), false).DoAndReturn(
		func(_ context.Context, s domain.Sow, refund bool) error {
			if s.Status != domain.StatusDelivered {
				t.Fatalf("status = %q, want delivered", s.Status)
			}
			if s.ScreeningRef != "decision-1" {
				t.Fatalf("ref = %q, want the decision's reference", s.ScreeningRef)
			}
			return nil
		})

	service := New(NewMockScreening(ctrl), acceptance, NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1)
	if _, err := service.Review(context.Background(), "review-1", true, "decision-1"); err != nil {
		t.Fatal(err)
	}
}

func TestRefusingAReviewGivesTheSeedBack(t *testing.T) {
	// M4-ABUSE-01: the seed is refunded on failure. It is asked for in the
	// same call that stores the rejection, because a refusal that recorded
	// the outcome and lost the refund would take a member's seed for a sow
	// that was never delivered.
	ctrl := gomock.NewController(t)
	acceptance := NewMockAcceptance(ctrl)
	acceptance.EXPECT().FindByScreening(gomock.Any(), "review-1").Return(heldSow(t), nil)
	acceptance.EXPECT().Settle(gomock.Any(), gomock.Any(), true).DoAndReturn(
		func(_ context.Context, s domain.Sow, refund bool) error {
			if s.Status != domain.StatusRejected {
				t.Fatalf("status = %q, want rejected", s.Status)
			}
			if !refund {
				t.Fatal("a refused sow did not ask for the seed back")
			}
			return nil
		})

	service := New(NewMockScreening(ctrl), acceptance, NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1)
	if _, err := service.Review(context.Background(), "review-1", false, "decision-1"); err != nil {
		t.Fatal(err)
	}
}

func TestASowIsNotDecidedTwice(t *testing.T) {
	// Deciding twice would refund a seed twice. The aggregate refuses, and
	// nothing is written.
	ctrl := gomock.NewController(t)
	acceptance := NewMockAcceptance(ctrl)
	settled := heldSow(t)
	delivered, err := settled.Release("decision-1", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	acceptance.EXPECT().FindByScreening(gomock.Any(), "review-1").Return(delivered, nil)
	// No Settle expectation: a second decision must write nothing.

	service := New(NewMockScreening(ctrl), acceptance, NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1)
	if _, err := service.Review(context.Background(), "review-1", false, "decision-2"); !errors.Is(err, domain.ErrNotPending) {
		t.Fatalf("err = %v, want ErrNotPending", err)
	}
}
