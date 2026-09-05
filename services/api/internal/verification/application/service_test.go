package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
)

var testNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
var testDOB = time.Date(1998, time.April, 12, 0, 0, 0, 0, time.UTC)

func fixedNow() time.Time { return testNow }

func newService(t *testing.T) (VerificationService, *MockCaseRepository, *MockVerificationProvider, *MockTierTransitions) {
	t.Helper()
	ctrl := gomock.NewController(t)
	cases := NewMockCaseRepository(ctrl)
	provider := NewMockVerificationProvider(ctrl)
	tiers := NewMockTierTransitions(ctrl)
	keyer := NewMockCardKeyer(ctrl)
	keyer.EXPECT().Key(gomock.Any()).DoAndReturn(func(card string) (string, error) {
		return "key_" + card, nil
	}).AnyTimes()
	service := NewVerificationService(cases, provider, tiers, keyer, adultAgeGate{}, fixedNow, func() string { return "vc_test" })
	return service, cases, provider, tiers
}

func TestMatchApprovesAndPromotes(t *testing.T) {
	service, cases, provider, tiers := newService(t)
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), gomock.Any()).Return("", ErrCaseNotFound)
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(
		ProviderResult{Outcome: "match", ProviderRef: "ref-1", Reason: "issuer match"}, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.VerificationCase) error {
			if c.Status() != domain.StatusApproved {
				t.Fatalf("status = %q", c.Status())
			}
			return nil
		})
	tiers.EXPECT().Transition(gomock.Any(), "id_1", 1, gomock.Any(), gomock.Any()).Return(nil)

	verificationCase, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-1", testDOB)
	if err != nil {
		t.Fatal(err)
	}
	if verificationCase.Status() != domain.StatusApproved {
		t.Fatalf("status = %q", verificationCase.Status())
	}
}

func TestMismatchRejects(t *testing.T) {
	service, cases, provider, _ := newService(t)
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), gomock.Any()).Return("", ErrCaseNotFound)
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(
		ProviderResult{Outcome: "mismatch", ProviderRef: "ref-2", Reason: "no issuer record"}, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.VerificationCase) error {
			if c.Status() != domain.StatusRejected {
				t.Fatalf("status = %q", c.Status())
			}
			return nil
		})

	if _, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-X", testDOB); !errors.Is(err, ErrProviderRejected) {
		t.Fatalf("SubmitGhanaCard = %v, want ErrProviderRejected", err)
	}
}

func TestOutageAndUncertaintyQueueManual(t *testing.T) {
	for name, setup := range map[string]func(*MockVerificationProvider){
		"outage": func(provider *MockVerificationProvider) {
			provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(ProviderResult{}, context.DeadlineExceeded)
		},
		"uncertain": func(provider *MockVerificationProvider) {
			provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(ProviderResult{Outcome: "uncertain", Reason: "unreadable"}, nil)
		},
	} {
		t.Run(name, func(t *testing.T) {
			service, cases, provider, _ := newService(t)
			cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), gomock.Any()).Return("", ErrCaseNotFound)
			cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			setup(provider)
			cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, c domain.VerificationCase) error {
					if c.Status() != domain.StatusQueuedManual {
						t.Fatalf("status = %q, want queued_manual", c.Status())
					}
					return nil
				})

			verificationCase, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-1", testDOB)
			if err != nil {
				t.Fatal(err)
			}
			if verificationCase.Status() != domain.StatusQueuedManual {
				t.Fatalf("status = %q", verificationCase.Status())
			}
		})
	}
}

func TestDecideManualApprovalPromotes(t *testing.T) {
	service, cases, _, tiers := newService(t)
	queued := domain.ReconstituteCase("vc_1", "id_1", "key_1", "0001", domain.StatusQueuedManual, "", "provider uncertain", testDOB, 2, testNow, nil)
	cases.EXPECT().FindByID(gomock.Any(), "vc_1").Return(queued, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	tiers.EXPECT().Transition(gomock.Any(), "id_1", 1, gomock.Any(), "agent-1").Return(nil)

	verificationCase, err := service.DecideManual(context.Background(), "vc_1", true, "verified in person", "agent-1")
	if err != nil {
		t.Fatal(err)
	}
	if verificationCase.Status() != domain.StatusApproved {
		t.Fatalf("status = %q", verificationCase.Status())
	}
}

func TestIdentityAlreadyVerifiedOnDifferentAccount(t *testing.T) {
	service, cases, _, _ := newService(t)
	// Mock to report the card is already verified on a different account
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), "key_GHA-000000000-1").
		Return("other-account", nil).
		Times(1)
	// Create must NOT be called
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	_, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-1", testDOB)
	if !errors.Is(err, ErrIdentityAlreadyVerified) {
		t.Fatalf("SubmitGhanaCard = %v, want ErrIdentityAlreadyVerified", err)
	}
}

func TestIdentityAlreadyVerifiedOnSameAccount(t *testing.T) {
	service, cases, provider, tiers := newService(t)
	// Mock to report the card is already verified on THE SAME account
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), "key_GHA-000000000-1").
		Return("id_1", nil)
	// Submission proceeds normally
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(
		ProviderResult{Outcome: "match", ProviderRef: "ref-1", Reason: "issuer match"}, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	tiers.EXPECT().Transition(gomock.Any(), "id_1", 1, gomock.Any(), gomock.Any()).Return(nil)

	verificationCase, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-1", testDOB)
	if err != nil {
		t.Fatalf("SubmitGhanaCard = %v", err)
	}
	if verificationCase.Status() != domain.StatusApproved {
		t.Fatalf("status = %q", verificationCase.Status())
	}
}

func TestIdentityCheckErrorPropagates(t *testing.T) {
	service, cases, _, _ := newService(t)
	// Mock to return an unexpected error
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), gomock.Any()).
		Return("", errors.New("database error"))
	// Create must NOT be called
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	_, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-1", testDOB)
	if !strings.Contains(err.Error(), "database error") {
		t.Fatalf("SubmitGhanaCard = %v, want 'database error'", err)
	}
}

func TestCardKeysAndMasksStored(t *testing.T) {
	service, cases, provider, tiers := newService(t)
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), gomock.Any()).
		Return("", ErrCaseNotFound).AnyTimes()
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.VerificationCase) error {
			// Assert the case stores the key and mask, not the plaintext card
			if c.CardKey() != "key_GHA-123" {
				t.Fatalf("cardKey = %q, want key_GHA-123", c.CardKey())
			}
			if !strings.Contains(c.CardMask(), "123") {
				t.Fatalf("cardMask = %q, want to contain last 4 digits", c.CardMask())
			}
			return nil
		})
	provider.EXPECT().Verify(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req ProviderRequest) (ProviderResult, error) {
			// Assert the provider receives the plaintext card number
			if req.CardNumber != "GHA-123" {
				t.Fatalf("CardNumber = %q, want GHA-123", req.CardNumber)
			}
			return ProviderResult{Outcome: "match", ProviderRef: "ref-1", Reason: "issuer match"}, nil
		})
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	tiers.EXPECT().Transition(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	_, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-123", testDOB)
	if err != nil {
		t.Fatalf("SubmitGhanaCard = %v", err)
	}
}

// adultAgeGate stands in for the safeguarding context in tests that are not
// about age. It admits everybody, which is what makes the age tests below
// meaningful: they supply a gate that refuses and check that the refusal
// reaches the caller.
type adultAgeGate struct{}

func (adultAgeGate) Assess(context.Context, string, string, string, time.Time) error { return nil }
func (adultAgeGate) MinimumAge() int                                                 { return 18 }

// blockingAgeGate refuses the way the safeguarding bridge does.
type blockingAgeGate struct{ err error }

func (gate blockingAgeGate) Assess(context.Context, string, string, string, time.Time) error {
	return gate.err
}
func (blockingAgeGate) MinimumAge() int { return 18 }

func TestAnUnderageSubmissionIsRefusedBeforeAnythingIsWrittenDown(t *testing.T) {
	// The whole point of gating here rather than after the case is created:
	// a minor's card number and date of birth must never reach the store.
	ctrl := gomock.NewController(t)
	cases := NewMockCaseRepository(ctrl)
	provider := NewMockVerificationProvider(ctrl)
	tiers := NewMockTierTransitions(ctrl)
	keyer := NewMockCardKeyer(ctrl)
	keyer.EXPECT().Key(gomock.Any()).DoAndReturn(func(card string) (string, error) {
		return "key_" + card, nil
	}).AnyTimes()
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), gomock.Any()).
		Return("", ErrCaseNotFound).AnyTimes()
	// No Create, no provider call: the gomock controller fails the test if
	// either happens, which is the assertion.
	service := NewVerificationService(cases, provider, tiers, keyer,
		blockingAgeGate{err: ErrBelowMinimumAge}, fixedNow, func() string { return "vc_test" })

	_, err := service.SubmitGhanaCard(context.Background(), "member-1", "GHA-000000000-0",
		time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC))
	if !errors.Is(err, ErrBelowMinimumAge) {
		t.Fatalf("err = %v, want ErrBelowMinimumAge", err)
	}
}

func TestAnUnassessableAgeRefusesRatherThanPasses(t *testing.T) {
	// An outage in the age gate must not read as an adult. This is the
	// direction that admits a child, so it fails closed.
	ctrl := gomock.NewController(t)
	cases := NewMockCaseRepository(ctrl)
	provider := NewMockVerificationProvider(ctrl)
	tiers := NewMockTierTransitions(ctrl)
	keyer := NewMockCardKeyer(ctrl)
	keyer.EXPECT().Key(gomock.Any()).DoAndReturn(func(card string) (string, error) {
		return "key_" + card, nil
	}).AnyTimes()
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), gomock.Any()).
		Return("", ErrCaseNotFound).AnyTimes()

	for name, gate := range map[string]AgeGate{
		"gate reports an outage": blockingAgeGate{err: ErrAgeGateUnavailable},
		"no gate composed":       nil,
	} {
		service := NewVerificationService(cases, provider, tiers, keyer, gate,
			fixedNow, func() string { return "vc_test" })
		if _, err := service.SubmitGhanaCard(context.Background(), "member-1", "GHA-000000000-0",
			time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)); err == nil {
			t.Fatalf("%s: an unassessed date of birth was admitted", name)
		}
	}
}

func TestAnApprovedCaseRecordsThatTheAgeWasAssuredAndHow(t *testing.T) {
	// TS-AGE-ASSURANCE. Until this record existed, a case that passed the age
	// gate proved it only by existing — and thirty days later retention
	// strips the birth date, leaving nothing at all to show a check happened.
	service, cases, provider, tiers := newService(t)
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), gomock.Any()).Return("", ErrCaseNotFound)
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.VerificationCase) error {
			// At creation the date is still only the member's word for it.
			if got := c.AgeAssurance().Method; got != domain.AgeSelfDeclared {
				t.Fatalf("method at creation = %q, want self-declared", got)
			}
			if c.AgeAssurance().MinimumAge != 18 || c.AgeAssurance().AssuredAt.IsZero() {
				t.Fatalf("assurance = %#v", c.AgeAssurance())
			}
			return nil
		})
	provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(
		ProviderResult{Outcome: "match", ProviderRef: "ref-1", Reason: "issuer match"}, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, c domain.VerificationCase) error {
			// The birth date went to the provider in the request it matched,
			// so the claim is stronger now and the record has to say so.
			if got := c.AgeAssurance().Method; got != domain.AgeIssuerCorroborated {
				t.Fatalf("method after a match = %q, want issuer-corroborated", got)
			}
			return nil
		})
	tiers.EXPECT().Transition(gomock.Any(), "id_1", 1, gomock.Any(), gomock.Any()).Return(nil)

	if _, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-1", testDOB); err != nil {
		t.Fatal(err)
	}
}

func TestAnUnmatchedCaseDoesNotClaimTheIssuerAgreed(t *testing.T) {
	// Overstating the weaker claim is the failure this distinction exists to
	// prevent: a queued case has been checked against nothing but the form.
	service, cases, provider, _ := newService(t)
	cases.EXPECT().ApprovedAccountByCardKey(gomock.Any(), gomock.Any()).Return("", ErrCaseNotFound)
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(
		ProviderResult{Outcome: "uncertain", ProviderRef: "ref-1", Reason: "unreadable"}, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	verificationCase, err := service.SubmitGhanaCard(context.Background(), "id_1", "GHA-000000000-1", testDOB)
	if err != nil {
		t.Fatal(err)
	}
	if got := verificationCase.AgeAssurance().Method; got != domain.AgeSelfDeclared {
		t.Fatalf("an unmatched case claims %q", got)
	}
}
