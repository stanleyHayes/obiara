package fireauthority

import (
	"context"
	"errors"
	"testing"
	"time"

	firedomain "github.com/stanleyHayes/obiara/services/api/internal/fire/domain"
)

type stubFires struct {
	fire firedomain.Fire
	err  error
}

func (stub stubFires) FindByID(context.Context, string) (firedomain.Fire, error) {
	return stub.fire, stub.err
}

func fireHostedBy(t *testing.T, hostID string) firedomain.Fire {
	t.Helper()
	fire, err := firedomain.NewFire("fire_1", hostID, "circle_1", "An evening",
		time.Now().Add(time.Hour), 10, time.Now())
	if err != nil {
		t.Fatalf("NewFire: %v", err)
	}
	return fire
}

func TestHostMayRunTheirFire(t *testing.T) {
	authority := New(stubFires{fire: fireHostedBy(t, "mem_host")})
	if err := authority.RequireHostOrCohost(context.Background(), "fire_1", "mem_host"); err != nil {
		t.Fatalf("host was refused: %v", err)
	}
}

// TestNonHostIsRefused is the point: a run sheet drives what everyone at a
// fire sees, so only the person running it may touch it.
func TestNonHostIsRefused(t *testing.T) {
	authority := New(stubFires{fire: fireHostedBy(t, "mem_host")})
	if err := authority.RequireHostOrCohost(context.Background(), "fire_1", "mem_guest"); !errors.Is(err, ErrNotHost) {
		t.Fatalf("error = %v, want ErrNotHost", err)
	}
}

// TestUnreadableFireIsRefused keeps a missing fire from reading as an open
// one, and stops the check being used to discover which fires exist.
func TestUnreadableFireIsRefused(t *testing.T) {
	authority := New(stubFires{err: errors.New("no documents")})
	if err := authority.RequireHostOrCohost(context.Background(), "fire_1", "mem_host"); !errors.Is(err, ErrNotHost) {
		t.Fatalf("error = %v, want ErrNotHost", err)
	}
}

// TestEmptyHostGrantsNobody covers a fire whose host is somehow blank: an
// absent host must not match an absent member.
func TestEmptyHostGrantsNobody(t *testing.T) {
	authority := New(stubFires{fire: firedomain.ReconstituteFire(
		"fire_1", "", "circle_1", "An evening", time.Now(), 10, 0,
		firedomain.StatusScheduled, 1, time.Now())})
	if err := authority.RequireHostOrCohost(context.Background(), "fire_1", ""); !errors.Is(err, ErrNotHost) {
		t.Fatalf("error = %v, want ErrNotHost", err)
	}
}
