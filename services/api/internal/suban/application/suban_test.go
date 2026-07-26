package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
)

var subanSvcNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestRecordEnforcesPeriodCap(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockEventStore(ctrl)
	service := NewSubanService(events, func() time.Time { return subanSvcNow }, func() string { return "sub_test" })

	events.EXPECT().CountForSubjectSince(gomock.Any(), "m-1", domain.KindKindClosure, gomock.Any()).Return(domain.PeriodCap, nil)
	if err := service.Record(context.Background(), "m-1", domain.KindKindClosure, domain.Provenance{Source: "closure", Ref: "r-1"}); err != ErrPeriodCapReached {
		t.Fatalf("Record = %v, want cap", err)
	}

	events.EXPECT().CountForSubjectSince(gomock.Any(), "m-1", domain.KindKindClosure, gomock.Any()).Return(domain.PeriodCap-1, nil)
	events.EXPECT().Append(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, event domain.Event) error {
			if event.ID != "sub_test" || event.Provenance.Ref != "r-1" {
				t.Fatalf("event = %#v", event)
			}
			return nil
		})
	if err := service.Record(context.Background(), "m-1", domain.KindKindClosure, domain.Provenance{Source: "closure", Ref: "r-1"}); err != nil {
		t.Fatal(err)
	}
}

func TestMarksRecomputeFromLedger(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockEventStore(ctrl)
	service := NewSubanService(events, func() time.Time { return subanSvcNow }, func() string { return "sub_test" })

	ledger := []domain.Event{
		{ID: "e1", SubjectID: "m-1", Kind: domain.KindMeetingFollowThrough, OccurredAt: subanSvcNow.Add(-time.Hour)},
		{ID: "e2", SubjectID: "m-1", Kind: domain.KindMeetingFollowThrough, OccurredAt: subanSvcNow.Add(-2 * time.Hour)},
		{ID: "e3", SubjectID: "m-1", Kind: domain.KindMeetingFollowThrough, OccurredAt: subanSvcNow.Add(-3 * time.Hour)},
	}
	events.EXPECT().ListForSubject(gomock.Any(), "m-1").Return(ledger, nil)

	marks, err := service.Marks(context.Background(), "m-1")
	if err != nil || len(marks) != 1 || marks[0] != domain.MarkKeepsWord {
		t.Fatalf("marks = %v, %v", marks, err)
	}
}
