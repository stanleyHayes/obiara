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
	keyer.EXPECT().Key("participant", "raw-target").Return("target-key", nil)
	keyer.EXPECT().Key("screening", "raw-screen").Return("screen-key", nil)
	keyer.EXPECT().Key("media", "raw-media").Return("media-key", nil)
	ids.EXPECT().NewID().Return("sow-1")
	acceptance.EXPECT().Accept(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, s domain.Sow) (domain.Sow, bool, error) {
		if s.ActorKey != "actor-key" || s.Media[0].Key != "media-key" || s.AllowanceUnits != 1 {
			t.Fatalf("unsafe candidate %#v", s)
		}
		return s, false, nil
	})
	service := New(screening, acceptance, keyer, ids, func() time.Time { return now }, 1).
		WithMediaOwnership(ownedMedia{owned: true}).
		WithReachRules(openReach(), openReach(), openReach()).
		WithDelivery(&placements{})
	result, err := service.Send(context.Background(), Command{ID: "command-1", ActorID: "raw-actor", TargetID: "raw-target", Body: " hello ", MediaRefs: []string{"raw-media"}, Confirmed: true})
	if err != nil || result.Sow.ID != "sow-1" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestSendRejectsWithoutConfirmationOrScreening(t *testing.T) {
	ctrl := gomock.NewController(t)
	screening := NewMockScreening(ctrl)
	service := New(screening, NewMockAcceptance(ctrl), NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1).
		WithMediaOwnership(ownedMedia{owned: true}).
		WithReachRules(openReach(), openReach(), openReach()).
		WithDelivery(&placements{})
	if _, err := service.Send(context.Background(), Command{ID: "c", ActorID: "a", TargetID: "t", Body: "body", MediaRefs: []string{"mine"}}); !errors.Is(err, domain.ErrNotConfirmed) {
		t.Fatalf("got %v", err)
	}
	screening.EXPECT().Screen(gomock.Any(), "body", gomock.Any()).Return(ScreeningDecision{Approved: false}, nil)
	if _, err := service.Send(context.Background(), Command{ID: "c", ActorID: "a", TargetID: "t", Body: "body", MediaRefs: []string{"mine"}, Confirmed: true}); !errors.Is(err, domain.ErrScreeningRejected) {
		t.Fatalf("got %v", err)
	}
}

func FuzzFingerprintIsDeterministicAndInputBound(f *testing.F) {
	f.Add("c", "a", "body", "media")
	f.Fuzz(func(t *testing.T, c, a, b, m string) {
		one := fingerprint(c, a, "target", b, []string{m}, 1)
		two := fingerprint(c, a, "target", b, []string{m}, 1)
		if one != two || len(one) != 64 {
			t.Fatal("unstable fingerprint")
		}
		if m != "x" && one == fingerprint(c, a, "target", b, []string{"x"}, 1) {
			t.Fatal("media not bound")
		}
		// The same command id toward a different person is a different sow,
		// not a replay of the first one.
		if one == fingerprint(c, a, "somebody-else", b, []string{m}, 1) {
			t.Fatal("target not bound")
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

	service := New(screening, acceptance, keyer, ids, time.Now, 1).
		WithMediaOwnership(ownedMedia{owned: true}).
		WithReachRules(openReach(), openReach(), openReach()).
		WithDelivery(&placements{})
	result, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body", MediaRefs: []string{"mine"}, Confirmed: true,
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

	service := New(screening, acceptance, NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1).
		WithMediaOwnership(ownedMedia{owned: true}).
		WithReachRules(openReach(), openReach(), openReach()).
		WithDelivery(&placements{})
	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body", MediaRefs: []string{"mine"}, Confirmed: true,
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

// heldSow builds a sow that is waiting on a person.
func heldSow(t *testing.T) domain.Sow {
	t.Helper()
	sow, err := domain.Accept("sow-1", "actor-key", "target-key", "body",
		[]domain.Media{{Key: "media-key", ScreeningKey: "screen-key"}},
		"command-1", "fingerprint", 1, domain.StatusPendingReview, "review-1",
		domain.Delivery{SowerID: "sower", TargetID: "target", MediaRefs: []string{"recording"}},
		time.Now())
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

	service := New(NewMockScreening(ctrl), acceptance, NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1).
		WithDelivery(&placements{})
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

// ownedMedia is a fixed answer about whose recordings these are.
type ownedMedia struct {
	owned bool
	err   error
}

func (m ownedMedia) OwnedBy(context.Context, string, []string) (bool, error) { return m.owned, m.err }

func TestASowMayOnlyCarryTheSowersOwnVoice(t *testing.T) {
	// People meet through their voices here, so sending somebody else's as
	// your own is impersonation. Nothing checked this before.
	ctrl := gomock.NewController(t)
	screening := NewMockScreening(ctrl)
	acceptance := NewMockAcceptance(ctrl)
	// No Screen and no Accept expectations: a sow carrying a voice that is
	// not the sower's must not even be screened, let alone stored.

	service := New(screening, acceptance, NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1).
		WithMediaOwnership(ownedMedia{owned: false}).WithReachRules(openReach(), openReach(), openReach())

	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body", MediaRefs: []string{"someone-elses"}, Confirmed: true,
	}); !errors.Is(err, ErrMediaNotOwned) {
		t.Fatalf("err = %v, want ErrMediaNotOwned", err)
	}
}

func TestAnUnansweredOwnershipCheckRefuses(t *testing.T) {
	// The direction that matters: if we cannot tell whose voice this is, the
	// sow does not go. Guessing yes is how an impersonation gets through.
	ctrl := gomock.NewController(t)
	service := New(NewMockScreening(ctrl), NewMockAcceptance(ctrl), NewMockKeyer(ctrl),
		NewMockIDSource(ctrl), time.Now, 1).
		WithMediaOwnership(ownedMedia{err: errors.New("media unavailable")}).
		WithReachRules(openReach(), openReach(), openReach())

	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body", MediaRefs: []string{"ref"}, Confirmed: true,
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}

	// And a service composed without the check at all refuses too, rather
	// than treating a missing check as permission.
	bare := New(NewMockScreening(ctrl), NewMockAcceptance(ctrl), NewMockKeyer(ctrl),
		NewMockIDSource(ctrl), time.Now, 1).
		WithReachRules(openReach(), openReach(), openReach())
	if _, err := bare.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body", MediaRefs: []string{"ref"}, Confirmed: true,
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestASowWithNothingToSayIsRefusedAtTheDoor(t *testing.T) {
	// A pod is a recording resting at somebody's house front. A sow with no
	// recording has nothing to place, so it would be accepted, charged a
	// seed, marked delivered and never arrive. Refused before any of that:
	// screening and acceptance carry no expectations here.
	ctrl := gomock.NewController(t)
	screening := NewMockScreening(ctrl)
	acceptance := NewMockAcceptance(ctrl)

	service := New(screening, acceptance, NewMockKeyer(ctrl), NewMockIDSource(ctrl), time.Now, 1).
		WithReachRules(openReach(), openReach(), openReach())
	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body", Confirmed: true,
	}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
}

// reach is a fixed set of answers to the three questions every reach toward a
// person has to pass. Each answer has its own error so a test can fail one
// rule without failing the others, which is what makes the ordering visible.
type reach struct {
	blocked, heard, locked            bool
	blockErr, heardErr, lockedErr     error
	blockCalls, heardCalls, lockCalls *int
}

func (r reach) Blocked(context.Context, string, string) (bool, error) {
	if r.blockCalls != nil {
		*r.blockCalls++
	}
	return r.blocked, r.blockErr
}
func (r reach) Heard(context.Context, string, string) (bool, error) {
	if r.heardCalls != nil {
		*r.heardCalls++
	}
	return r.heard, r.heardErr
}
func (r reach) Locked(context.Context, string, string) (bool, error) {
	if r.lockCalls != nil {
		*r.lockCalls++
	}
	return r.locked, r.lockedErr
}

// openReach is the answer set that lets a sow through: nobody blocked, the
// voice was heard, no decline standing. Tests about something else use it so
// they are about that something else.
func openReach() reach { return reach{heard: true} }

func TestASowAnswersTheSameThreeReachRulesAsASprout(t *testing.T) {
	// The defect this closes: POST /v1/seed/sows enforced the tier, the
	// confirmation, media ownership, screening and the allowance — and
	// nothing else. A member could sow having heard nobody, past a block,
	// past a decline, because a Sow had no target to check against.
	for _, c := range []struct {
		name  string
		rules reach
		want  error
	}{
		{"a block in either direction", reach{blocked: true, heard: true}, ErrReachNotAvailable},
		{"a voice never heard", reach{}, ErrNotHeard},
		{"a decline still standing", reach{heard: true, locked: true}, ErrReachNotAvailable},
		{"a block that cannot be read", reach{heard: true, blockErr: errors.New("down")}, ErrUnavailable},
		{"a listen gate that cannot be read", reach{heardErr: errors.New("down")}, ErrUnavailable},
		{"a decline lock that cannot be read", reach{heard: true, lockedErr: errors.New("down")}, ErrUnavailable},
	} {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			// No Screen and no Accept expectations: a refused reach must not
			// be screened, stored, or charged a seed.
			service := New(NewMockScreening(ctrl), NewMockAcceptance(ctrl), NewMockKeyer(ctrl),
				NewMockIDSource(ctrl), time.Now, 1).
				WithReachRules(c.rules, c.rules, c.rules)
			_, err := service.Send(context.Background(), Command{
				ID: "c", ActorID: "a", TargetID: "t", Body: "body", MediaRefs: []string{"mine"}, Confirmed: true,
			})
			if !errors.Is(err, c.want) {
				t.Fatalf("err = %v, want %v", err, c.want)
			}
		})
	}
}

func TestASowWithNoReachRulesComposedRefuses(t *testing.T) {
	// A missing check is not permission. If the composition root forgets to
	// wire the rules, the sow path must close rather than open — which is
	// how the gap above went unnoticed for as long as it did.
	ctrl := gomock.NewController(t)
	service := New(NewMockScreening(ctrl), NewMockAcceptance(ctrl), NewMockKeyer(ctrl),
		NewMockIDSource(ctrl), time.Now, 1)
	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body", MediaRefs: []string{"mine"}, Confirmed: true,
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestASowTowardNobodyIsRefusedBeforeAnythingIsAsked(t *testing.T) {
	// Without a target there is nothing to check, so the request is invalid
	// rather than quietly reaching everyone or no one.
	ctrl := gomock.NewController(t)
	var blocks, heards, locks int
	rules := reach{heard: true, blockCalls: &blocks, heardCalls: &heards, lockCalls: &locks}
	service := New(NewMockScreening(ctrl), NewMockAcceptance(ctrl), NewMockKeyer(ctrl),
		NewMockIDSource(ctrl), time.Now, 1).WithReachRules(rules, rules, rules)

	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "  ", Body: "body", Confirmed: true,
	}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
	if blocks+heards+locks != 0 {
		t.Fatal("a targetless sow was checked against a target")
	}
}

func TestTheReachRulesAreAskedBeforeScreeningAndTheSeed(t *testing.T) {
	// Order is the point: a sow that will not be delivered must not be sent
	// to a screener, and must not have reached the allowance. Screening and
	// acceptance carry no expectations, so either being called fails here.
	ctrl := gomock.NewController(t)
	service := New(NewMockScreening(ctrl), NewMockAcceptance(ctrl), NewMockKeyer(ctrl),
		NewMockIDSource(ctrl), time.Now, 1).
		WithMediaOwnership(ownedMedia{owned: true}).
		WithReachRules(reach{heard: true}, reach{blocked: true, heard: true}, reach{heard: true})

	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body",
		MediaRefs: []string{"ref"}, Confirmed: true,
	}); !errors.Is(err, ErrReachNotAvailable) {
		t.Fatalf("err = %v, want ErrReachNotAvailable", err)
	}
}

// placements records what was delivered, so a test can say whether anything
// arrived and what it carried.
type placements struct {
	placed []Deliverable
	err    error
}

func (p *placements) Place(_ context.Context, sow Deliverable) error {
	if p.err != nil {
		return p.err
	}
	p.placed = append(p.placed, sow)
	return nil
}

// approvedService is a sow service whose screening clears everything and
// whose store accepts everything, so a test can be about what happens after.
func approvedService(t *testing.T, ctrl *gomock.Controller, delivery Delivery) Service {
	t.Helper()
	screening, acceptance := NewMockScreening(ctrl), NewMockAcceptance(ctrl)
	keyer, ids := NewMockKeyer(ctrl), NewMockIDSource(ctrl)
	keyer.EXPECT().Key(gomock.Any(), gomock.Any()).DoAndReturn(
		func(namespace, value string) (string, error) { return namespace + ":" + value, nil }).AnyTimes()
	ids.EXPECT().NewID().Return("sow-1").AnyTimes()
	screening.EXPECT().Screen(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(ScreeningDecision{Approved: true, Reference: "screen-1"}, nil).AnyTimes()
	acceptance.EXPECT().Accept(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s domain.Sow) (domain.Sow, bool, error) { return s, false, nil }).AnyTimes()
	return New(screening, acceptance, keyer, ids, time.Now, 1).
		WithMediaOwnership(ownedMedia{owned: true}).
		WithReachRules(openReach(), openReach(), openReach()).
		WithDelivery(delivery)
}

func TestAClearedSowArrivesAtTheRecipientsHouseFront(t *testing.T) {
	// Until this existed nothing read StatusDelivered. A sow was accepted,
	// the seed was spent, the status said delivered, and the recipient never
	// learned anything had been sent.
	ctrl := gomock.NewController(t)
	placed := &placements{}
	service := approvedService(t, ctrl, placed)

	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "the-sower", TargetID: "the-recipient", Body: "body",
		MediaRefs: []string{"recording-1"}, Confirmed: true,
	}); err != nil {
		t.Fatal(err)
	}
	if len(placed.placed) != 1 {
		t.Fatalf("%d sows delivered, want 1", len(placed.placed))
	}
	// Raw on both sides: a pod cannot be placed for a digest.
	got := placed.placed[0]
	if got.SowerID != "the-sower" || got.TargetID != "the-recipient" {
		t.Fatalf("delivered to %#v", got)
	}
	if len(got.MediaRefs) != 1 || got.MediaRefs[0] != "recording-1" {
		t.Fatalf("carried %v", got.MediaRefs)
	}
}

func TestAHeldSowDoesNotArriveUntilAPersonReleasesIt(t *testing.T) {
	// The whole reason screening holds a sow. Delivering one that is waiting
	// on a reviewer would make the review decorative.
	ctrl := gomock.NewController(t)
	screening, acceptance := NewMockScreening(ctrl), NewMockAcceptance(ctrl)
	keyer, ids := NewMockKeyer(ctrl), NewMockIDSource(ctrl)
	keyer.EXPECT().Key(gomock.Any(), gomock.Any()).DoAndReturn(
		func(namespace, value string) (string, error) { return namespace + ":" + value, nil }).AnyTimes()
	ids.EXPECT().NewID().Return("sow-1")
	screening.EXPECT().Screen(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(ScreeningDecision{Reference: "review-1"}, ErrHumanReviewRequired)
	acceptance.EXPECT().Accept(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s domain.Sow) (domain.Sow, bool, error) { return s, false, nil })

	placed := &placements{}
	service := New(screening, acceptance, keyer, ids, time.Now, 1).
		WithMediaOwnership(ownedMedia{owned: true}).
		WithReachRules(openReach(), openReach(), openReach()).
		WithDelivery(placed)
	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body",
		MediaRefs: []string{"recording-1"}, Confirmed: true,
	}); err != nil {
		t.Fatal(err)
	}
	if len(placed.placed) != 0 {
		t.Fatal("a sow waiting on a reviewer was delivered anyway")
	}
}

func TestReleasingAHeldSowDeliversIt(t *testing.T) {
	ctrl := gomock.NewController(t)
	acceptance := NewMockAcceptance(ctrl)
	acceptance.EXPECT().FindByScreening(gomock.Any(), "review-1").Return(heldSow(t), nil)
	acceptance.EXPECT().Settle(gomock.Any(), gomock.Any(), false).Return(nil)

	placed := &placements{}
	service := New(NewMockScreening(ctrl), acceptance, NewMockKeyer(ctrl),
		NewMockIDSource(ctrl), time.Now, 1).WithDelivery(placed)
	if _, err := service.Review(context.Background(), "review-1", true, "decision-1"); err != nil {
		t.Fatal(err)
	}
	if len(placed.placed) != 1 || placed.placed[0].TargetID != "target" {
		t.Fatalf("released sow delivered as %#v", placed.placed)
	}
}

func TestARefusedSowIsNeverDelivered(t *testing.T) {
	ctrl := gomock.NewController(t)
	acceptance := NewMockAcceptance(ctrl)
	acceptance.EXPECT().FindByScreening(gomock.Any(), "review-1").Return(heldSow(t), nil)
	acceptance.EXPECT().Settle(gomock.Any(), gomock.Any(), true).Return(nil)

	placed := &placements{}
	service := New(NewMockScreening(ctrl), acceptance, NewMockKeyer(ctrl),
		NewMockIDSource(ctrl), time.Now, 1).WithDelivery(placed)
	if _, err := service.Review(context.Background(), "review-1", false, "decision-1"); err != nil {
		t.Fatal(err)
	}
	if len(placed.placed) != 0 {
		t.Fatal("a refused sow was delivered")
	}
}

func TestASowThatCannotBeDeliveredSaysSo(t *testing.T) {
	// The sow exists and the seed is spent, so "service unavailable" would
	// be a lie the member could act on wrongly — retrying would charge them
	// again. ErrNotDelivered is its own answer for that reason.
	ctrl := gomock.NewController(t)
	service := approvedService(t, ctrl, &placements{err: errors.New("pod store down")})
	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body",
		MediaRefs: []string{"recording-1"}, Confirmed: true,
	}); !errors.Is(err, ErrNotDelivered) {
		t.Fatalf("err = %v, want ErrNotDelivered", err)
	}
}

func TestASowServiceWithNoDeliveryRefuses(t *testing.T) {
	// A missing step is not a step that succeeded. Without delivery a sow
	// would be accepted, charged and marked delivered while nobody received
	// anything — which is exactly what shipped.
	ctrl := gomock.NewController(t)
	screening, acceptance := NewMockScreening(ctrl), NewMockAcceptance(ctrl)
	keyer, ids := NewMockKeyer(ctrl), NewMockIDSource(ctrl)
	keyer.EXPECT().Key(gomock.Any(), gomock.Any()).DoAndReturn(
		func(namespace, value string) (string, error) { return namespace + ":" + value, nil }).AnyTimes()
	ids.EXPECT().NewID().Return("sow-1")
	screening.EXPECT().Screen(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(ScreeningDecision{Approved: true, Reference: "screen-1"}, nil)
	acceptance.EXPECT().Accept(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s domain.Sow) (domain.Sow, bool, error) { return s, false, nil })

	service := New(screening, acceptance, keyer, ids, time.Now, 1).
		WithMediaOwnership(ownedMedia{owned: true}).
		WithReachRules(openReach(), openReach(), openReach())
	if _, err := service.Send(context.Background(), Command{
		ID: "c", ActorID: "a", TargetID: "t", Body: "body",
		MediaRefs: []string{"recording-1"}, Confirmed: true,
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
