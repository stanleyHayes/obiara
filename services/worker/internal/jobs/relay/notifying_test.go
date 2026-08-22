package relay

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	routingapp "github.com/stanleyHayes/obiara/internal/notifications/routing/application"
	routingdomain "github.com/stanleyHayes/obiara/internal/notifications/routing/domain"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
)

type stubRouter struct {
	delivered []routingapp.Outbound
	result    routingapp.Result
	err       error
}

func (router *stubRouter) Deliver(_ context.Context, outbound routingapp.Outbound) (routingapp.Result, error) {
	router.delivered = append(router.delivered, outbound)
	return router.result, router.err
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func ritualRecord(t *testing.T, memberID, kind, discriminator string) outbox.Record {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"memberId": memberID, "kind": kind, "discriminator": discriminator,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return outbox.Record{
		ID:            "ritual_" + kind + "_" + memberID,
		AggregateType: "member",
		AggregateID:   memberID,
		EventType:     "notification.ritual." + kind,
		Payload:       payload,
		OccurredAt:    time.Date(2026, 8, 22, 6, 0, 0, 0, time.UTC),
	}
}

// TestRitualEventIsRouted is the regression this publisher exists for: the
// placeholder returned nil for every record, so the relay marked ritual
// events published and they never reached a member.
func TestRitualEventIsRouted(t *testing.T) {
	router := &stubRouter{result: routingapp.Result{Channel: routingdomain.ChannelInApp, ProviderRef: "inapp:1"}}
	publisher := NewNotifyingPublisher(router, quietLogger())

	record := ritualRecord(t, "mem_1", "morning_greeting", "2026-08-22")
	if err := publisher.Publish(context.Background(), record); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	if len(router.delivered) != 1 {
		t.Fatalf("router received %d deliveries, want 1", len(router.delivered))
	}
	got := router.delivered[0]
	if got.MemberID != "mem_1" {
		t.Errorf("MemberID = %q", got.MemberID)
	}
	if got.Purpose != routingdomain.PurposeRitual {
		t.Errorf("Purpose = %q, want ritual so the ladder and preference category are right", got.Purpose)
	}
	if got.Template != "morning_greeting" {
		t.Errorf("Template = %q", got.Template)
	}
	if got.Reference != "2026-08-22" {
		t.Errorf("Reference = %q", got.Reference)
	}
}

// TestFailedDeliveryIsRetried keeps a dark channel from silently consuming
// the event: an unpublished record is retried and eventually dead-lettered.
func TestFailedDeliveryIsRetried(t *testing.T) {
	router := &stubRouter{err: routingapp.ErrAllChannelsFailed}
	publisher := NewNotifyingPublisher(router, quietLogger())

	err := publisher.Publish(context.Background(), ritualRecord(t, "mem_1", "evening_check", "2026-08-22"))
	if err == nil {
		t.Fatal("Publish succeeded with every channel failing; the record would be marked published and lost")
	}
	if !errors.Is(err, routingapp.ErrAllChannelsFailed) {
		t.Errorf("error %v should wrap ErrAllChannelsFailed", err)
	}
}

func TestRouterErrorIsRetried(t *testing.T) {
	router := &stubRouter{err: errors.New("mongodb unavailable")}
	publisher := NewNotifyingPublisher(router, quietLogger())

	if err := publisher.Publish(context.Background(), ritualRecord(t, "mem_1", "evening_check", "d")); err == nil {
		t.Fatal("Publish swallowed a router error")
	}
}

// TestSuppressedByPreferencesIsSuccess covers the E13-S01 boundary: the
// router returns an empty channel with no error when preferences suppress a
// notification. That is a decision, not a failure, and must not be retried.
func TestSuppressedByPreferencesIsSuccess(t *testing.T) {
	router := &stubRouter{result: routingapp.Result{}}
	publisher := NewNotifyingPublisher(router, quietLogger())

	if err := publisher.Publish(context.Background(), ritualRecord(t, "mem_1", "morning_greeting", "d")); err != nil {
		t.Fatalf("Publish = %v; a suppressed notification must not be retried forever", err)
	}
}

// TestForeignEventTypesAreAcknowledged stops one unroutable record from
// blocking every record behind it. safety.report_filed has its own consumer
// reading the outbox by event type.
func TestForeignEventTypesAreAcknowledged(t *testing.T) {
	router := &stubRouter{}
	publisher := NewNotifyingPublisher(router, quietLogger())

	record := outbox.Record{
		ID: "report_1", AggregateType: "report", AggregateID: "rep_1",
		EventType: "safety.report_filed", Payload: []byte(`{"reportId":"rep_1"}`),
	}
	if err := publisher.Publish(context.Background(), record); err != nil {
		t.Fatalf("Publish = %v, want nil so the relay keeps draining", err)
	}
	if len(router.delivered) != 0 {
		t.Errorf("router received %d deliveries for a foreign event", len(router.delivered))
	}
}

// TestPoisonPayloadsAreAcknowledged stops a permanently malformed record
// from becoming a head-of-line block that stalls the whole queue.
func TestPoisonPayloadsAreAcknowledged(t *testing.T) {
	cases := map[string][]byte{
		"not json":     []byte(`{`),
		"no member":    []byte(`{"kind":"morning_greeting","discriminator":"d"}`),
		"empty object": []byte(`{}`),
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			router := &stubRouter{}
			publisher := NewNotifyingPublisher(router, quietLogger())
			record := outbox.Record{
				ID: "ritual_x", AggregateType: "member",
				EventType: "notification.ritual.morning_greeting", Payload: payload,
			}
			if err := publisher.Publish(context.Background(), record); err != nil {
				t.Fatalf("Publish = %v; a record that can never parse must not block the queue", err)
			}
			if len(router.delivered) != 0 {
				t.Errorf("router received a delivery for an undecodable record")
			}
		})
	}
}

// TestMissingDiscriminatorFallsBackToTheKind keeps the in-app inbox from
// rejecting the notification, since it requires a non-empty reference.
func TestMissingDiscriminatorFallsBackToTheKind(t *testing.T) {
	router := &stubRouter{result: routingapp.Result{Channel: routingdomain.ChannelInApp}}
	publisher := NewNotifyingPublisher(router, quietLogger())

	if err := publisher.Publish(context.Background(), ritualRecord(t, "mem_1", "morning_greeting", "")); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if len(router.delivered) != 1 {
		t.Fatalf("router received %d deliveries, want 1", len(router.delivered))
	}
	if got := router.delivered[0].Reference; got != "morning_greeting" {
		t.Errorf("Reference = %q, want the ritual kind as a fallback", got)
	}
}
