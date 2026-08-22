package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	routingapp "github.com/stanleyHayes/obiara/internal/notifications/routing/application"
	routingdomain "github.com/stanleyHayes/obiara/internal/notifications/routing/domain"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
)

// ritualEventPrefix is the event-type namespace the ritual dispatcher emits.
const ritualEventPrefix = "notification.ritual."

// Router is the channel-ladder port the publisher delivers through.
type Router interface {
	Deliver(context.Context, routingapp.Outbound) (routingapp.Result, error)
}

// NotifyingPublisher delivers outbox records that represent member-facing
// notifications through the routing ladder.
//
// It replaces the logging placeholder that stood here while consumers were
// unbuilt. That placeholder returned nil for every record, so the relay
// marked each one published and the event was gone: ritual notifications
// were dispatched by the scheduler, written durably to the outbox, and then
// silently discarded without ever reaching a member.
//
// Records this publisher does not route are acknowledged rather than
// failed. Some event types are consumed by their own jobs reading the
// outbox by event type (the safety case builder is one), and blocking the
// relay on them would stall every other record behind a head-of-line
// failure.
type NotifyingPublisher struct {
	router Router
	logger *slog.Logger
}

func NewNotifyingPublisher(router Router, logger *slog.Logger) NotifyingPublisher {
	return NotifyingPublisher{router: router, logger: logger}
}

// ritualPayload is the shape the ritual dispatcher writes.
type ritualPayload struct {
	MemberID      string `json:"memberId"`
	Kind          string `json:"kind"`
	Discriminator string `json:"discriminator"`
}

// Publish routes one record. An error leaves the record unpublished so the
// relay retries it and, past its attempt budget, dead-letters it.
func (publisher NotifyingPublisher) Publish(ctx context.Context, record outbox.Record) error {
	if !strings.HasPrefix(record.EventType, ritualEventPrefix) {
		// Not ours to route. Acknowledge so the relay keeps draining.
		publisher.log(ctx, slog.LevelDebug, "outbox record not routed", record, nil)
		return nil
	}

	var payload ritualPayload
	if err := json.Unmarshal(record.Payload, &payload); err != nil {
		// A record whose payload cannot be parsed will never parse. Retrying
		// it would block the queue forever, so it is acknowledged loudly and
		// left for triage.
		publisher.log(ctx, slog.LevelError, "ritual event payload could not be decoded", record, err)
		return nil
	}
	if payload.MemberID == "" {
		publisher.log(ctx, slog.LevelError, "ritual event has no member", record, nil)
		return nil
	}

	// The reference is what the in-app inbox renders from, and the inbox
	// rejects an empty one. The ritual kind is always present and is a
	// meaningful reference on its own when there is no discriminator.
	reference := payload.Discriminator
	if reference == "" {
		reference = strings.TrimPrefix(record.EventType, ritualEventPrefix)
	}
	if reference == "" {
		publisher.log(ctx, slog.LevelError, "ritual event has no reference", record, nil)
		return nil
	}

	result, err := publisher.router.Deliver(ctx, routingapp.Outbound{
		MemberID:  payload.MemberID,
		Purpose:   routingdomain.PurposeRitual,
		Template:  strings.TrimPrefix(record.EventType, ritualEventPrefix),
		Reference: reference,
	})
	if err != nil {
		if errors.Is(err, routingapp.ErrAllChannelsFailed) {
			return fmt.Errorf("deliver ritual %q for member: %w", record.EventType, err)
		}
		return fmt.Errorf("route ritual %q: %w", record.EventType, err)
	}

	// An empty channel means the preference boundary suppressed the
	// notification. That is a successful outcome, not a delivery.
	if result.Channel == "" {
		publisher.log(ctx, slog.LevelDebug, "ritual suppressed by member preferences", record, nil)
		return nil
	}
	publisher.logger.InfoContext(ctx, "ritual delivered",
		slog.String("eventType", record.EventType),
		slog.String("channel", string(result.Channel)))
	return nil
}

// log emits a record-scoped line carrying only bounded identifiers. Member
// ids and payloads stay out of the logs.
func (publisher NotifyingPublisher) log(ctx context.Context, level slog.Level, message string, record outbox.Record, err error) {
	attrs := []any{
		slog.String("eventType", record.EventType),
		slog.String("aggregateType", record.AggregateType),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	publisher.logger.Log(ctx, level, message, attrs...)
}
