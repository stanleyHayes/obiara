package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

var evidenceSvcNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestViewAuditsBeforeReading(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMockReportRepository(ctrl)
	audit := NewMockAccessAudit(ctrl)

	var order []string
	audit.EXPECT().Append(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, record domain.AccessRecord) error {
			order = append(order, "audit")
			if record.CaseID != "rep_1" || record.AgentID != "agent-1" || record.Purpose != domain.PurposeTriage {
				t.Fatalf("record = %#v", record)
			}
			return nil
		})
	reports.EXPECT().FindByID(gomock.Any(), "rep_1").DoAndReturn(
		func(_ context.Context, _ string) (domain.Report, error) {
			order = append(order, "read")
			return domain.ReconstituteReport("rep_1", "m-1", "m-2", domain.CategoryHarassment, domain.TierB, domain.SurfaceRoom, "room_1", "call +233550000101", domain.StatusReceived, 1, evidenceSvcNow), nil
		})

	service := NewEvidenceService(reports, audit, func() time.Time { return evidenceSvcNow }, func() string { return "acc_test" })
	bundle, err := service.View(context.Background(), "rep_1", "agent-1", domain.PurposeTriage)
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 2 || order[0] != "audit" || order[1] != "read" {
		t.Fatalf("order = %v, want audit before read", order)
	}
	if bundle.SubjectID != "m-2" {
		t.Fatalf("bundle = %#v", bundle)
	}
	if bundle.Description == "call +233550000101" {
		t.Fatal("bundle leaked the phone number")
	}
}

func TestViewRejectsCuriosity(t *testing.T) {
	ctrl := gomock.NewController(t)
	reports := NewMockReportRepository(ctrl)
	audit := NewMockAccessAudit(ctrl)
	// No Append or FindByID expectation.

	service := NewEvidenceService(reports, audit, func() time.Time { return evidenceSvcNow }, func() string { return "acc_test" })
	if _, err := service.View(context.Background(), "rep_1", "agent-1", domain.Purpose("curiosity")); err != domain.ErrInvalidPurpose {
		t.Fatalf("View = %v, want rejected", err)
	}
}
