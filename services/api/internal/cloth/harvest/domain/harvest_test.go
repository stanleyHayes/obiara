package domain

import (
	"errors"
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func k(n int) string { return fmt.Sprintf("%064x", n) }
func p() Payload {
	return Payload{RecipeKey: k(3), RecipeVersion: "grammar-v1", RenderSeed: k(4), ProductionTokens: []string{"warp_even", "weft_close", "edge_soft", "tone_warm", "mark_sparse", "finish_matte"}, Format: "woven_band", DeliveryRef: k(5), PolicyVersion: "policy-v1"}
}
func c(id string, actor int, rev uint64) Command {
	return Command{ID: id, ActorKey: k(actor), ExpectedRevision: rev, At: now}
}
func TestFullPayloadConsentAndTerminalProviderAccess(t *testing.T) {
	h, e := Create("harvest-1", []string{k(1), k(2)}, p(), c("create", 1, 0))
	if e != nil {
		t.Fatal(e)
	}
	h, e = h.Approve(c("approve-a", 1, 1))
	if e != nil {
		t.Fatal(e)
	}
	h, e = h.Approve(c("approve-b", 2, 2))
	if e != nil || h.Status() != StatusReady {
		t.Fatal(e)
	}
	changed := p()
	changed.Format = "digital_archive"
	h, e = h.Revise(changed, c("revise", 1, 3))
	if e != nil || len(h.Approvals()) != 0 || h.Status() != StatusAwaiting {
		t.Fatal("consent not reset")
	}
	if _, e = h.Handoff("handoff", c("handoff", 1, 4)); !errors.Is(e, ErrConsent) {
		t.Fatal(e)
	}
}
func TestRecipeChangeProperty(t *testing.T) {
	property := func(index uint8) bool {
		base := p()
		h, e := Create("harvest-1", []string{k(1), k(2)}, base, c("create", 1, 0))
		if e != nil {
			return false
		}
		h, _ = h.Approve(c("a", 1, 1))
		h, _ = h.Approve(c("b", 2, 2))
		changed := base
		switch index % 4 {
		case 0:
			changed.RecipeVersion = "grammar-v2"
		case 1:
			changed.ProductionTokens[5] = "finish_lustre"
		case 2:
			changed.Format = "framed_cloth"
		case 3:
			changed.DeliveryRef = k(9)
		}
		h, e = h.Revise(changed, c("revise", 1, 3))
		return e == nil && len(h.Approvals()) == 0 && h.Status() == StatusAwaiting
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 1000}); e != nil {
		t.Fatal(e)
	}
}

func TestProviderAccessDeniedAfterCancelAndExpiry(t *testing.T) {
	ready := func() Harvest {
		h, _ := Create("harvest-1", []string{k(1), k(2)}, p(), c("create", 1, 0))
		h, _ = h.Approve(c("a", 1, 1))
		h, _ = h.Approve(c("b", 2, 2))
		return h
	}
	handed, err := ready().Handoff("handoff-1", c("handoff", 1, 3))
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := handed.Cancel(c("cancel", 2, 4))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cancelled.ProviderEnvelope(now.Add(time.Minute)); !errors.Is(err, ErrTransition) {
		t.Fatalf("cancelled provider access=%v", err)
	}
	expired, err := ready().Expire(now.Add(ReadyValidity), c("expire", 1, 3))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = expired.ProviderEnvelope(now.Add(ReadyValidity)); !errors.Is(err, ErrTransition) {
		t.Fatalf("expired provider access=%v", err)
	}
}
