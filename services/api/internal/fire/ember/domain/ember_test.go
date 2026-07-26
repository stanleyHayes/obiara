package domain

import (
	"testing"
	"time"
)

var emberNow = time.Date(2026, time.July, 26, 22, 0, 0, 0, time.UTC)

func TestNewEmberValidation(t *testing.T) {
	if _, err := NewEmber("", "fire_1", "m-1", "m-2", emberNow); err != ErrEmberIDRequired {
		t.Fatalf("missing id = %v", err)
	}
	if _, err := NewEmber("ember_1", "fire_1", "m-1", "m-1", emberNow); err != ErrSelfEmber {
		t.Fatalf("self ember = %v", err)
	}
	ember, err := NewEmber("ember_1", "fire_1", "m-1", "m-2", emberNow)
	if err != nil {
		t.Fatal(err)
	}
	if ember.ExpiresAt() != emberNow.Add(EmberLifetime) {
		t.Fatalf("expiry = %v, want 24h (FR-402)", ember.ExpiresAt())
	}
	if ember.Status() != StatusIssued {
		t.Fatalf("status = %q", ember.Status())
	}
}

func TestRedeemWithinWindow(t *testing.T) {
	ember, _ := NewEmber("ember_1", "fire_1", "m-1", "m-2", emberNow)
	if err := ember.Redeem(emberNow.Add(23 * time.Hour)); err != nil {
		t.Fatal(err)
	}
	if ember.Status() != StatusRedeemed || ember.RedeemedAt() == nil {
		t.Fatalf("ember = %#v", ember)
	}
	if err := ember.Redeem(emberNow.Add(23 * time.Hour)); err != ErrEmberNotOpen {
		t.Fatalf("re-redeem = %v, want not open", err)
	}
}

func TestRedeemAfterWindowExpires(t *testing.T) {
	ember, _ := NewEmber("ember_1", "fire_1", "m-1", "m-2", emberNow)
	if err := ember.Redeem(emberNow.Add(EmberLifetime)); err != ErrEmberExpired {
		t.Fatalf("at expiry = %v, want expired", err)
	}
	if ember.Status() != StatusExpired {
		t.Fatalf("status = %q, want expired", ember.Status())
	}
}

func TestMarkMutual(t *testing.T) {
	ember, _ := NewEmber("ember_1", "fire_1", "m-1", "m-2", emberNow)
	ember.MarkMutual()
	if ember.Status() != StatusMutual {
		t.Fatalf("status = %q", ember.Status())
	}
	// Mutual embers remain redeemable.
	if err := ember.Redeem(emberNow.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	// Redeemed embers never flip back.
	ember.MarkMutual()
	if ember.Status() != StatusRedeemed {
		t.Fatalf("status = %q", ember.Status())
	}
}
