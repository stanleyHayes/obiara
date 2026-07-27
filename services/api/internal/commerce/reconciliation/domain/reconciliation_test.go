package domain

import (
	"fmt"
	"testing"
	"time"
)

const (
	k1 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	k2 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	k3 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

func fact(t *testing.T, minor int64) StatementFact {
	t.Helper()
	f, err := NewFact("fact:1", k1, k2, k3, "ledger:1", CurrencyGHS, StatusSettled, minor, time.Unix(10, 0), time.Unix(11, 0))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestExactComparisonAndExplicitExceptions(t *testing.T) {
	f := fact(t, 500)
	good := LedgerProof{CommandID: "ledger:1", ReferenceKey: k3, Currency: CurrencyGHS, Minor: 500, Balanced: true}
	if d := Compare(f, good, true); d.Outcome() != OutcomeReconciled || d.Exception() != "" {
		t.Fatalf("%+v", d)
	}
	cases := []struct {
		name  string
		proof LedgerProof
		found bool
		code  ExceptionCode
	}{
		{"missing", LedgerProof{}, false, ExceptionLedgerMissing},
		{"reference", func() LedgerProof { p := good; p.ReferenceKey = k2; return p }(), true, ExceptionReference},
		{"currency", func() LedgerProof { p := good; p.Currency = CurrencyUSD; return p }(), true, ExceptionCurrency},
		{"one-minor", func() LedgerProof { p := good; p.Minor = 501; return p }(), true, ExceptionAmount},
		{"unbalanced", func() LedgerProof { p := good; p.Balanced = false; return p }(), true, ExceptionUnbalancedLedger},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := Compare(f, tc.proof, tc.found)
			if d.Outcome() != OutcomeException || d.Exception() != tc.code {
				t.Fatalf("got=%s/%s", d.Outcome(), d.Exception())
			}
		})
	}
}

func TestFailedStatementIsNotDueForLedgerMutation(t *testing.T) {
	f, err := NewFact("fact:1", k1, k2, k3, "ledger:1", CurrencyGHS, StatusFailed, 500, time.Unix(10, 0), time.Unix(11, 0))
	if err != nil {
		t.Fatal(err)
	}
	if d := Compare(f, LedgerProof{}, false); d.Outcome() != OutcomeNotDue {
		t.Fatalf("%+v", d)
	}
}

func TestComparisonPropertyRequiresExactAmount(t *testing.T) {
	f := fact(t, 5_000)
	for delta := int64(-1000); delta <= 1000; delta++ {
		p := LedgerProof{CommandID: "ledger:1", ReferenceKey: k3, Currency: CurrencyGHS, Minor: 5_000 + delta, Balanced: true}
		d := Compare(f, p, true)
		if (d.Outcome() == OutcomeReconciled) != (delta == 0) {
			t.Fatalf("delta=%d outcome=%s", delta, d.Outcome())
		}
	}
}

func FuzzFactMinorBounds(f *testing.F) {
	f.Add(int64(1))
	f.Add(MaxMinor)
	f.Add(int64(0))
	f.Add(MaxMinor + 1)
	f.Fuzz(func(t *testing.T, minor int64) {
		_, err := NewFact("fact:1", k1, k2, k3, "ledger:1", CurrencyGHS, StatusSettled, minor, time.Unix(10, 0), time.Unix(11, 0))
		valid := minor > 0 && minor <= MaxMinor
		if (err == nil) != valid {
			t.Fatalf("minor=%s err=%v", fmt.Sprint(minor), err)
		}
	})
}
