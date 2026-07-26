package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

var fixedTime = time.Date(2026, 7, 26, 16, 0, 0, 0, time.UTC)

func TestSingleEventAndOutOfWindowEventsNeverSignal(t *testing.T) {
	rules := reviewedRules(t)
	single := mustSequence(t, []Event{event(1, CategoryUrgencyPressure, fixedTime.Add(-time.Hour))})
	if _, err := Evaluate(opaqueKey("f"), single, rules, fixedTime); !errors.Is(err, ErrNoPattern) {
		t.Fatalf("single event = %v", err)
	}
	old := mustSequence(t, []Event{
		event(1, CategoryUrgencyPressure, fixedTime.Add(-8*24*time.Hour)),
		event(2, CategorySecrecyRequest, fixedTime.Add(-7*24*time.Hour)),
	})
	if _, err := Evaluate(opaqueKey("f"), old, rules, fixedTime); !errors.Is(err, ErrNoPattern) {
		t.Fatalf("old events = %v", err)
	}
}

func TestReviewedRulesEscalateLeastHarmRecommendation(t *testing.T) {
	rules := reviewedRules(t)
	categories := []Category{
		CategoryChannelShift,
		CategoryUrgencyPressure,
		CategorySecrecyRequest,
		CategoryIsolationPrompt,
	}
	wants := map[int]Recommendation{
		2: RecommendObserve,
		3: RecommendEducation,
		5: RecommendFriction,
		8: RecommendHumanCase,
	}
	var events []Event
	for count := 1; count <= 8; count++ {
		events = append(events, event(count, categories[(count-1)%len(categories)], fixedTime.Add(-time.Duration(9-count)*time.Hour)))
		signal, err := Evaluate(opaqueKey("f"), mustSequence(t, events), rules, fixedTime)
		if count == 1 {
			if !errors.Is(err, ErrNoPattern) {
				t.Fatalf("count 1 = %v", err)
			}
			continue
		}
		want, checkpoint := wants[count]
		if checkpoint && (err != nil || signal.Recommendation != want) {
			t.Fatalf("count %d recommendation=%s err=%v want=%s", count, signal.Recommendation, err, want)
		}
	}
}

func TestEvaluationIsDeterministicAcrossEventLoadOrder(t *testing.T) {
	events := []Event{
		event(1, CategoryUrgencyPressure, fixedTime.Add(-3*time.Hour)),
		event(2, CategorySecrecyRequest, fixedTime.Add(-2*time.Hour)),
		event(3, CategoryChannelShift, fixedTime.Add(-time.Hour)),
	}
	forward, err := Evaluate(opaqueKey("f"), mustSequence(t, events), reviewedRules(t), fixedTime)
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(events)
	reversed, err := Evaluate(opaqueKey("f"), mustSequence(t, events), reviewedRules(t), fixedTime)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(forward, reversed) {
		t.Fatalf("load order changed signal: forward=%+v reversed=%+v", forward, reversed)
	}
}

func TestRulesRequireVersionedReviewAndMultiEventSteps(t *testing.T) {
	spec := reviewedRuleSpec()
	spec.Version = 0
	if _, err := NewRuleSet(spec); !errors.Is(err, ErrInvalidRules) {
		t.Fatalf("unversioned = %v", err)
	}
	spec = reviewedRuleSpec()
	spec.Review = Review{}
	if _, err := NewRuleSet(spec); !errors.Is(err, ErrInvalidRules) {
		t.Fatalf("unreviewed = %v", err)
	}
	spec = reviewedRuleSpec()
	spec.Steps[0].MinimumEvents = 1
	if _, err := NewRuleSet(spec); !errors.Is(err, ErrInvalidRules) {
		t.Fatalf("single-event rule = %v", err)
	}
}

func TestSignalHasNoRawOrEnforcementShape(t *testing.T) {
	signal, err := Evaluate(
		opaqueKey("f"),
		mustSequence(t, []Event{
			event(1, CategoryUrgencyPressure, fixedTime.Add(-time.Hour)),
			event(2, CategorySecrecyRequest, fixedTime),
		}),
		reviewedRules(t),
		fixedTime,
	)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(signal)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"message", "voice", "payment", "content", "accus", "score", "member",
		"block", "account", "charge", "vendor", "model", "auto",
	} {
		if strings.Contains(strings.ToLower(string(encoded)), forbidden) {
			t.Fatalf("signal leaked %q: %s", forbidden, encoded)
		}
	}
}

func TestSequenceRecommendationProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(20260726))
	categories := []Category{
		CategoryChannelShift, CategoryUrgencyPressure, CategorySecrecyRequest,
		CategoryIsolationPrompt, CategoryResourceRequest,
	}
	for trial := 0; trial < 1000; trial++ {
		count := rng.Intn(MaxEvents-1) + 2
		events := make([]Event, count)
		for index := range events {
			events[index] = event(
				index+1,
				categories[rng.Intn(len(categories))],
				fixedTime.Add(-time.Duration(rng.Intn(6*24))*time.Hour),
			)
		}
		signal, err := Evaluate(opaqueKey("f"), mustSequence(t, events), reviewedRules(t), fixedTime)
		if err != nil && !errors.Is(err, ErrNoPattern) {
			t.Fatalf("trial %d: %v", trial, err)
		}
		if err == nil && recommendationOrder(signal.Recommendation) < 0 {
			t.Fatalf("trial %d invalid recommendation %+v", trial, signal)
		}
	}
}

func TestPureEvaluationIsRaceSafe(t *testing.T) {
	sequence := mustSequence(t, []Event{
		event(1, CategoryUrgencyPressure, fixedTime.Add(-time.Hour)),
		event(2, CategorySecrecyRequest, fixedTime),
	})
	rules := reviewedRules(t)
	want, err := Evaluate(opaqueKey("f"), sequence, rules, fixedTime)
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				got, evaluateErr := Evaluate(opaqueKey("f"), sequence, rules, fixedTime)
				if evaluateErr != nil || !reflect.DeepEqual(got, want) {
					t.Errorf("concurrent evaluate: %v", evaluateErr)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func FuzzSingleEventNeverCreatesSignal(f *testing.F) {
	f.Add(uint8(0), int64(0))
	f.Add(uint8(4), int64(24))
	f.Fuzz(func(t *testing.T, categoryRaw uint8, hoursAgo int64) {
		categories := []Category{
			CategoryChannelShift, CategoryUrgencyPressure, CategorySecrecyRequest,
			CategoryIsolationPrompt, CategoryResourceRequest,
		}
		if hoursAgo < 0 {
			hoursAgo = -hoursAgo
		}
		hoursAgo %= 24 * 365
		sequence := mustSequence(t, []Event{
			event(1, categories[int(categoryRaw)%len(categories)], fixedTime.Add(-time.Duration(hoursAgo)*time.Hour)),
		})
		if _, err := Evaluate(opaqueKey("f"), sequence, reviewedRules(t), fixedTime); !errors.Is(err, ErrNoPattern) {
			t.Fatalf("single event produced signal: %v", err)
		}
	})
}

func reviewedRules(t *testing.T) RuleSet {
	t.Helper()
	rules, err := NewRuleSet(reviewedRuleSpec())
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

func reviewedRuleSpec() RuleSpec {
	return RuleSpec{
		ID:      "scam.arc.reviewed",
		Version: 3,
		Window:  7 * 24 * time.Hour,
		Review: Review{
			ID:          "review.v3",
			ReviewerKey: opaqueKey("e"),
			ReviewedAt:  fixedTime.Add(-24 * time.Hour),
		},
		Steps: []Step{
			{MinimumEvents: 2, MinimumCategories: 1, Recommendation: RecommendObserve},
			{MinimumEvents: 3, MinimumCategories: 2, Recommendation: RecommendEducation},
			{MinimumEvents: 5, MinimumCategories: 3, Recommendation: RecommendFriction},
			{MinimumEvents: 8, MinimumCategories: 4, Recommendation: RecommendHumanCase},
		},
	}
}

func mustSequence(t *testing.T, events []Event) Sequence {
	t.Helper()
	sequence, err := NewSequence(opaqueKey("d"), opaqueKey("a"), events)
	if err != nil {
		t.Fatal(err)
	}
	return sequence
}

func event(number int, category Category, at time.Time) Event {
	hexadecimal := "0123456789abcdef"
	suffix := string(hexadecimal[(number/16)%16]) + string(hexadecimal[number%16])
	return Event{
		ID:            strings.Repeat("0", 62) + suffix,
		PairKey:       opaqueKey("a"),
		Category:      category,
		ObservedAt:    at,
		SourceVersion: 1,
	}
}

func opaqueKey(character string) string {
	return strings.Repeat(character, 64)
}
