package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

var safetyNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

type outboxStub struct {
	appended []outbox.Record
	err      error
}

func (stub *outboxStub) Append(_ context.Context, record outbox.Record) error {
	stub.appended = append(stub.appended, record)
	return stub.err
}

func TestFilePersistsAndEmits(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMockReportRepository(ctrl)
	blocks := NewMockBlockRepository(ctrl)
	events := &outboxStub{}

	reports.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, report domain.Report) error {
			if report.ID() != "rep_test" || report.Tier() != domain.TierA || report.Status() != domain.StatusReceived {
				t.Fatalf("report = %#v", report)
			}
			return nil
		})

	service := NewSafetyService(reports, blocks, events, func() time.Time { return safetyNow }, func() string { return "rep_test" })
	id, tier, err := service.File(context.Background(), "m-1", "m-2", domain.CategoryFraud, domain.SurfaceRoom, "room_1", "asked for money")
	if err != nil {
		t.Fatal(err)
	}
	if id != "rep_test" || tier != domain.TierA {
		t.Fatalf("ack = %q %s", id, tier)
	}
	if len(events.appended) != 1 || events.appended[0].EventType != "safety.report_filed" {
		t.Fatalf("events = %#v", events.appended)
	}
}

func TestFileRejectsInvalidWithoutPersistence(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMockReportRepository(ctrl)
	blocks := NewMockBlockRepository(ctrl)
	events := &outboxStub{}
	// No Create expectation: invalid input must not persist or emit.

	service := NewSafetyService(reports, blocks, events, func() time.Time { return safetyNow }, func() string { return "rep_test" })
	if _, _, err := service.File(context.Background(), "m-1", "m-1", domain.CategoryFraud, domain.SurfaceRoom, "", ""); err != domain.ErrSelfReport {
		t.Fatalf("File = %v, want ErrSelfReport", err)
	}
	if len(events.appended) != 0 {
		t.Fatal("invalid report emitted an event")
	}
}

func TestBlockLifecycle(t *testing.T) {
	ctrl := gomock.NewController(t)
	blocks := NewMockBlockRepository(ctrl)
	service := NewSafetyService(nil, blocks, nil, func() time.Time { return safetyNow }, func() string { return "rep_test" })

	blocks.EXPECT().Add(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, block domain.Block) error {
			if block.BlockerID() != "m-1" || block.BlockedID() != "m-2" {
				t.Fatalf("block = %#v", block)
			}
			return nil
		})
	if err := service.Block(context.Background(), "m-1", "m-2"); err != nil {
		t.Fatal(err)
	}

	blocks.EXPECT().Remove(gomock.Any(), "m-1", "m-2").Return(nil)
	if err := service.Unblock(context.Background(), "m-1", "m-2"); err != nil {
		t.Fatal(err)
	}

	blocks.EXPECT().Exists(gomock.Any(), "m-1", "m-2").Return(false, nil)
	exists, err := service.IsBlocked(context.Background(), "m-1", "m-2")
	if err != nil || exists {
		t.Fatalf("exists = %v, %v", exists, err)
	}
}
