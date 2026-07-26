package domain

import (
	"fmt"
	"testing"
	"time"
)

func k(n int) string { return fmt.Sprintf("%064x", n) }

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func TestExplicitConfirmationAndProviderConfirmation(t *testing.T) {
	i, e := Create(k(1), k(2), k(3), 1250, "create", now)
	if e != nil || i.State().Status != AwaitingMemberConfirmation {
		t.Fatal(e)
	}
	if _, e = i.ApplyProvider("callback", k(5), true, now); e == nil {
		t.Fatal("success before confirmation")
	}
	i, e = i.Confirm("confirm", now)
	if e != nil {
		t.Fatal(e)
	}
	i, e = i.MarkRequested("request", k(4), now)
	if e != nil {
		t.Fatal(e)
	}
	i, e = i.ApplyProvider("callback", k(5), true, now)
	if e != nil || i.State().Status != Succeeded {
		t.Fatal(e)
	}
	events := i.State().Events
	events[0].Status = Failed
	if i.State().Events[0].Status == Failed {
		t.Fatal("mutable audit")
	}
}
func TestAmountCannotMutate(t *testing.T) {
	i, _ := Create(k(1), k(2), k(3), 1250, "create", now)
	i, _ = i.Confirm("confirm", now)
	i, _ = i.MarkRequested("request", k(4), now)
	if i.State().AmountPesewas != 1250 {
		t.Fatal("amount mutated")
	}
}
func FuzzAmountBounds(f *testing.F) {
	f.Add(uint64(1))
	f.Add(MaxAmountPesewas)
	f.Fuzz(func(t *testing.T, a uint64) {
		i, e := Create(k(1), k(2), k(3), a, "create", now)
		valid := a > 0 && a <= MaxAmountPesewas
		if valid != (e == nil) {
			t.Fatalf("%d %v", a, e)
		}
		if e == nil && i.State().AmountPesewas != a {
			t.Fatal("amount changed")
		}
	})
}
