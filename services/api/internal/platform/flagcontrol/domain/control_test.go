package domain

import (
	"errors"
	"testing"
	"time"
)

const (
	actorA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	actorB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func proposal(t *testing.T, action Action) Proposal {
	t.Helper()
	p, err := NewProposal("proposal:1", "command:1", actorA, CapabilityPayments, EnvironmentProduction, MarketGH, action, ReasonIncident, time.Unix(10, 0), time.Unix(10, 0).Add(MaxLifetime))
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func TestDistinctApprovalApplyAndExpiryFailClosed(t *testing.T) {
	p := proposal(t, ActionEnable)
	if _, err := p.Approve(actorA, time.Unix(11, 0)); !errors.Is(err, ErrSameActor) {
		t.Fatalf("same actor=%v", err)
	}
	approved, err := p.Approve(actorB, time.Unix(11, 0))
	if err != nil {
		t.Fatal(err)
	}
	applied, change, err := approved.Apply(time.Unix(12, 0))
	if err != nil || !change.Enabled || change.Killed {
		t.Fatalf("change=%+v err=%v", change, err)
	}
	expired, fallback, err := applied.Expire(time.Unix(10, 0).Add(MaxLifetime))
	if err != nil || fallback.Enabled || !fallback.Killed || expired.Status() != StatusExpired {
		t.Fatalf("fallback=%+v status=%s err=%v", fallback, expired.Status(), err)
	}
	if _, err = Rehydrate(expired.State()); err != nil {
		t.Fatal(err)
	}
}

func TestAppliedChangeCanBeReplayedOnlyBeforeExpiry(t *testing.T) {
	proposed := proposal(t, ActionKill)
	approved, _ := proposed.Approve(actorB, time.Unix(11, 0))
	applied, _, _ := approved.Apply(time.Unix(12, 0))
	change, err := applied.AppliedChange(time.Unix(13, 0))
	if err != nil || change.Enabled || !change.Killed {
		t.Fatalf("change=%+v err=%v", change, err)
	}
	if _, err := approved.AppliedChange(time.Unix(13, 0)); !errors.Is(err, ErrState) {
		t.Fatalf("approved replay err=%v", err)
	}
	if _, err := applied.AppliedChange(time.Unix(10, 0).Add(MaxLifetime)); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired replay err=%v", err)
	}
}
func TestEveryCapabilityActionAndScopeIsBoundedProperty(t *testing.T) {
	caps := []Capability{CapabilitySow, CapabilityFires, CapabilityAI, CapabilityPayments, CapabilityGate}
	actions := []Action{ActionEnable, ActionDisable, ActionKill, ActionUnkill}
	envs := []Environment{EnvironmentStaging, EnvironmentProduction}
	for _, capability := range caps {
		for _, action := range actions {
			for _, env := range envs {
				if _, err := NewProposal("proposal:1", "command:1", actorA, capability, env, MarketGH, action, ReasonStagedRollout, time.Unix(1, 0), time.Unix(2, 0)); err != nil {
					t.Fatalf("%s/%s/%s: %v", capability, action, env, err)
				}
			}
		}
	}
}
func TestUnknownOrGlobalScopeRejected(t *testing.T) {
	for _, market := range []Market{"", "GLOBAL", "NG"} {
		if _, err := NewProposal("proposal:1", "command:1", actorA, CapabilitySow, EnvironmentProduction, market, ActionKill, ReasonIncident, time.Unix(1, 0), time.Unix(2, 0)); err == nil {
			t.Fatalf("market %q accepted", market)
		}
	}
}
func FuzzLifetimeBound(f *testing.F) {
	f.Add(int64(1))
	f.Add(int64(MaxLifetime))
	f.Add(int64(MaxLifetime + time.Nanosecond))
	f.Add(int64(0))
	f.Fuzz(func(t *testing.T, nanos int64) {
		start := time.Unix(1, 0)
		_, err := NewProposal("proposal:1", "command:1", actorA, CapabilitySow, EnvironmentProduction, MarketGH, ActionEnable, ReasonStagedRollout, start, start.Add(time.Duration(nanos)))
		valid := nanos > 0 && time.Duration(nanos) <= MaxLifetime
		if (err == nil) != valid {
			t.Fatalf("nanos=%d err=%v", nanos, err)
		}
	})
}
