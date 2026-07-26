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

func TestEveryNonSafetyDestinationIsDenied(t *testing.T) {
	for _, destination := range []Destination{
		DestinationMatchingFeature,
		DestinationExplanation,
		DestinationRanking,
		DestinationTrust,
		DestinationAIPrompt,
	} {
		if err := Permit(destination, minimalSafetyFields); !errors.Is(err, ErrDenied) {
			t.Fatalf("%s = %v", destination, err)
		}
		if err := Permit(destination, nil); !errors.Is(err, ErrDenied) {
			t.Fatalf("%s empty = %v", destination, err)
		}
	}
}

func TestSafetyEgressRequiresExactMinimalShape(t *testing.T) {
	if err := Permit(DestinationSafetyEscalation, minimalSafetyFields); err != nil {
		t.Fatal(err)
	}
	for index := range minimalSafetyFields {
		missing := append([]Field(nil), minimalSafetyFields...)
		missing = append(missing[:index], missing[index+1:]...)
		if err := Permit(DestinationSafetyEscalation, missing); !errors.Is(err, ErrDenied) {
			t.Fatalf("missing field %d = %v", index, err)
		}
	}
	extra := append(append([]Field(nil), minimalSafetyFields...), Field("content"))
	if err := Permit(DestinationSafetyEscalation, extra); !errors.Is(err, ErrDenied) {
		t.Fatalf("content field = %v", err)
	}
	duplicate := append([]Field(nil), minimalSafetyFields...)
	duplicate[len(duplicate)-1] = FieldSubjectKey
	if err := Permit(DestinationSafetyEscalation, duplicate); !errors.Is(err, ErrDenied) {
		t.Fatalf("duplicate field = %v", err)
	}
}

func TestSafetyEventCannotCarryCounselData(t *testing.T) {
	event, err := NewSafetyEvent(
		strings.Repeat("a", 64),
		strings.Repeat("b", 64),
		ReasonExplicitSafetySupport,
		time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC),
		7,
	)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"content", "topic", "attendance", "attendee", "session", "outcome",
		"matching", "explanation", "ranking", "trust", "prompt", "actor",
	} {
		if strings.Contains(strings.ToLower(string(encoded)), forbidden) {
			t.Fatalf("event leaked %q: %s", forbidden, encoded)
		}
	}
}

func TestPermitOrderProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(20260726))
	for trial := 0; trial < 1000; trial++ {
		fields := append([]Field(nil), minimalSafetyFields...)
		rng.Shuffle(len(fields), func(i, j int) { fields[i], fields[j] = fields[j], fields[i] })
		if err := Permit(DestinationSafetyEscalation, fields); err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
	}
}

func TestPurePolicyIsRaceSafe(t *testing.T) {
	fields := append([]Field(nil), minimalSafetyFields...)
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				if err := Permit(DestinationSafetyEscalation, fields); err != nil {
					t.Errorf("concurrent permit: %v", err)
					return
				}
				if !reflect.DeepEqual(fields, minimalSafetyFields) {
					t.Error("policy mutated caller fields")
					return
				}
			}
		}()
	}
	wait.Wait()
}

func FuzzOnlyExactSafetyShapeLeaves(f *testing.F) {
	f.Add("safety_escalation", []byte{0, 1, 2, 3, 4})
	f.Add("matching_feature", []byte{0, 1, 2, 3, 4})
	f.Fuzz(func(t *testing.T, destinationRaw string, fieldIndexes []byte) {
		if len(fieldIndexes) > 10 {
			fieldIndexes = fieldIndexes[:10]
		}
		all := append(append([]Field(nil), minimalSafetyFields...),
			Field("content"), Field("topic"), Field("attendance"), Field("outcome"))
		fields := make([]Field, len(fieldIndexes))
		for index, value := range fieldIndexes {
			fields[index] = all[int(value)%len(all)]
		}
		err := Permit(Destination(destinationRaw), fields)
		allowed := Destination(destinationRaw) == DestinationSafetyEscalation &&
			len(fields) == len(minimalSafetyFields)
		if allowed {
			sorted := append([]Field(nil), fields...)
			required := append([]Field(nil), minimalSafetyFields...)
			slices.Sort(sorted)
			slices.Sort(required)
			allowed = slices.Equal(sorted, required)
		}
		if (err == nil) != allowed {
			t.Fatalf("destination=%q fields=%v err=%v", destinationRaw, fields, err)
		}
	})
}
