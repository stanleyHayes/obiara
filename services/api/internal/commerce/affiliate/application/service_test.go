package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/domain"
	"go.uber.org/mock/gomock"
)

var now = time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)

type keyerStub struct{ err error }

func (k keyerStub) MemberKey(string) (string, error) {
	if k.err != nil {
		return "", k.err
	}
	return strings.Repeat("a", 64), nil
}

type fixedID string

func (i fixedID) NewID() string { return string(i) }

func notAMember(context.Context, string) (bool, error) { return false, nil }

func affiliateFor(t *testing.T) domain.Affiliate {
	t.Helper()
	affiliate, err := domain.Register(
		"aff_1", "Campus Reps", "CAMPUS26", "reps@example.test",
		domain.Command{ID: "cmd_1", ReasonCode: "partner_agreement", At: now})
	if err != nil {
		t.Fatal(err)
	}
	return affiliate
}

func service(
	ctrl *gomock.Controller, repository *MockRepository, referrals *MockReferrals,
	qualification *MockQualification,
) Service {
	return New(repository, referrals, qualification, nil, keyerStub{}, fixedID("aff_1"),
		2000, func() time.Time { return now })
}

func TestAMemberCannotBeAnAffiliate(t *testing.T) {
	// The owner's decision, and the one with the most consequence in it:
	// paying somebody inside the community to recruit changes what "why is
	// this person talking to me" means.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	// No Create expectation: nothing may be written.

	_, err := service(ctrl, repository, NewMockReferrals(ctrl), NewMockQualification(ctrl)).
		Register(context.Background(), RegisterCommand{
			CommandID: "cmd_1", ReasonCode: "partner_agreement",
			Name: "Someone", Code: "CAMPUS26", Email: "member@example.test",
		}, func(context.Context, string) (bool, error) { return true, nil })
	if !errors.Is(err, ErrMemberAffiliate) {
		t.Fatalf("err = %v, want ErrMemberAffiliate", err)
	}
}

func TestWithNoWayToTellAMemberNothingIsRegistered(t *testing.T) {
	// A missing check is not permission — the same rule the reach rules, the
	// media policy and the sow's arrival check all follow.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	_, err := service(ctrl, repository, NewMockReferrals(ctrl), NewMockQualification(ctrl)).
		Register(context.Background(), RegisterCommand{
			CommandID: "cmd_1", ReasonCode: "partner_agreement",
			Name: "Someone", Code: "CAMPUS26", Email: "someone@example.test",
		}, nil)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestAttributingACodeAccruesNothingYet(t *testing.T) {
	// The whole difference between paying for a signup and paying for
	// somebody who stayed.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	referrals := NewMockReferrals(ctrl)
	repository.EXPECT().FindByCode(gomock.Any(), "CAMPUS26").Return(affiliateFor(t), nil)
	referrals.EXPECT().Find(gomock.Any(), gomock.Any()).Return(Referral{}, ErrNotFound)
	referrals.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, referral Referral) error {
			if !referral.QualifiesAt.Equal(now.Add(WaitingPeriod)) {
				t.Fatalf("qualifies at %v, want thirty days out", referral.QualifiesAt)
			}
			if referral.Settled != "" {
				t.Fatal("a referral was settled the moment it was made")
			}
			// Keyed. Nothing reads this back to a person.
			if len(referral.MemberKey) != 64 {
				t.Fatalf("member recorded as %q", referral.MemberKey)
			}
			return nil
		})
	// No Append expectation on the repository: nothing accrues here.

	if err := service(ctrl, repository, referrals, NewMockQualification(ctrl)).
		Attribute(context.Background(), "campus26", "member-1"); err != nil {
		t.Fatal(err)
	}
}

func TestAMistypedCodeDoesNotStopSomebodyJoining(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByCode(gomock.Any(), "WRONG").Return(domain.Affiliate{}, ErrNotFound)

	if err := service(ctrl, repository, NewMockReferrals(ctrl), NewMockQualification(ctrl)).
		Attribute(context.Background(), "wrong", "member-1"); err != nil {
		t.Fatalf("a mistyped code refused a signup: %v", err)
	}
}

func TestTheFirstCodeKeepsTheReferral(t *testing.T) {
	// Otherwise a second code typed later moves a referral somebody has
	// already been waiting thirty days on.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	referrals := NewMockReferrals(ctrl)
	repository.EXPECT().FindByCode(gomock.Any(), "CAMPUS26").Return(affiliateFor(t), nil)
	referrals.EXPECT().Find(gomock.Any(), gomock.Any()).
		Return(Referral{AffiliateID: "somebody_else"}, nil)
	// No Record expectation: the existing referral stands.

	if err := service(ctrl, repository, referrals, NewMockQualification(ctrl)).
		Attribute(context.Background(), "CAMPUS26", "member-1"); err != nil {
		t.Fatal(err)
	}
}

func TestASuspendedAffiliateBringsNobody(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	suspended, err := affiliateFor(t).Suspend(domain.Command{
		ID: "cmd_2", ReasonCode: "agreement_ended", At: now})
	if err != nil {
		t.Fatal(err)
	}
	repository.EXPECT().FindByCode(gomock.Any(), "CAMPUS26").Return(suspended, nil)
	// No Record expectation.

	if err := service(ctrl, repository, NewMockReferrals(ctrl), NewMockQualification(ctrl)).
		Attribute(context.Background(), "CAMPUS26", "member-1"); err != nil {
		t.Fatal(err)
	}
}

func dueReferral() Referral {
	return Referral{
		MemberKey: strings.Repeat("a", 64), MemberID: "member-1", AffiliateID: "aff_1", Code: "CAMPUS26",
		RecordedAt: now.Add(-WaitingPeriod), QualifiesAt: now.Add(-time.Hour),
	}
}

func TestCommissionAccruesOnlyWhenEveryQuestionSaysYes(t *testing.T) {
	// Tier 1, still there after the waiting period, and no upheld safety
	// finding. Any "no" is a no: getting this wrong in the generous direction
	// pays people to bring accounts the safety model exists to keep out.
	for name, answers := range map[string]struct{ verified, clean bool }{
		"not verified":      {false, true},
		"an upheld finding": {true, false},
		"neither":           {false, false},
	} {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repository := NewMockRepository(ctrl)
			referrals := NewMockReferrals(ctrl)
			qualification := NewMockQualification(ctrl)
			referrals.EXPECT().DueForQualification(gomock.Any(), now, 50).
				Return([]Referral{dueReferral()}, nil)
			qualification.EXPECT().Verified(gomock.Any(), gomock.Any()).
				Return(answers.verified, nil).AnyTimes()
			qualification.EXPECT().Clean(gomock.Any(), gomock.Any()).
				Return(answers.clean, nil).AnyTimes()
			referrals.EXPECT().MarkSettled(gomock.Any(), gomock.Any(), "refused").Return(nil)
			// No Append expectation: nothing may accrue.

			accrued, refused, err := service(ctrl, repository, referrals, qualification).
				QualifyDue(context.Background(), 50)
			if err != nil || accrued != 0 || refused != 1 {
				t.Fatalf("accrued %d refused %d err %v", accrued, refused, err)
			}
		})
	}
}

func TestAQualifiedConversionEarnsTheFlatCommission(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	referrals := NewMockReferrals(ctrl)
	qualification := NewMockQualification(ctrl)
	referrals.EXPECT().DueForQualification(gomock.Any(), now, 50).
		Return([]Referral{dueReferral()}, nil)
	qualification.EXPECT().Verified(gomock.Any(), gomock.Any()).Return(true, nil)
	qualification.EXPECT().Clean(gomock.Any(), gomock.Any()).Return(true, nil)
	repository.EXPECT().FindByID(gomock.Any(), "aff_1").Return(affiliateFor(t), nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), gomock.Any()).DoAndReturn(
		func(_ context.Context, affiliate domain.Affiliate, _ uint64, _ string) error {
			// A flat amount, so an affiliate's statement says nothing about
			// what any individual member paid.
			if affiliate.Balance() != 2000 {
				t.Fatalf("balance = %d", affiliate.Balance())
			}
			if affiliate.Conversions() != 1 {
				t.Fatalf("conversions = %d", affiliate.Conversions())
			}
			return nil
		})
	referrals.EXPECT().MarkSettled(gomock.Any(), gomock.Any(), "qualified").Return(nil)

	accrued, refused, err := service(ctrl, repository, referrals, qualification).
		QualifyDue(context.Background(), 50)
	if err != nil || accrued != 1 || refused != 0 {
		t.Fatalf("accrued %d refused %d err %v", accrued, refused, err)
	}
}

func TestAQuestionThatCannotBeAnsweredLeavesTheReferralPending(t *testing.T) {
	// Not a conversion and not a refusal. Leaving it pending means the next
	// sweep asks again rather than writing down a guess about somebody's
	// safety record.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	referrals := NewMockReferrals(ctrl)
	qualification := NewMockQualification(ctrl)
	referrals.EXPECT().DueForQualification(gomock.Any(), now, 50).
		Return([]Referral{dueReferral()}, nil)
	qualification.EXPECT().Verified(gomock.Any(), gomock.Any()).
		Return(false, errors.New("identity unavailable"))
	// No MarkSettled and no Append expectations.

	accrued, refused, err := service(ctrl, repository, referrals, qualification).
		QualifyDue(context.Background(), 50)
	if err != nil || accrued != 0 || refused != 0 {
		t.Fatalf("accrued %d refused %d err %v", accrued, refused, err)
	}
}

func TestAClawbackReversesAQualifiedReferral(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	referrals := NewMockReferrals(ctrl)
	earned, err := affiliateFor(t).Accrue(strings.Repeat("a", 64), 2000, domain.Command{
		ID: "accrue:" + strings.Repeat("a", 64), ReasonCode: "qualified_conversion", At: now})
	if err != nil {
		t.Fatal(err)
	}
	settled := dueReferral()
	settled.Settled = "qualified"
	referrals.EXPECT().Find(gomock.Any(), gomock.Any()).Return(settled, nil)
	repository.EXPECT().FindByID(gomock.Any(), "aff_1").Return(earned, nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), earned.Revision(), gomock.Any()).DoAndReturn(
		func(_ context.Context, affiliate domain.Affiliate, _ uint64, _ string) error {
			if affiliate.Balance() != 0 {
				t.Fatalf("balance = %d after a clawback", affiliate.Balance())
			}
			return nil
		})

	if err := service(ctrl, repository, referrals, NewMockQualification(ctrl)).
		ClawBack(context.Background(), "member-1", "membership_refunded"); err != nil {
		t.Fatal(err)
	}
}

func TestClawingBackAReferralThatNeverQualifiedDoesNothing(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	referrals := NewMockReferrals(ctrl)
	referrals.EXPECT().Find(gomock.Any(), gomock.Any()).Return(dueReferral(), nil)
	// No FindByID and no Append: nothing to reverse.

	if err := service(ctrl, repository, referrals, NewMockQualification(ctrl)).
		ClawBack(context.Background(), "member-1", "membership_refunded"); err != nil {
		t.Fatal(err)
	}
}

func TestAServiceWithNoCommissionRefuses(t *testing.T) {
	// A commission of nothing is not a scheme, and silently accruing zero
	// would look like it was working.
	bare := New(nil, nil, nil, nil, keyerStub{}, fixedID("aff_1"), 0, func() time.Time { return now })
	if _, err := bare.List(context.Background(), 10); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
