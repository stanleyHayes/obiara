package domain

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var now = time.Date(2026, 7, 26, 21, 0, 0, 0, time.UTC)

const title = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func cmd(id string, r uint64) Command { return Command{ID: id, ExpectedRevision: r, At: now} }
func sku(t *testing.T) SKU {
	t.Helper()
	p, _ := NewPrice(CurrencyGHS, 2500)
	s, e := Create("sku:1", "event.entry", title, 1, KindEventTicket, p, cmd("create", 0))
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestPublishedPriceIsImmutableAndNewPriceRequiresVersion(t *testing.T) {
	s, e := sku(t).Publish(cmd("publish", 1))
	if e != nil {
		t.Fatal(e)
	}
	p, _ := NewPrice(CurrencyGHS, 3000)
	next, e := NextVersion(s, "sku:2", p, cmd("next", 0))
	if e != nil || next.Version() != 2 || next.Price() != p || s.Price() == p {
		t.Fatalf("next=%+v err=%v", next, e)
	}
	if _, e = s.Publish(cmd("publish-again", 2)); e == nil {
		t.Fatal("published sku mutated")
	}
}
func TestProductLawPriceProperty(t *testing.T) {
	r := rand.New(rand.NewSource(834))
	for i := 0; i < 1000; i++ {
		minor := r.Int63n(MaxPriceMinor*2) - MaxPriceMinor/2
		p, e := NewPrice(CurrencyGHS, minor)
		valid := minor >= 1 && minor <= MaxPriceMinor
		if valid != (e == nil) {
			t.Fatalf("case %d minor=%d err=%v price=%+v", i, minor, e, p)
		}
	}
}
func TestOnlyAllowlistedKinds(t *testing.T) {
	p, _ := NewPrice(CurrencyGHS, 1)
	for _, k := range []Kind{KindPhysicalGood, KindEventTicket, KindDigitalService} {
		if _, e := Create("sku:"+string(k), "catalog.item", title, 1, k, p, cmd("create:"+string(k), 0)); e != nil {
			t.Fatal(e)
		}
	}
	for _, bad := range []Kind{"seed", "approach", "visibility", "rank", "matching_advantage", "suban", "trust", "urgency", "member_transfer"} {
		if _, e := Create("bad:"+string(bad), "catalog.item", title, 1, bad, p, cmd("bad:"+string(bad), 0)); e == nil {
			t.Fatalf("%s accepted", bad)
		}
	}
}
func FuzzPriceBounds(f *testing.F) {
	f.Add(int64(1))
	f.Add(int64(0))
	f.Add(MaxPriceMinor + 1)
	f.Fuzz(func(t *testing.T, n int64) {
		_, e := NewPrice(CurrencyUSD, n)
		valid := n >= 1 && n <= MaxPriceMinor
		if valid != (e == nil) {
			t.Fatalf("n=%d valid=%v err=%v", n, valid, e)
		}
	})
}

var _ = fmt.Sprintf
