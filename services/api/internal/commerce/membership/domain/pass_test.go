package domain

import (
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func grant(t *testing.T, grace time.Duration) Pass {
	t.Helper()
	pass, err := New(Grant{
		ID: key(1), MemberKey: key(2), PassID: "obiara.pass", PassVersion: 3,
		ReceiptRef: key(3), GrantedAt: now, PaidThrough: now.Add(30 * 24 * time.Hour),
		GraceUntil: now.Add(30 * 24 * time.Hour).Add(grace), GraceDuration: grace,
	}, "grant-1")
	if err != nil {
		t.Fatal(err)
	}
	return pass
}

func TestCancellationIsNonPunitiveAndDoesNotMutateTerm(t *testing.T) {
	pass := grant(t, 3*24*time.Hour)
	before := pass.State().Grant
	cancelled, err := pass.Cancel("cancel-1", now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status(now.Add(20*24*time.Hour)) != StatusActive {
		t.Fatal("cancel revoked paid access")
	}
	if cancelled.Status(now.Add(31*24*time.Hour)) != StatusGrace {
		t.Fatal("cancel removed grace")
	}
	if cancelled.State().Grant != before {
		t.Fatal("historical term mutated")
	}
}

func TestRefundRequiresCancellationRequestAndProviderConfirmation(t *testing.T) {
	pass := grant(t, 0)
	if _, err := pass.ConfirmRefund("confirm-1", key(4), key(5), now); err == nil {
		t.Fatal("refund confirmed before request")
	}
	pass, _ = pass.Cancel("cancel-1", now.Add(time.Hour))
	pass, _ = pass.RequestRefund("refund-1", key(4), now.Add(2*time.Hour))
	if pass.Status(now) != StatusRefundPending {
		t.Fatal("not pending")
	}
	pass, err := pass.ConfirmRefund("confirm-1", key(4), key(5), now.Add(3*time.Hour))
	if err != nil || pass.Status(now) != StatusRefunded {
		t.Fatal(err)
	}
	events := pass.State().Events
	events[0].Kind = EventRefundConfirmed
	if pass.State().Events[0].Kind != EventGranted {
		t.Fatal("audit mutable")
	}
}

func TestGraceIsBounded(t *testing.T) {
	if _, err := New(Grant{ID: key(1), MemberKey: key(2), PassID: "obiara.pass", PassVersion: 1, ReceiptRef: key(3), GrantedAt: now, PaidThrough: now.Add(time.Hour), GraceUntil: now.Add(time.Hour + MaxGrace + time.Second), GraceDuration: MaxGrace + time.Second}, "grant-1"); err == nil {
		t.Fatal("unbounded grace")
	}
}

func FuzzGraceNeverExceedsBound(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(MaxGrace))
	f.Add(int64(MaxGrace + 1))
	f.Fuzz(func(t *testing.T, nanos int64) {
		grace := time.Duration(nanos)
		_, err := New(Grant{ID: key(1), MemberKey: key(2), PassID: "obiara.pass", PassVersion: 1, ReceiptRef: key(3), GrantedAt: now, PaidThrough: now.Add(time.Hour), GraceUntil: now.Add(time.Hour).Add(grace), GraceDuration: grace}, "grant-1")
		valid := grace >= 0 && grace <= MaxGrace
		if valid != (err == nil) {
			t.Fatalf("grace=%v err=%v", grace, err)
		}
	})
}
