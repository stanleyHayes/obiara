package domain

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var now = time.Date(2026, 7, 26, 19, 0, 0, 0, time.UTC)

const subject = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const hold = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const proof = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

func policy(t *testing.T) Policy {
	t.Helper()
	p, e := NewPolicy("messages.metadata", "safety.audit", 2, 30*24*time.Hour, now.Add(-time.Hour))
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func command(id string, r uint64) Command { return Command{ID: id, ExpectedRevision: r, At: now} }
func record(t *testing.T) Record {
	t.Helper()
	r, e := Create("retention:1", subject, policy(t), command("create", 0))
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestLegalHoldPrecedesErasure(t *testing.T) {
	r, e := record(t).PlaceHold(hold, command("hold", 1))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = r.RequestErasure(command("erase", 2)); !errors.Is(e, ErrHeld) {
		t.Fatalf("request=%v", e)
	}
	r, e = r.ReleaseHold(hold, command("release", 2))
	if e != nil {
		t.Fatal(e)
	}
	r, e = r.RequestErasure(command("request", 3))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = r.CompleteErasure("", command("complete", 4)); e == nil {
		t.Fatal("unverified erasure completed")
	}
	r, e = r.CompleteErasure(proof, command("complete", 4))
	if e != nil || r.Status() != StatusErased || r.Counters().VerifiedErasures != 1 {
		t.Fatalf("record=%+v err=%v", r, e)
	}
}
func TestEveryHeldRecordRejectsErasureProperty(t *testing.T) {
	rnd := rand.New(rand.NewSource(825))
	for i := 0; i < 1000; i++ {
		r, _ := Create(fmt.Sprintf("retention:%d", i+2), subject, policy(t), Command{ID: fmt.Sprintf("create:%d", i), At: now})
		r, _ = r.PlaceHold(hold, Command{ID: fmt.Sprintf("hold:%d", i), ExpectedRevision: 1, At: now.Add(time.Duration(rnd.Intn(100)) * time.Second)})
		if _, e := r.RequestErasure(Command{ID: fmt.Sprintf("erase:%d", i), ExpectedRevision: 2, At: now.Add(time.Minute)}); !errors.Is(e, ErrHeld) {
			t.Fatalf("case %d: %v", i, e)
		}
	}
}
func FuzzPolicyRejectsUnboundedRetention(f *testing.F) {
	f.Add(int64(24 * time.Hour))
	f.Add(int64(-1))
	f.Fuzz(func(t *testing.T, n int64) {
		d := time.Duration(n)
		_, e := NewPolicy("messages.metadata", "safety.audit", 1, d, now)
		valid := d >= 24*time.Hour && d <= 10*365*24*time.Hour
		if valid != (e == nil) {
			t.Fatalf("duration=%v valid=%v err=%v", d, valid, e)
		}
	})
}
