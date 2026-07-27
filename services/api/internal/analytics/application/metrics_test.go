package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

var metricsNow = time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

func TestFunnelComputesRates(t *testing.T) {
	ctrl := gomock.NewController(t)
	counts := NewMockEventCounts(ctrl)
	cohorts := NewMockActiveCohorts(ctrl)

	counts.EXPECT().CountEvents(gomock.Any(), "epono.seed_sown", gomock.Any()).Return(100, nil)
	counts.EXPECT().CountEvents(gomock.Any(), "epono.pod_heard", gomock.Any()).Return(70, nil)
	counts.EXPECT().CountEvents(gomock.Any(), "epono.sprout_opened", gomock.Any()).Return(30, nil)
	counts.EXPECT().CountEvents(gomock.Any(), "epono.room_opened", gomock.Any()).Return(12, nil)
	counts.EXPECT().CountDistinctSubjects(gomock.Any(), "gyaase.fire_attended", gomock.Any()).Return(45, nil)
	cohorts.EXPECT().CountActive(gomock.Any()).Return(100, nil)
	counts.EXPECT().CountEvents(gomock.Any(), "wellbeing.regret_reported", gomock.Any()).Return(3, nil)
	counts.EXPECT().CountEvents(gomock.Any(), "wellbeing.regret_reported", gomock.Any()).Return(8, nil)

	service := NewMetricsService(counts, cohorts, func() time.Time { return metricsNow })
	report, err := service.Funnel(context.Background(), 30)
	if err != nil {
		t.Fatal(err)
	}
	if report.PodsHeardRate != 0.7 {
		t.Fatalf("podsHeard = %v, want 0.7", report.PodsHeardRate)
	}
	if report.SeedToSproutRate != 0.3 {
		t.Fatalf("seedToSprout = %v, want 0.3", report.SeedToSproutRate)
	}
	if report.SproutToRoomRate != 0.4 {
		t.Fatalf("sproutToRoom = %v, want 0.4", report.SproutToRoomRate)
	}
	if report.FireAttendanceRate != 0.45 {
		t.Fatalf("fireAttendance = %v, want 0.45", report.FireAttendanceRate)
	}
	// Current window 3 regrets; prior window 8-3=5 → trending down.
	if report.RegretTrend != "down" {
		t.Fatalf("trend = %q, want down", report.RegretTrend)
	}
}

func TestZeroDenominatorsYieldZeroRates(t *testing.T) {
	ctrl := gomock.NewController(t)
	counts := NewMockEventCounts(ctrl)
	cohorts := NewMockActiveCohorts(ctrl)

	counts.EXPECT().CountEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(0, nil).AnyTimes()
	counts.EXPECT().CountDistinctSubjects(gomock.Any(), gomock.Any(), gomock.Any()).Return(0, nil)
	cohorts.EXPECT().CountActive(gomock.Any()).Return(0, nil)

	service := NewMetricsService(counts, cohorts, func() time.Time { return metricsNow })
	report, err := service.Funnel(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if report.PodsHeardRate != 0 || report.FireAttendanceRate != 0 {
		t.Fatalf("report = %#v, want zero rates without division errors", report)
	}
	if report.WindowDays != 30 {
		t.Fatalf("windowDays = %d, want default 30", report.WindowDays)
	}
}
