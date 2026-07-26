package domain

import (
	"testing"
	"time"
)

var fireNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
var fireStart = fireNow.Add(6 * time.Hour)

func TestNewFireValidation(t *testing.T) {
	if _, err := NewFire("", "host-1", "circle-1", "Friday Fire", fireStart, 50, fireNow); err != ErrFireIDRequired {
		t.Fatalf("missing id = %v", err)
	}
	if _, err := NewFire("fire_1", " ", "circle-1", "Friday Fire", fireStart, 50, fireNow); err != ErrHostRequired {
		t.Fatalf("missing host = %v", err)
	}
	if _, err := NewFire("fire_1", "host-1", "circle-1", " ", fireStart, 50, fireNow); err != ErrTitleRequired {
		t.Fatalf("missing title = %v", err)
	}
	for _, capacity := range []int{0, -5, MaxCapacity + 1} {
		if _, err := NewFire("fire_1", "host-1", "circle-1", "Friday Fire", fireStart, capacity, fireNow); err != ErrInvalidCapacity {
			t.Fatalf("capacity %d = %v", capacity, err)
		}
	}
	if _, err := NewFire("fire_1", "host-1", "circle-1", "Friday Fire", time.Time{}, 50, fireNow); err != ErrInvalidStart {
		t.Fatalf("zero start = %v", err)
	}
}

func TestAdmitUpToCapacityThenWaitlist(t *testing.T) {
	fire, err := NewFire("fire_1", "host-1", "circle-1", "Friday Fire", fireStart, 2, fireNow)
	if err != nil {
		t.Fatal(err)
	}

	first, err := fire.Admit("m-1", 0, fireNow)
	if err != nil || first.Status() != RSVPGoing {
		t.Fatalf("first = %#v, %v", first, err)
	}
	second, err := fire.Admit("m-2", 0, fireNow)
	if err != nil || second.Status() != RSVPGoing {
		t.Fatalf("second = %#v, %v", second, err)
	}
	third, err := fire.Admit("m-3", 0, fireNow)
	if err != nil || third.Status() != RSVPWaitlisted || third.Position() != 1 {
		t.Fatalf("third = %#v, %v", third, err)
	}
	fourth, err := fire.Admit("m-4", 1, fireNow)
	if err != nil || fourth.Status() != RSVPWaitlisted || fourth.Position() != 2 {
		t.Fatalf("fourth = %#v, %v", fourth, err)
	}
	if fire.GoingCount() != 2 {
		t.Fatalf("going = %d, want 2", fire.GoingCount())
	}
}

func TestClosedFireRejectsAdmission(t *testing.T) {
	fire := ReconstituteFire("fire_1", "host-1", "circle-1", "F", fireStart, 10, 0, StatusEnded, 1, fireNow)
	if _, err := fire.Admit("m-1", 0, fireNow); err != ErrFireNotOpen {
		t.Fatalf("ended fire = %v, want not open", err)
	}
	cancelled := ReconstituteFire("fire_2", "host-1", "circle-1", "F", fireStart, 10, 0, StatusCancelled, 1, fireNow)
	if _, err := cancelled.Admit("m-1", 0, fireNow); err != ErrFireNotOpen {
		t.Fatalf("cancelled fire = %v, want not open", err)
	}
}

func TestReleaseAndPromote(t *testing.T) {
	fire, _ := NewFire("fire_1", "host-1", "circle-1", "F", fireStart, 1, fireNow)
	going, _ := fire.Admit("m-1", 0, fireNow)
	waiting, _ := fire.Admit("m-2", 0, fireNow)
	if going.Status() != RSVPGoing || waiting.Status() != RSVPWaitlisted {
		t.Fatal("setup failed")
	}

	fire.Release()
	if fire.GoingCount() != 0 {
		t.Fatalf("after release going = %d", fire.GoingCount())
	}
	fire.Promote(&waiting, fireNow)
	if waiting.Status() != RSVPGoing || waiting.Position() != 0 {
		t.Fatalf("promoted = %#v", waiting)
	}
	if fire.GoingCount() != 1 {
		t.Fatalf("after promote going = %d", fire.GoingCount())
	}
}
