package domain

import (
	"testing"
	"time"
)

var packNow = time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

func TestNewPackValidation(t *testing.T) {
	if _, err := NewPack("", MarketGhanaTwi, "term:gh-tw:1", nil, "actor-1", packNow); err != ErrPackIDRequired {
		t.Fatalf("missing id = %v", err)
	}
	if _, err := NewPack("p-1", Market("ng_en"), "term:gh-tw:1", nil, "actor-1", packNow); err != ErrInvalidMarket {
		t.Fatalf("unknown market = %v", err)
	}
	if _, err := NewPack("p-1", MarketGhanaTwi, " ", nil, "actor-1", packNow); err != ErrTerminologyRequired {
		t.Fatalf("missing terminology = %v", err)
	}
	if _, err := NewPack("p-1", MarketGhanaTwi, "term:gh-tw:1", nil, " ", packNow); err != ErrActorRequired {
		t.Fatalf("missing actor = %v", err)
	}
}

func TestPublishFourEyes(t *testing.T) {
	pack, err := NewPack("p-1", MarketGhanaTwi, "term:gh-tw:1", map[string]bool{"fires": true}, "proposer-1", packNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := pack.Publish("proposer-1", packNow); err != ErrSelfApproval {
		t.Fatalf("self approval = %v, want rejected (four eyes)", err)
	}
	if err := pack.Publish(" ", packNow); err != ErrActorRequired {
		t.Fatalf("blank approver = %v", err)
	}
	if err := pack.Publish("approver-1", packNow); err != nil {
		t.Fatal(err)
	}
	if pack.Status() != StatusPublished || pack.ApprovedBy() != "approver-1" || pack.PublishedAt() == nil {
		t.Fatalf("pack = %#v", pack)
	}
	if err := pack.Publish("approver-2", packNow); err != ErrPackNotDraft {
		t.Fatalf("re-publish = %v, want not draft", err)
	}
}

func TestRetireLifecycle(t *testing.T) {
	draft, _ := NewPack("p-1", MarketGhanaGa, "term:gh-ga:1", nil, "actor-1", packNow)
	if err := draft.Retire("actor-1"); err != ErrPackNotPublished {
		t.Fatalf("retire draft = %v, want not published", err)
	}
	if err := draft.Publish("actor-2", packNow); err != nil {
		t.Fatal(err)
	}
	if err := draft.Retire("actor-2"); err != nil {
		t.Fatal(err)
	}
	if draft.Status() != StatusRetired {
		t.Fatalf("status = %q", draft.Status())
	}
}
