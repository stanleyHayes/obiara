package relay

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/platform/outbox"
)

type storeStub struct {
	pending        []outbox.Record
	published      []string
	failedAttempts []string
	pendingErr     error
}

func (stub *storeStub) Pending(context.Context, int) ([]outbox.Record, error) {
	return stub.pending, stub.pendingErr
}

func (stub *storeStub) MarkPublished(_ context.Context, id string) error {
	stub.published = append(stub.published, id)
	return nil
}

func (stub *storeStub) MarkAttemptFailed(_ context.Context, id string) error {
	stub.failedAttempts = append(stub.failedAttempts, id)
	return nil
}

func record(id string) outbox.Record {
	return outbox.Record{
		ID:            id,
		AggregateType: "member",
		AggregateID:   "member-1",
		EventType:     "member.registered",
		Payload:       []byte(`{}`),
		OccurredAt:    time.Now(),
	}
}

func TestNewOutboxJobShape(t *testing.T) {
	ctrl := gomock.NewController(t)
	publisher := NewMockPublisher(ctrl)
	job := NewOutboxJob(&storeStub{}, publisher, 100, 5*time.Second)
	if job.Name != "outbox.relay" {
		t.Fatalf("name = %q", job.Name)
	}
	if job.Interval != 5*time.Second || job.Timeout != 30*time.Second || job.MaxAttempts != 5 {
		t.Fatalf("job shape = %+v", job)
	}
}

func TestRelayPublishesPendingAndMarksPublished(t *testing.T) {
	ctrl := gomock.NewController(t)
	publisher := NewMockPublisher(ctrl)
	publisher.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	store := &storeStub{pending: []outbox.Record{record("evt-1"), record("evt-2")}}
	job := NewOutboxJob(store, publisher, 100, time.Second)
	if err := job.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(store.published) != 2 || store.published[0] != "evt-1" || store.published[1] != "evt-2" {
		t.Fatalf("published = %v", store.published)
	}
	if len(store.failedAttempts) != 0 {
		t.Fatalf("failedAttempts = %v", store.failedAttempts)
	}
}

func TestRelayFailureMarksAttemptAndStops(t *testing.T) {
	ctrl := gomock.NewController(t)
	publisher := NewMockPublisher(ctrl)
	publisher.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(errors.New("provider down"))

	store := &storeStub{pending: []outbox.Record{record("evt-1"), record("evt-2")}}
	job := NewOutboxJob(store, publisher, 100, time.Second)
	if err := job.Run(context.Background()); err == nil {
		t.Fatal("Run must report the publish failure")
	}
	if len(store.failedAttempts) != 1 || store.failedAttempts[0] != "evt-1" {
		t.Fatalf("failedAttempts = %v", store.failedAttempts)
	}
	if len(store.published) != 0 {
		t.Fatal("no record may be marked published after a failure")
	}
}

func TestRelayPropagatesPendingReadFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	publisher := NewMockPublisher(ctrl)
	store := &storeStub{pendingErr: errors.New("mongo down")}
	job := NewOutboxJob(store, publisher, 100, time.Second)
	if err := job.Run(context.Background()); err == nil {
		t.Fatal("Run must report the read failure")
	}
}
