package domain

import (
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestRibbonRequiresTwoCurrentGrantsAndRevokesImmediately(t *testing.T) {
	r, _ := New("honesty-room-01", []string{key(1), key(2)})
	now := time.Now().UTC()
	r, _ = r.Grant(Command{"grant-command-01", key(1), 0, now})
	if r.Visible() {
		t.Fatal("one grant visible")
	}
	r, _ = r.Grant(Command{"grant-command-02", key(2), 1, now})
	if !r.Visible() {
		t.Fatal("mutual grants not visible")
	}
	r, _ = r.Revoke(Command{"revoke-command-01", key(1), 2, now})
	if r.Visible() {
		t.Fatal("revoked ribbon visible")
	}
}
