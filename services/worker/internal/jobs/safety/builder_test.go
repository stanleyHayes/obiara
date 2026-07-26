package safety

import (
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/safety/application"
)

// The happy path runs in the Testcontainers integration suite (the builder
// depends on the concrete inbox store). Unit coverage here exercises error
// propagation through the mockable ports.
func TestRunOncePropagatesEventReadFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	events := NewMockEventLister(ctrl)
	reports := NewMockReportReader(ctrl)

	events.EXPECT().FindByEventType(gomock.Any(), "safety.report_filed", 50).
		Return(nil, context.DeadlineExceeded)

	builder := NewCaseBuilder(events, reports, application.CaseService{}, nil)
	if err := builder.RunOnce(context.Background(), 50); err == nil {
		t.Fatal("RunOnce must propagate the read failure")
	}
}

func TestReportPayloadShape(t *testing.T) {
	payload, err := json.Marshal(map[string]string{"reportId": "rep_1", "tier": "A", "category": "fraud"})
	if err != nil {
		t.Fatal(err)
	}
	var decoded reportPayload
	if err := json.Unmarshal(payload, &decoded); err != nil || decoded.ReportID != "rep_1" {
		t.Fatalf("payload = %v, %v", decoded, err)
	}
}
