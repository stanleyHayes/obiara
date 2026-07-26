package domain

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var now = time.Date(2026, 7, 26, 22, 0, 0, 0, time.UTC)

const a = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const b = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const ref = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

func TestDoubleEntryConservationAndClassLaw(t *testing.T) {
	p, e := NewPosting("posting:1", "command:1", ref, PurposeSaleSettlement, CurrencyGHS, []Line{{a, ClassAsset, SideDebit, 500}, {b, ClassRevenue, SideCredit, 500}}, now)
	if e != nil {
		t.Fatal(e)
	}
	asset, e := RecomputeBalance(a, ClassAsset, CurrencyGHS, []BookedLine{{p.ID(), p.Currency(), p.Lines()[0]}})
	if e != nil || asset != 500 {
		t.Fatalf("asset=%d err=%v", asset, e)
	}
	revenue, e := RecomputeBalance(b, ClassRevenue, CurrencyGHS, []BookedLine{{p.ID(), p.Currency(), p.Lines()[1]}})
	if e != nil || revenue != 500 {
		t.Fatalf("revenue=%d err=%v", revenue, e)
	}
}
func TestEveryAcceptedPostingConservesProperty(t *testing.T) {
	r := rand.New(rand.NewSource(840))
	for i := 0; i < 1000; i++ {
		amount := r.Int63n(MaxMinor) + 1
		p, e := NewPosting(fmt.Sprintf("posting:%d", i), "command", ref, PurposeCatalogReceivable, CurrencyGHS, []Line{{a, ClassAsset, SideDebit, amount}, {b, ClassLiability, SideCredit, amount}}, now)
		if e != nil || len(p.Lines()) != 2 {
			t.Fatalf("case %d amount=%d err=%v", i, amount, e)
		}
	}
}
func TestRejectsSingleEntryUnbalancedAndMemberTransferPurpose(t *testing.T) {
	tests := [][]Line{{{a, ClassAsset, SideDebit, 10}}, {{a, ClassAsset, SideDebit, 10}, {b, ClassRevenue, SideCredit, 9}}}
	for _, lines := range tests {
		if _, e := NewPosting("posting:x", "command:x", ref, PurposeSaleSettlement, CurrencyGHS, lines, now); e == nil {
			t.Fatal("invalid posting accepted")
		}
	}
	if _, e := NewPosting("posting:x", "command:x", ref, Purpose("member_transfer"), CurrencyGHS, []Line{{a, ClassAsset, SideDebit, 10}, {b, ClassRevenue, SideCredit, 10}}, now); e == nil {
		t.Fatal("member transfer accepted")
	}
}
func FuzzBalancedPostingAmount(f *testing.F) {
	f.Add(int64(1))
	f.Add(int64(-1))
	f.Add(MaxMinor + 1)
	f.Fuzz(func(t *testing.T, n int64) {
		_, e := NewPosting("posting:fuzz", "command:fuzz", ref, PurposeRefundSettlement, CurrencyUSD, []Line{{a, ClassExpense, SideDebit, n}, {b, ClassAsset, SideCredit, n}}, now)
		valid := n > 0 && n <= MaxMinor
		if valid != (e == nil) {
			t.Fatalf("n=%d valid=%v err=%v", n, valid, e)
		}
	})
}
