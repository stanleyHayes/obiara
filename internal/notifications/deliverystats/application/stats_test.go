package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

var statsNow = time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

func TestStatsAggregatesChannels(t *testing.T) {
	ctrl := gomock.NewController(t)
	counts := NewMockDeliveryCounts(ctrl)

	counts.EXPECT().CountByStatus(gomock.Any(), "whatsapp_deliveries", gomock.Any()).Return(
		map[string]int{"sent": 8, "failed": 2}, nil)
	counts.EXPECT().CountByStatus(gomock.Any(), "email_deliveries", gomock.Any()).Return(
		map[string]int{"sent": 5, "delivered": 3, "bounced": 1, "complained": 1}, nil)

	service := NewStatsService(counts, func() time.Time { return statsNow })
	report, err := service.Stats(context.Background(), 30)
	if err != nil {
		t.Fatal(err)
	}

	whatsapp := report.Channels["whatsapp"]
	if whatsapp.Attempted != 10 || whatsapp.Failed != 2 || whatsapp.SuccessRate != 0.8 {
		t.Fatalf("whatsapp = %#v", whatsapp)
	}
	email := report.Channels["email"]
	if email.Attempted != 10 || email.Failed != 2 || email.SuccessRate != 0.8 {
		t.Fatalf("email = %#v", email)
	}
}

func TestEmptyChannelsYieldZeroRate(t *testing.T) {
	ctrl := gomock.NewController(t)
	counts := NewMockDeliveryCounts(ctrl)
	counts.EXPECT().CountByStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]int{}, nil).Times(2)

	service := NewStatsService(counts, func() time.Time { return statsNow })
	report, err := service.Stats(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if report.Channels["whatsapp"].SuccessRate != 0 {
		t.Fatalf("report = %#v, want zero rate", report)
	}
	if report.WindowDays != 30 {
		t.Fatalf("windowDays = %d, want default 30", report.WindowDays)
	}
}
