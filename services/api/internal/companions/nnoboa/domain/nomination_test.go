package domain

import (
	"testing"
	"time"
)

var testNow = time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

func validNomination(t *testing.T) Nomination {
	t.Helper()
	n, err := NewNomination("mem_12345678", "Auntie Efua", "+233550000101", "aunt", testNow)
	if err != nil {
		t.Fatalf("NewNomination: %v", err)
	}
	return n
}

func TestNewNominationValidates(t *testing.T) {
	cases := []struct {
		name                                      string
		memberID, kinName, kinPhone, relationship string
	}{
		{"bad member", "x", "Auntie Efua", "+233550000101", "aunt"},
		{"blank kin name", "mem_12345678", "  ", "+233550000101", "aunt"},
		{"bad phone", "mem_12345678", "Auntie Efua", "0550000101", "aunt"},
		{"bad relationship", "mem_12345678", "Auntie Efua", "+233550000101", "cousin"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewNomination(tc.memberID, tc.kinName, tc.kinPhone, tc.relationship, testNow); err != ErrInvalidNomination {
				t.Fatalf("want ErrInvalidNomination, got %v", err)
			}
		})
	}
	if _, err := NewNomination("mem_12345678", "Auntie Efua", "+233550000101", "aunt", time.Time{}); err != ErrInvalidNomination {
		t.Fatalf("zero time: want ErrInvalidNomination, got %v", err)
	}
}

func TestNewNominationPending(t *testing.T) {
	n := validNomination(t)
	if n.Status != StatusPending || n.Version != 1 || n.RespondedAt != nil {
		t.Fatalf("unexpected initial state: %+v", n)
	}
	if n.ID == "" || n.CreatedAt != testNow {
		t.Fatalf("unexpected identity: %+v", n)
	}
}

func TestConsentAndDeclineTransitions(t *testing.T) {
	n := validNomination(t)
	if err := n.Consent(testNow.Add(time.Hour)); err != nil {
		t.Fatalf("Consent: %v", err)
	}
	if n.Status != StatusConsented || n.RespondedAt == nil || n.Version != 2 {
		t.Fatalf("unexpected consented state: %+v", n)
	}
	if err := n.Decline(testNow.Add(2 * time.Hour)); err != ErrNotPending {
		t.Fatalf("decline after consent: want ErrNotPending, got %v", err)
	}

	m := validNomination(t)
	if err := m.Decline(testNow.Add(time.Hour)); err != nil {
		t.Fatalf("Decline: %v", err)
	}
	if m.Status != StatusDeclined {
		t.Fatalf("want declined, got %s", m.Status)
	}
	if err := m.Consent(testNow.Add(2 * time.Hour)); err != ErrNotPending {
		t.Fatalf("consent after decline: want ErrNotPending, got %v", err)
	}
}

func TestExpiryWindow(t *testing.T) {
	n := validNomination(t)
	if n.ExpiredAt(testNow.Add(NominationExpiry - time.Second)) {
		t.Fatal("should not be expired before the window")
	}
	if !n.ExpiredAt(testNow.Add(NominationExpiry)) {
		t.Fatal("should be expired at the window edge")
	}
	if err := n.Expire(testNow.Add(NominationExpiry - time.Second)); err != ErrInvalidNomination {
		t.Fatalf("early Expire: want ErrInvalidNomination, got %v", err)
	}
	if err := n.Expire(testNow.Add(NominationExpiry)); err != nil {
		t.Fatalf("Expire: %v", err)
	}
	if n.Status != StatusExpired {
		t.Fatalf("want expired, got %s", n.Status)
	}
	if n.ExpiredAt(testNow.Add(NominationExpiry + time.Hour)) {
		t.Fatal("non-pending nomination must not report expired")
	}
}

func TestEveryRelationshipAccepted(t *testing.T) {
	for _, rel := range []string{"aunt", "uncle", "mother", "father", "elder"} {
		if _, err := NewNomination("mem_12345678", "Kin", "+233550000101", rel, testNow); err != nil {
			t.Fatalf("relationship %s: %v", rel, err)
		}
	}
}
