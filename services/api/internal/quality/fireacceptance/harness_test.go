package fireacceptance

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func TestE09S10Deterministic150SeatConstrained3GAcceptance(t *testing.T) {
	harness := New(DeterministicTransport{})
	first, err := harness.Run(context.Background(), AcceptanceSeats, ReferenceDevices(), Constrained3G())
	if err != nil {
		t.Fatalf("acceptance run failed: %v", err)
	}
	second, err := harness.Run(context.Background(), AcceptanceSeats, ReferenceDevices(), Constrained3G())
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same inputs did not produce identical evidence")
	}
	if first.Seats != 150 || first.ListenOnlySeats != 150 {
		t.Fatalf("seat/degradation proof = %+v", first)
	}
	if first.UnavailableControls != 0 {
		t.Fatalf("unavailable safety controls = %d", first.UnavailableControls)
	}
	if first.P90ControlLatency > ControlLatencyLimit {
		t.Fatalf("p90 control latency = %s", first.P90ControlLatency)
	}
	if len(first.Matrix) != 5 {
		t.Fatalf("device matrix size = %d", len(first.Matrix))
	}
	t.Logf(
		"evidence seats=%d listen_only=%d unavailable_controls=%d control_p90=%s matrix=%+v",
		first.Seats,
		first.ListenOnlySeats,
		first.UnavailableControls,
		first.P90ControlLatency,
		first.Matrix,
	)
	for _, result := range first.Matrix {
		if result.Seats != 30 || result.ListenOnlySeats != 30 || result.UnavailableControls != 0 {
			t.Fatalf("device result = %+v", result)
		}
	}
}

func TestHarnessFailsClosedWhenAControlIsUnavailable(t *testing.T) {
	controller := gomock.NewController(t)
	transport := NewMockProbeTransport(controller)
	transport.EXPECT().Probe(gomock.Any(), gomock.Any()).Return(Observation{
		Mode: ModeListenOnly,
		ControlLatency: map[Control]time.Duration{
			ControlSafety:       200 * time.Millisecond,
			ControlLeave:        200 * time.Millisecond,
			ControlConsentOptIn: 200 * time.Millisecond,
		},
	}, nil).Times(AcceptanceSeats)

	report, err := New(transport).Run(context.Background(), AcceptanceSeats, ReferenceDevices(), Constrained3G())
	if err == nil || report.UnavailableControls != AcceptanceSeats {
		t.Fatalf("expected unavailable-control failure, report=%+v err=%v", report, err)
	}
}

func TestHarnessStopsOnProbeFailure(t *testing.T) {
	controller := gomock.NewController(t)
	transport := NewMockProbeTransport(controller)
	transport.EXPECT().Probe(gomock.Any(), gomock.Any()).Return(Observation{}, errors.New("synthetic transport failure"))

	_, err := New(transport).Run(context.Background(), AcceptanceSeats, ReferenceDevices(), Constrained3G())
	if err == nil {
		t.Fatal("expected transport failure")
	}
}

func TestHarnessRejectsAnythingExceptTheAcceptanceCapacity(t *testing.T) {
	_, err := New(DeterministicTransport{}).Run(context.Background(), AcceptanceSeats-1, ReferenceDevices(), Constrained3G())
	if err == nil {
		t.Fatal("149 seats must not satisfy the 150-seat gate")
	}
}
