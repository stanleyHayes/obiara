package domain

import (
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

var now = time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)

func fixture(t *testing.T) Escrow {
	t.Helper()
	x, e := Fund(key(1), key(2), key(3), key(4), 3000, Terms{ID: "terms.1", Version: 2, Milestones: []MilestoneTerm{{ID: "first", GrossPesewas: 1000, FeePesewas: 100}, {ID: "second", GrossPesewas: 2000, FeePesewas: 200}}}, "fund-1", now)
	if e != nil {
		t.Fatal(e)
	}
	return x
}
func TestDualEvidenceBoundedSettlementAndExplicitFee(t *testing.T) {
	x := fixture(t)
	if _, _, e := x.Settle("first", key(5), "settle-1", now); e == nil {
		t.Fatal("released without evidence")
	}
	x, _ = x.AddEvidence("first", DeliveryEvidence, key(3), "delivery-1", now)
	if _, _, e := x.Settle("first", key(5), "settle-1", now); e == nil {
		t.Fatal("released unilaterally")
	}
	x, _ = x.AddEvidence("first", AcceptanceEvidence, key(4), "accept-1", now)
	x, statement, e := x.Settle("first", key(5), "settle-1", now)
	if e != nil || statement.GrossPesewas != 1000 || statement.FeePesewas != 100 || statement.NetPesewas != 900 || x.State().SettledPesewas != 1000 {
		t.Fatal(statement, e)
	}
	if x.State().FundedPesewas != 3000 || x.State().TermsVersion != 2 {
		t.Fatal("terms mutated")
	}
}
func TestDisputePermanentlyFreezesAndKeepsEscalation(t *testing.T) {
	x := fixture(t)
	x, e := x.RaiseDispute(key(8), key(9), "dispute-1", now)
	if e != nil {
		t.Fatal(e)
	}
	if x.State().Dispute.EscalationRef != key(9) {
		t.Fatal("missing escalation")
	}
	if _, e = x.AddEvidence("first", DeliveryEvidence, key(3), "delivery-1", now); e == nil {
		t.Fatal("dispute not frozen")
	}
}
func FuzzTermsMustExactlyEqualFunding(f *testing.F) {
	f.Add(uint64(1), uint64(2))
	f.Fuzz(func(t *testing.T, a, b uint64) {
		if a > 1000000 || b > 1000000 {
			t.Skip()
		}
		_, e := Fund(key(1), key(2), key(3), key(4), a+b, Terms{ID: "terms.1", Version: 1, Milestones: []MilestoneTerm{{ID: "one", GrossPesewas: a}, {ID: "two", GrossPesewas: b}}}, "fund-1", now)
		valid := a > 0 && b > 0 && a+b >= a
		if valid != (e == nil) {
			t.Fatalf("%d %d %v", a, b, e)
		}
	})
}
