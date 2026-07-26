package simulator

import (
	"context"
	"errors"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
)

func TestProviderContract(t *testing.T) {
	tests := []struct {
		name        string
		voice       string
		outcome     application.ProviderOutcome
		unavailable bool
	}{
		{"live", "voice:live", application.OutcomeLive, false},
		{"not live", "voice:fail", application.OutcomeNotLive, false},
		{"uncertain", "voice:uncertain", application.OutcomeUncertain, false},
		{"outage", "voice:outage", "", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := NewProvider()
			result, err := provider.Assess(context.Background(), application.ProviderRequest{
				CommandID: "command:1", AttemptID: "attempt:1",
				VoiceArtifactRef: test.voice, FaceArtifactRef: "face:1",
			})
			if test.unavailable {
				if !errors.Is(err, ErrUnavailable) {
					t.Fatalf("expected outage, got %v", err)
				}
				return
			}
			if err != nil || result.Outcome != test.outcome || result.ProviderRef != "sim:attempt:1" {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
			if requests := provider.Requests(); len(requests) != 1 || requests[0].CommandID != "command:1" {
				t.Fatalf("requests=%+v", requests)
			}
		})
	}
}

func TestProviderRequestHistoryIsRaceSafeAndCopied(t *testing.T) {
	provider := NewProvider()
	request := application.ProviderRequest{
		CommandID: "command:1", AttemptID: "attempt:1",
		VoiceArtifactRef: "voice:live", FaceArtifactRef: "face:1",
	}
	done := make(chan struct{})
	for range 16 {
		go func() {
			_, _ = provider.Assess(context.Background(), request)
			done <- struct{}{}
		}()
	}
	for range 16 {
		<-done
	}
	requests := provider.Requests()
	requests[0].CommandID = "mutated"
	if provider.Requests()[0].CommandID != "command:1" {
		t.Fatal("request history leaked mutable storage")
	}
}
