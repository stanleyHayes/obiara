package application

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	sowdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
	"go.uber.org/mock/gomock"
)

var reviewedAt = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func TestReviewedMultilingualInputIsNormalizedNotTranslated(t *testing.T) {
	for _, test := range []struct {
		locale string
		body   string
		want   string
	}{
		{locale: "en", body: "  Hello   there  ", want: "Hello there"},
		{locale: "tw", body: "  Meda   wo   ase  ", want: "Meda wo ase"},
		{locale: "gaa", body: "  Oyiwaladɔŋŋ  ", want: "Oyiwaladɔŋŋ"},
	} {
		t.Run(test.locale, func(t *testing.T) {
			controller := gomock.NewController(t)
			locales := NewMockLocaleSource(controller)
			catalog := NewMockLocaleCatalog(controller)
			media := NewMockMediaInspector(controller)
			advisor := NewMockAdvisor(controller)
			adjudicator := NewMockAdjudicator(controller)
			human := NewMockHumanReview(controller)
			ids := NewMockIDSource(controller)
			input := ScreeningInput{
				Text: test.want, LocaleTag: test.locale, LocaleVersion: 3,
				Media: []MediaMetadata{{MIME: "audio/ogg", Bytes: 1024, DurationMs: 900}},
			}
			advisory := Advisory{Status: StatusApproved, Reasons: []ReasonCode{ReasonClear}, Confidence: 97}
			reference := screeningKey("e")
			gomock.InOrder(
				locales.EXPECT().CurrentLocale(gomock.Any()).Return(test.locale, nil),
				catalog.EXPECT().Resolve(gomock.Any(), test.locale).Return(
					LocaleReview{Tag: test.locale, Version: 3, Reviewed: true, ReviewedAt: reviewedAt}, nil,
				),
				media.EXPECT().Inspect(gomock.Any(), "media-ref").Return(input.Media[0], nil),
				advisor.EXPECT().Screen(gomock.Any(), input).Return(advisory, nil),
				adjudicator.EXPECT().Decide(gomock.Any(), input, advisory).Return(
					Adjudication{
						Status: StatusApproved, Reason: ReasonClear,
						Reference: reference, HumanReviewed: true,
					}, nil,
				),
			)
			decision, err := New(locales, catalog, media, advisor, adjudicator, human, ids).
				Screen(context.Background(), test.body, []string{"media-ref"})
			if err != nil || !decision.Approved || decision.Reference != reference {
				t.Fatalf("decision=%+v err=%v", decision, err)
			}
		})
	}
}

func TestUncertainUnsupportedAndProviderFailureRouteHumanReview(t *testing.T) {
	tests := []struct {
		name          string
		localeResult  LocaleReview
		localeErr     error
		advisory      Advisory
		advisorErr    error
		wantReason    ReasonCode
		expectAdvisor bool
	}{
		{
			name: "unsupported locale", localeErr: errors.New("unsupported"),
			wantReason: ReasonUnsupportedLocale,
		},
		{
			name: "uncertain", localeResult: reviewedLocale("tw"),
			advisory:   Advisory{Status: StatusUncertain, Reasons: []ReasonCode{ReasonUncertain}, Confidence: 51},
			wantReason: ReasonUncertain, expectAdvisor: true,
		},
		{
			name: "provider error", localeResult: reviewedLocale("tw"),
			advisorErr: errors.New("provider unavailable"),
			wantReason: ReasonProviderFailure, expectAdvisor: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			locales := NewMockLocaleSource(controller)
			catalog := NewMockLocaleCatalog(controller)
			media := NewMockMediaInspector(controller)
			advisor := NewMockAdvisor(controller)
			adjudicator := NewMockAdjudicator(controller)
			human := NewMockHumanReview(controller)
			ids := NewMockIDSource(controller)
			locales.EXPECT().CurrentLocale(gomock.Any()).Return("tw", nil)
			catalog.EXPECT().Resolve(gomock.Any(), "tw").Return(test.localeResult, test.localeErr)
			input := ScreeningInput{Text: "Meda wo ase", LocaleTag: "tw"}
			if test.localeErr == nil {
				input.LocaleVersion = test.localeResult.Version
				if test.expectAdvisor {
					advisor.EXPECT().Screen(gomock.Any(), input).Return(test.advisory, test.advisorErr)
				}
			}
			caseID, reviewRef := screeningKey("a"), screeningKey("b")
			ids.EXPECT().NewID().Return(caseID)
			human.EXPECT().Route(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, review ReviewCase) (string, error) {
					if review.ID != caseID || review.Reason != test.wantReason ||
						review.Input.Text != "Meda wo ase" {
						t.Fatalf("review case = %+v", review)
					}
					return reviewRef, nil
				},
			)
			decision, err := New(locales, catalog, media, advisor, adjudicator, human, ids).
				Screen(context.Background(), "Meda wo ase", nil)
			if !errors.Is(err, ErrHumanReviewRequired) || decision.Approved || decision.Reference != reviewRef {
				t.Fatalf("decision=%+v err=%v", decision, err)
			}
		})
	}
}

func TestProviderCannotFinalAdjudicate(t *testing.T) {
	controller := gomock.NewController(t)
	locales := NewMockLocaleSource(controller)
	catalog := NewMockLocaleCatalog(controller)
	media := NewMockMediaInspector(controller)
	advisor := NewMockAdvisor(controller)
	adjudicator := NewMockAdjudicator(controller)
	human := NewMockHumanReview(controller)
	ids := NewMockIDSource(controller)
	input := ScreeningInput{Text: "hello", LocaleTag: "en", LocaleVersion: 1}
	advisory := Advisory{Status: StatusRejected, Reasons: []ReasonCode{ReasonThreat}, Confidence: 100}
	locales.EXPECT().CurrentLocale(gomock.Any()).Return("en", nil)
	catalog.EXPECT().Resolve(gomock.Any(), "en").Return(reviewedLocale("en"), nil)
	advisor.EXPECT().Screen(gomock.Any(), input).Return(advisory, nil)
	adjudicator.EXPECT().Decide(gomock.Any(), input, advisory).Return(
		Adjudication{Status: StatusRejected, Reason: ReasonThreat, Reference: screeningKey("c"), HumanReviewed: false},
		nil,
	)
	ids.EXPECT().NewID().Return(screeningKey("d"))
	human.EXPECT().Route(gomock.Any(), gomock.Any()).Return(screeningKey("e"), nil)

	decision, err := New(locales, catalog, media, advisor, adjudicator, human, ids).
		Screen(context.Background(), "hello", nil)
	if !errors.Is(err, ErrHumanReviewRequired) || decision.Approved {
		t.Fatalf("provider became adjudicator: decision=%+v err=%v", decision, err)
	}
}

func TestHumanRejectedSowNeverReachesAcceptanceOrAllowanceSpend(t *testing.T) {
	controller := gomock.NewController(t)
	locales := NewMockLocaleSource(controller)
	catalog := NewMockLocaleCatalog(controller)
	media := NewMockMediaInspector(controller)
	advisor := NewMockAdvisor(controller)
	adjudicator := NewMockAdjudicator(controller)
	human := NewMockHumanReview(controller)
	screeningIDs := NewMockIDSource(controller)
	input := ScreeningInput{Text: "body", LocaleTag: "en", LocaleVersion: 1}
	advisory := Advisory{Status: StatusRejected, Reasons: []ReasonCode{ReasonPaymentRequest}, Confidence: 95}
	locales.EXPECT().CurrentLocale(gomock.Any()).Return("en", nil)
	catalog.EXPECT().Resolve(gomock.Any(), "en").Return(reviewedLocale("en"), nil)
	advisor.EXPECT().Screen(gomock.Any(), input).Return(advisory, nil)
	adjudicator.EXPECT().Decide(gomock.Any(), input, advisory).Return(
		Adjudication{
			Status: StatusRejected, Reason: ReasonPaymentRequest,
			Reference: screeningKey("f"), HumanReviewed: true,
		}, nil,
	)
	acceptance := &countingAcceptance{}
	sowService := sowapplication.New(
		New(locales, catalog, media, advisor, adjudicator, human, screeningIDs),
		acceptance,
		noopKeyer{},
		staticID{},
		func() time.Time { return reviewedAt },
		1,
	)
	_, err := sowService.Send(context.Background(), sowapplication.Command{
		ID: "command-one", ActorID: "member-one", Body: "body", Confirmed: true,
	})
	if !errors.Is(err, sowdomain.ErrScreeningRejected) || acceptance.calls != 0 {
		t.Fatalf("send err=%v acceptance calls=%d", err, acceptance.calls)
	}
}

func TestAdapterHasNoRawRetentionState(t *testing.T) {
	adapterType := reflect.TypeOf(Adapter{})
	for index := 0; index < adapterType.NumField(); index++ {
		field := adapterType.Field(index)
		if field.Type.Kind() == reflect.String || field.Type.Kind() == reflect.Slice ||
			field.Type == reflect.TypeOf(ScreeningInput{}) || field.Type == reflect.TypeOf(ReviewCase{}) {
			t.Fatalf("adapter can retain raw screening material in %s", field.Name)
		}
	}
}

type countingAcceptance struct{ calls int }

func (acceptance *countingAcceptance) Accept(_ context.Context, sow sowdomain.Sow) (sowdomain.Sow, bool, error) {
	acceptance.calls++
	return sow, false, nil
}

// These tests are about screening, not about settling a review, so both
// review-side methods refuse rather than pretend.
func (*countingAcceptance) FindByScreening(context.Context, string) (sowdomain.Sow, error) {
	return sowdomain.Sow{}, sowapplication.ErrSowNotFound
}

func (*countingAcceptance) Settle(context.Context, sowdomain.Sow, bool) error {
	return sowapplication.ErrSowNotFound
}

type noopKeyer struct{}

func (noopKeyer) Key(_, value string) (string, error) { return value, nil }

type staticID struct{}

func (staticID) NewID() string { return "sow-one" }

func reviewedLocale(tag string) LocaleReview {
	return LocaleReview{Tag: tag, Version: 1, Reviewed: true, ReviewedAt: reviewedAt}
}

func screeningKey(character string) string {
	return strings.Repeat(character, 64)
}
