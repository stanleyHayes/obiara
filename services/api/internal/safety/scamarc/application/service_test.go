package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/safety/scamarc/domain"
	"go.uber.org/mock/gomock"
)

var applicationTime = time.Date(2026, 7, 26, 16, 0, 0, 0, time.UTC)

func TestMultiEventSignalAlwaysRoutesHumanAfterFreshRevalidation(t *testing.T) {
	controller := gomock.NewController(t)
	consent := NewMockConsent(controller)
	authority := NewMockAuthority(controller)
	rules := NewMockRuleCatalog(controller)
	events := NewMockEventSource(controller)
	human := NewMockHumanRoute(controller)
	ids := NewMockIDSource(controller)
	clock := NewMockClock(controller)
	pair := applicationKey("a")
	ruleSet := applicationRules(t)
	eventInput := applicationEvents(pair, 3)
	signalID, sequenceID := applicationKey("f"), applicationKey("d")
	consent.EXPECT().CurrentAllows(gomock.Any(), pair, MonitoringPurpose, uint64(4)).Return(true, nil)
	authority.EXPECT().AuthorizeEvaluation(gomock.Any(), pair).Return(nil)
	rules.EXPECT().Current(gomock.Any()).Return(ruleSet, nil)
	events.EXPECT().Current(gomock.Any(), pair, domain.MaxEvents).Return(sequenceID, eventInput, nil)
	ids.EXPECT().NewID().Return(signalID)
	clock.EXPECT().Now().Return(applicationTime)
	consent.EXPECT().CurrentAllows(gomock.Any(), pair, MonitoringPurpose, uint64(4)).Return(true, nil)
	authority.EXPECT().AuthorizeEvaluation(gomock.Any(), pair).Return(nil)
	human.EXPECT().Route(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, signal domain.Signal) error {
			if signal.ID != signalID || signal.PairKey != pair ||
				signal.RuleVersion != 1 || signal.EventCount != 3 ||
				signal.Recommendation != domain.RecommendEducation {
				t.Fatalf("signal = %+v", signal)
			}
			return nil
		},
	)
	result, err := New(consent, authority, rules, events, human, ids, clock).Evaluate(
		context.Background(),
		Request{PairKey: pair, ConsentVersion: 4},
	)
	if err != nil || result.Signal == nil {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestSingleEventProducesNoSignalAndNoHumanRoute(t *testing.T) {
	controller := gomock.NewController(t)
	consent := NewMockConsent(controller)
	authority := NewMockAuthority(controller)
	rules := NewMockRuleCatalog(controller)
	events := NewMockEventSource(controller)
	human := NewMockHumanRoute(controller)
	ids := NewMockIDSource(controller)
	clock := NewMockClock(controller)
	pair := applicationKey("a")
	authority.EXPECT().AuthorizeEvaluation(gomock.Any(), pair).Return(nil)
	consent.EXPECT().CurrentAllows(gomock.Any(), pair, MonitoringPurpose, uint64(1)).Return(true, nil)
	rules.EXPECT().Current(gomock.Any()).Return(applicationRules(t), nil)
	events.EXPECT().Current(gomock.Any(), pair, domain.MaxEvents).Return(
		applicationKey("d"), applicationEvents(pair, 1), nil,
	)
	ids.EXPECT().NewID().Return(applicationKey("f"))
	clock.EXPECT().Now().Return(applicationTime)
	result, err := New(consent, authority, rules, events, human, ids, clock).Evaluate(
		context.Background(),
		Request{PairKey: pair, ConsentVersion: 1},
	)
	if err != nil || result.Signal != nil {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestConsentWithdrawalBeforeRouteFailsClosed(t *testing.T) {
	controller := gomock.NewController(t)
	consent := NewMockConsent(controller)
	authority := NewMockAuthority(controller)
	rules := NewMockRuleCatalog(controller)
	events := NewMockEventSource(controller)
	human := NewMockHumanRoute(controller)
	ids := NewMockIDSource(controller)
	clock := NewMockClock(controller)
	pair := applicationKey("a")
	authority.EXPECT().AuthorizeEvaluation(gomock.Any(), pair).Return(nil)
	consent.EXPECT().CurrentAllows(gomock.Any(), pair, MonitoringPurpose, uint64(2)).Return(true, nil)
	rules.EXPECT().Current(gomock.Any()).Return(applicationRules(t), nil)
	events.EXPECT().Current(gomock.Any(), pair, domain.MaxEvents).Return(
		applicationKey("d"), applicationEvents(pair, 3), nil,
	)
	ids.EXPECT().NewID().Return(applicationKey("f"))
	clock.EXPECT().Now().Return(applicationTime)
	consent.EXPECT().CurrentAllows(gomock.Any(), pair, MonitoringPurpose, uint64(2)).Return(false, nil)
	if _, err := New(consent, authority, rules, events, human, ids, clock).Evaluate(
		context.Background(),
		Request{PairKey: pair, ConsentVersion: 2},
	); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("withdrawal = %v", err)
	}
}

func applicationRules(t *testing.T) domain.RuleSet {
	t.Helper()
	rules, err := domain.NewRuleSet(domain.RuleSpec{
		ID:      "scam.arc.reviewed",
		Version: 1,
		Window:  7 * 24 * time.Hour,
		Review: domain.Review{
			ID:          "review.v1",
			ReviewerKey: applicationKey("e"),
			ReviewedAt:  applicationTime.Add(-time.Hour),
		},
		Steps: []domain.Step{
			{MinimumEvents: 2, MinimumCategories: 1, Recommendation: domain.RecommendObserve},
			{MinimumEvents: 3, MinimumCategories: 2, Recommendation: domain.RecommendEducation},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

func applicationEvents(pair string, count int) []domain.Event {
	categories := []domain.Category{
		domain.CategoryUrgencyPressure,
		domain.CategorySecrecyRequest,
		domain.CategoryChannelShift,
	}
	events := make([]domain.Event, count)
	for index := range events {
		suffix := "0123456789abcdef"
		events[index] = domain.Event{
			ID:            strings.Repeat("0", 63) + string(suffix[index+1]),
			PairKey:       pair,
			Category:      categories[index%len(categories)],
			ObservedAt:    applicationTime.Add(-time.Duration(count-index) * time.Hour),
			SourceVersion: 1,
		}
	}
	return events
}

func applicationKey(character string) string {
	return strings.Repeat(character, 64)
}
