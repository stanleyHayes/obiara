package domain

import (
	"testing"
)

func TestCloseToEmbersHostOnly(t *testing.T) {
	fire, err := NewFire("fire_1", "host-1", "circle-1", "Friday Fire", fireStart, 50, fireNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := fire.CloseToEmbers("m-random", fireNow); err != ErrNotHost {
		t.Fatalf("non-host close = %v, want ErrNotHost (FR-401)", err)
	}
	if err := fire.CloseToEmbers("host-1", fireNow); err != nil {
		t.Fatal(err)
	}
	if fire.Status() != StatusEmbers {
		t.Fatalf("status = %q, want embers", fire.Status())
	}
}

func TestEmbersStopsAdmissions(t *testing.T) {
	fire, _ := NewFire("fire_1", "host-1", "circle-1", "Friday Fire", fireStart, 50, fireNow)
	if err := fire.CloseToEmbers("host-1", fireNow); err != nil {
		t.Fatal(err)
	}
	if _, err := fire.Admit("m-1", 0, fireNow); err != ErrFireNotOpen {
		t.Fatalf("admission after close = %v, want ErrFireNotOpen", err)
	}
}

func TestCloseRequiresOpenState(t *testing.T) {
	for _, status := range []FireStatus{StatusEnded, StatusCancelled, StatusEmbers} {
		fire := ReconstituteFire("fire_1", "host-1", "circle-1", "F", fireStart, 10, 0, status, 1, fireNow)
		if err := fire.CloseToEmbers("host-1", fireNow); err != ErrFireNotClosable {
			t.Fatalf("close from %s = %v, want not closable", status, err)
		}
	}
	live := ReconstituteFire("fire_2", "host-1", "circle-1", "F", fireStart, 10, 0, StatusLive, 1, fireNow)
	if err := live.CloseToEmbers("host-1", fireNow); err != nil {
		t.Fatalf("live close = %v", err)
	}
}
