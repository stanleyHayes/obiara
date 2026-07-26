package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

var analyticsNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestEmitValidatesConsentsAndPseudonymizes(t *testing.T) {
	ctrl := gomock.NewController(t)
	sink := NewMockEventSink(ctrl)
	consent := NewMockConsentGate(ctrl)

	consent.EXPECT().AllowsAnalytics(gomock.Any(), "m-1").Return(true, nil)
	sink.EXPECT().Append(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, event Event) error {
			if event.Name != "epono.pod_heard" {
				t.Fatalf("event = %#v", event)
			}
			if event.SubjectRef == "m-1" || event.SubjectRef == "" {
				t.Fatalf("subject not pseudonymized: %q", event.SubjectRef)
			}
			if event.SubjectRef != Pseudonym("m-1") {
				t.Fatal("pseudonym must be deterministic")
			}
			return nil
		})

	service := NewAnalyticsService(sink, consent, func() time.Time { return analyticsNow })
	err := service.Emit(context.Background(), "m-1", "epono.pod_heard", map[string]any{"durationPct": 80, "trustPathType": "ember"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEmitRejectedBeforeConsentOrSink(t *testing.T) {
	ctrl := gomock.NewController(t)
	sink := NewMockEventSink(ctrl)
	consent := NewMockConsentGate(ctrl)
	// No consent or sink expectations: invalid events fail first.

	service := NewAnalyticsService(sink, consent, func() time.Time { return analyticsNow })
	if err := service.Emit(context.Background(), "m-1", "unregistered.event", nil); err == nil {
		t.Fatal("unregistered event must fail at the boundary")
	}
}

func TestEmitOptedOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	sink := NewMockEventSink(ctrl)
	consent := NewMockConsentGate(ctrl)
	consent.EXPECT().AllowsAnalytics(gomock.Any(), "m-1").Return(false, nil)
	// No Append expectation.

	service := NewAnalyticsService(sink, consent, func() time.Time { return analyticsNow })
	err := service.Emit(context.Background(), "m-1", "gyaase.ember_converted", nil)
	if err != ErrAnalyticsOptedOut {
		t.Fatalf("Emit = %v, want opted out", err)
	}
}

func TestPseudonymStableAndDistinct(t *testing.T) {
	if Pseudonym("m-1") != Pseudonym("m-1") {
		t.Fatal("pseudonym must be stable")
	}
	if Pseudonym("m-1") == Pseudonym("m-2") {
		t.Fatal("distinct members must get distinct pseudonyms")
	}
	if len(Pseudonym("m-1")) != 32 {
		t.Fatalf("pseudonym length = %d", len(Pseudonym("m-1")))
	}
}
