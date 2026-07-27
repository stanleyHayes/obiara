package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc/domain"
)

var scamNow = time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

func newService(t *testing.T) (ScamArcService, *MockSignalStore, *MockMonitoringConsent, *MockCaseOpener) {
	t.Helper()
	ctrl := gomock.NewController(t)
	signals := NewMockSignalStore(ctrl)
	consent := NewMockMonitoringConsent(ctrl)
	cases := NewMockCaseOpener(ctrl)
	return NewScamArcService(signals, consent, cases, func() time.Time { return scamNow }, func() string { return "sig_test" }), signals, consent, cases
}

func TestObserveOptedOutRoom(t *testing.T) {
	service, _, consent, _ := newService(t)
	consent.EXPECT().MonitoringAllowed(gomock.Any(), "room_1").Return(false, nil)
	// No Append expectation.

	if _, _, err := service.Observe(context.Background(), "room_1", "m-1", domain.SignalAskPattern); err != ErrMonitoringOptedOut {
		t.Fatalf("Observe = %v, want opted out", err)
	}
}

func TestEducationCardOnRungCrossing(t *testing.T) {
	service, signals, consent, _ := newService(t)
	consent.EXPECT().MonitoringAllowed(gomock.Any(), "room_1").Return(true, nil)
	signals.EXPECT().Append(gomock.Any(), gomock.Any()).Return(nil)
	// affection(1) + ask(3) × 1.25 = 5.0 → education rung.
	signals.EXPECT().KindsForRoom(gomock.Any(), "room_1").Return(
		[]domain.SignalKind{domain.SignalAffectionCadence, domain.SignalAskPattern}, nil)
	signals.EXPECT().StateForRoom(gomock.Any(), "room_1").Return(domain.RoomState{RoomID: "room_1", Ladder: domain.LadderWatch}, nil)
	signals.EXPECT().SaveState(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, state domain.RoomState) error {
			if state.Ladder != domain.LadderEducation {
				t.Fatalf("state = %#v", state)
			}
			return nil
		})

	state, card, err := service.Observe(context.Background(), "room_1", "m-1", domain.SignalAskPattern)
	if err != nil {
		t.Fatal(err)
	}
	if state.Ladder != domain.LadderEducation {
		t.Fatalf("ladder = %v", state.Ladder)
	}
	if card == nil || card.ContentKey != educationContentKey {
		t.Fatalf("card = %#v", card)
	}
}

func TestNoCardWhenAlreadyAtRung(t *testing.T) {
	service, signals, consent, _ := newService(t)
	consent.EXPECT().MonitoringAllowed(gomock.Any(), "room_1").Return(true, nil)
	signals.EXPECT().Append(gomock.Any(), gomock.Any()).Return(nil)
	signals.EXPECT().KindsForRoom(gomock.Any(), "room_1").Return(
		[]domain.SignalKind{domain.SignalAffectionCadence, domain.SignalAskPattern}, nil)
	signals.EXPECT().StateForRoom(gomock.Any(), "room_1").Return(domain.RoomState{RoomID: "room_1", Ladder: domain.LadderEducation}, nil)
	signals.EXPECT().SaveState(gomock.Any(), gomock.Any()).Return(nil)

	_, card, err := service.Observe(context.Background(), "room_1", "m-1", domain.SignalAskPattern)
	if err != nil {
		t.Fatal(err)
	}
	if card != nil {
		t.Fatal("already at education rung: no new card")
	}
}

func TestCaseOpensAtTopRung(t *testing.T) {
	service, signals, consent, cases := newService(t)
	consent.EXPECT().MonitoringAllowed(gomock.Any(), "room_1").Return(true, nil)
	signals.EXPECT().Append(gomock.Any(), gomock.Any()).Return(nil)
	// affection(1) + emergency(2) + pull(2.5) × 1.5 = 8.25 → case.
	signals.EXPECT().KindsForRoom(gomock.Any(), "room_1").Return(
		[]domain.SignalKind{domain.SignalAffectionCadence, domain.SignalEmergencyNarrative, domain.SignalOffPlatformPull}, nil)
	signals.EXPECT().StateForRoom(gomock.Any(), "room_1").Return(domain.RoomState{RoomID: "room_1", Ladder: domain.LadderFriction}, nil)
	signals.EXPECT().SaveState(gomock.Any(), gomock.Any()).Return(nil)
	cases.EXPECT().OpenScamCase(gomock.Any(), "room_1", "m-1", 8.25).Return(nil)

	state, _, err := service.Observe(context.Background(), "room_1", "m-1", domain.SignalOffPlatformPull)
	if err != nil {
		t.Fatal(err)
	}
	if state.Ladder != domain.LadderCase {
		t.Fatalf("ladder = %v, want case", state.Ladder)
	}
}

func TestCaseNotReopened(t *testing.T) {
	service, signals, consent, _ := newService(t)
	consent.EXPECT().MonitoringAllowed(gomock.Any(), "room_1").Return(true, nil)
	signals.EXPECT().Append(gomock.Any(), gomock.Any()).Return(nil)
	signals.EXPECT().KindsForRoom(gomock.Any(), "room_1").Return(
		[]domain.SignalKind{domain.SignalAffectionCadence, domain.SignalEmergencyNarrative, domain.SignalOffPlatformPull}, nil)
	signals.EXPECT().StateForRoom(gomock.Any(), "room_1").Return(domain.RoomState{RoomID: "room_1", Ladder: domain.LadderCase}, nil)
	signals.EXPECT().SaveState(gomock.Any(), gomock.Any()).Return(nil)
	// No OpenScamCase expectation.

	if _, _, err := service.Observe(context.Background(), "room_1", "m-1", domain.SignalOffPlatformPull); err != nil {
		t.Fatal(err)
	}
}
