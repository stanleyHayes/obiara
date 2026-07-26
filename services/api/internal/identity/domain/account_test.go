package domain

import (
	"testing"
)

func TestTierTransitionMatrix(t *testing.T) {
	cases := []struct {
		name    string
		from    Tier
		target  Tier
		wantErr error
	}{
		{"promote 0 to 1", TierUnverified, TierVerified, nil},
		{"promote 1 to 2", TierVerified, TierSowing, nil},
		{"demote 2 to 1", TierSowing, TierVerified, nil},
		{"demote 2 to 0", TierSowing, TierUnverified, nil},
		{"demote 1 to 0", TierVerified, TierUnverified, nil},
		{"skip 0 to 2 rejected", TierUnverified, TierSowing, ErrInvalidTierTransition},
		{"same tier rejected", TierVerified, TierVerified, ErrInvalidTierTransition},
		{"out of range rejected", TierUnverified, Tier(3), ErrInvalidTierTransition},
		{"negative rejected", TierVerified, Tier(-1), ErrInvalidTierTransition},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			account := ReconstituteAccount("id_1", "+233550000101", AccountActive, tc.from, 1, nil, testNow)
			transition, err := account.ApplyTransition(tc.target, "verification passed", "actor-1", testNow)
			if err != tc.wantErr {
				t.Fatalf("ApplyTransition = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil {
				if account.Tier() != tc.target {
					t.Fatalf("tier = %d, want %d", account.Tier(), tc.target)
				}
				if account.Version() != 2 {
					t.Fatalf("version = %d, want 2", account.Version())
				}
				if transition.From != tc.from || transition.To != tc.target || transition.ActorID != "actor-1" || transition.Reason == "" {
					t.Fatalf("transition = %#v", transition)
				}
			}
		})
	}
}

func TestTransitionRequiresReasonAndActor(t *testing.T) {
	account := ReconstituteAccount("id_1", "+233550000101", AccountActive, TierUnverified, 1, nil, testNow)
	if _, err := account.ApplyTransition(TierVerified, "  ", "actor-1", testNow); err != ErrTransitionReasonRequired {
		t.Fatalf("blank reason = %v", err)
	}
	if _, err := account.ApplyTransition(TierVerified, "ok", " ", testNow); err != ErrTransitionActorRequired {
		t.Fatalf("blank actor = %v", err)
	}
}

func TestBlockedAccountCannotTransition(t *testing.T) {
	account := ReconstituteAccount("id_1", "+233550000101", AccountBlocked, TierVerified, 1, nil, testNow)
	if _, err := account.ApplyTransition(TierSowing, "ok", "actor-1", testNow); err != ErrAccountNotUsable {
		t.Fatalf("blocked transition = %v, want ErrAccountNotUsable", err)
	}
}
