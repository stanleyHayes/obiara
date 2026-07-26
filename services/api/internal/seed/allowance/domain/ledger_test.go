package domain

import (
	"errors"
	"testing"
	"time"
)

func TestWeekPolicyUsesMondayMidnightInConfiguredTimezone(t *testing.T) {
	policy, err := NewWeekPolicy("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	before := time.Date(2026, 3, 9, 3, 59, 59, 0, time.UTC)
	after := before.Add(time.Second)
	if got, want := policy.Start(before), time.Date(2026, 3, 2, 5, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("before boundary = %s, want %s", got, want)
	}
	if got, want := policy.Start(after), time.Date(2026, 3, 9, 4, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("after DST boundary = %s, want %s", got, want)
	}
}

func TestSpendRenewsOnceAndNeverOverdraws(t *testing.T) {
	now := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	ledger, err := Issue("subject-key", 5, now, now, "issue-1", "fp-issue")
	if err != nil {
		t.Fatal(err)
	}
	ledger, _, err = ledger.Spend(4, now, now, "spend-1", "fp-spend-1")
	if err != nil {
		t.Fatal(err)
	}
	nextWeek := now.AddDate(0, 0, 7)
	ledger, changed, err := ledger.Spend(3, nextWeek, nextWeek, "spend-2", "fp-spend-2")
	if err != nil || !changed {
		t.Fatalf("spend after renewal: changed=%v err=%v", changed, err)
	}
	if ledger.Balance() != 2 {
		t.Fatalf("balance=%d", ledger.Balance())
	}
	entries := ledger.Entries()
	if len(entries) != 4 || entries[2].Kind != EntryRenewal || entries[3].Kind != EntrySpend {
		t.Fatalf("unexpected audit entries: %#v", entries)
	}
	unchanged, changed, err := ledger.Spend(3, nextWeek, nextWeek, "spend-3", "fp-spend-3")
	if !errors.Is(err, ErrInsufficient) || changed || unchanged.Balance() != 2 {
		t.Fatalf("overdraw was not blocked: %#v %v", unchanged, err)
	}
}

func TestCommandFingerprintMakesReplaySafe(t *testing.T) {
	now := time.Now().UTC()
	ledger, _ := Issue("subject-key", 10, now, now, "issue", "issue-fp")
	spent, _, _ := ledger.Spend(2, now, now, "command", "same")
	replayed, changed, err := spent.Spend(2, now, now, "command", "same")
	if err != nil || changed || replayed.Version() != spent.Version() {
		t.Fatalf("replay changed ledger: %v", err)
	}
	_, _, err = spent.Spend(3, now, now, "command", "different")
	if !errors.Is(err, ErrCommandConflict) {
		t.Fatalf("got %v", err)
	}
}

func FuzzSpendNeverCreatesNegativeBalance(f *testing.F) {
	f.Add(int64(10), int64(3))
	f.Add(int64(1), int64(2))
	f.Fuzz(func(t *testing.T, allowance, spend int64) {
		if allowance <= 0 || allowance > 1_000_000 {
			t.Skip()
		}
		now := time.Unix(0, 0).UTC()
		ledger, err := Issue("key", allowance, now, now, "issue", "fp")
		if err != nil {
			t.Fatal(err)
		}
		next, _, err := ledger.Spend(spend, now, now, "spend", "spend-fp")
		if err == nil && next.Balance() < 0 {
			t.Fatalf("negative balance %d", next.Balance())
		}
	})
}
