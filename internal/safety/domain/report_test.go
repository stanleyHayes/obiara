package domain

import (
	"strings"
	"testing"
	"time"
)

var safetyNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestNewReportValidation(t *testing.T) {
	if _, err := NewReport("r-1", " ", "m-2", CategoryFraud, SurfaceRoom, "", "", safetyNow); err != ErrReporterRequired {
		t.Fatalf("missing reporter = %v", err)
	}
	if _, err := NewReport("r-1", "m-1", " ", CategoryFraud, SurfaceRoom, "", "", safetyNow); err != ErrSubjectRequired {
		t.Fatalf("missing subject = %v", err)
	}
	if _, err := NewReport("r-1", "m-1", "m-1", CategoryFraud, SurfaceRoom, "", "", safetyNow); err != ErrSelfReport {
		t.Fatalf("self report = %v", err)
	}
	if _, err := NewReport("r-1", "m-1", "m-2", Category("nonsense"), SurfaceRoom, "", "", safetyNow); err != ErrInvalidCategory {
		t.Fatalf("bad category = %v", err)
	}
	if _, err := NewReport("r-1", "m-1", "m-2", CategoryFraud, Surface("sky"), "", "", safetyNow); err != ErrInvalidSurface {
		t.Fatalf("bad surface = %v", err)
	}
	if _, err := NewReport("r-1", "m-1", "m-2", CategoryFraud, SurfaceRoom, "", strings.Repeat("x", 501), safetyNow); err != ErrReasonTooLong {
		t.Fatalf("long reason = %v", err)
	}
}

func TestTierMapping(t *testing.T) {
	for category, want := range map[Category]Tier{
		CategoryFraud:         TierA,
		CategoryMinorSafety:   TierA,
		CategorySexualContent: TierA,
		CategoryHarassment:    TierB,
		CategorySpam:          TierC,
		CategoryOther:         TierC,
	} {
		if got := TierFor(category); got != want {
			t.Fatalf("TierFor(%s) = %s, want %s", category, got, want)
		}
	}
}

func TestReportAcknowledgementIsReporterSafe(t *testing.T) {
	report, err := NewReport("r-1", "m-1", "m-2", CategoryHarassment, SurfacePod, "pod_1", "kept pushing after decline", safetyNow)
	if err != nil {
		t.Fatal(err)
	}
	id, tier, createdAt := report.Acknowledgement()
	if id != "r-1" || tier != TierB || !createdAt.Equal(safetyNow) {
		t.Fatalf("ack = %q %s %v", id, tier, createdAt)
	}
	if report.Status() != StatusReceived {
		t.Fatalf("status = %q", report.Status())
	}
}

func TestBlockValidation(t *testing.T) {
	if _, err := NewBlock("m-1", "m-1", safetyNow); err != ErrSelfReport {
		t.Fatalf("self block = %v", err)
	}
	if _, err := NewBlock("m-1", " ", safetyNow); err != ErrSelfReport {
		t.Fatalf("missing target = %v", err)
	}
	block, err := NewBlock("m-1", "m-2", safetyNow)
	if err != nil || block.BlockerID() != "m-1" || block.BlockedID() != "m-2" {
		t.Fatalf("block = %#v, %v", block, err)
	}
}
