package domain

import (
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func quote(now time.Time, c Currency, amount int64) Quote {
	return Quote{key(1), 2, c, amount, now.Add(time.Hour)}
}
func TestCheckoutRequiresDiasporaCurrencyAndConfirmationBeforeAccounting(t *testing.T) {
	now := time.Now().UTC()
	if _, e := Create(key(2), key(3), Quote{key(1), 1, Currency("GHS"), 100, now.Add(time.Hour)}, key(4), "prepare-1", now); e == nil {
		t.Fatal("accepted GHS")
	}
	c, e := Create(key(2), key(3), quote(now, CurrencyGBP, 1499), key(4), "prepare-1", now)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.RecordLedger(key(7), "ledger-1", now); e == nil {
		t.Fatal("accounted before provider confirmation")
	}
	confirmed, e := c.Confirm(key(5), "confirm-1", true, now)
	if e != nil {
		t.Fatal(e)
	}
	accounted, e := confirmed.RecordLedger(key(6), "ledger-1", now)
	if e != nil || accounted.State().Status != Accounted {
		t.Fatal(e)
	}
	if _, e = accounted.RecordLedger(key(7), "ledger-2", now); e == nil {
		t.Fatal("double accounting accepted")
	}
}
func TestFailedCheckoutCanNeverAccount(t *testing.T) {
	now := time.Now().UTC()
	c, _ := Create(key(2), key(3), quote(now, CurrencyEUR, 1399), key(4), "prepare-1", now)
	failed, _ := c.Confirm(key(5), "confirm-1", false, now)
	if _, e := failed.RecordLedger(key(6), "ledger-1", now); e == nil {
		t.Fatal("failed checkout accounted")
	}
}
func TestMinorUnitBoundsProperty(t *testing.T) {
	now := time.Now().UTC()
	if e := quick.Check(func(amount uint64) bool {
		minor := int64(amount)
		_, err := Create(key(2), key(3), quote(now, CurrencyUSD, minor), key(4), "prepare-1", now)
		want := amount > 0 && amount <= uint64(MaxMinor)
		return (err == nil) == want
	}, nil); e != nil {
		t.Fatal(e)
	}
}
func FuzzUnsupportedCurrency(f *testing.F) {
	f.Add("GHS")
	f.Add("JPY")
	f.Fuzz(func(t *testing.T, currency string) {
		now := time.Now().UTC()
		_, e := Create(key(2), key(3), quote(now, Currency(currency), 100), key(4), "prepare-1", now)
		if e == nil && currency != "GBP" && currency != "USD" && currency != "EUR" {
			t.Fatal("unsupported currency accepted")
		}
	})
}
