package domain

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var now = time.Date(2026, 7, 26, 20, 0, 0, 0, time.UTC)

const member = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const refKey = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const redact = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
const tokenKey = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"

func command(id string, r uint64, at time.Time) Command {
	return Command{ID: id, ExpectedRevision: r, At: at}
}
func requested(t *testing.T) Export {
	t.Helper()
	e, x := Request("export:1", member, PurposeVictimSupport, []Reference{{Kind: KindIncidentSummary, RefKey: refKey, RedactionKey: redact}}, command("request", 0, now))
	if x != nil {
		t.Fatal(x)
	}
	return e
}
func TestAuthorizationIsOneTimeRevocableAndExactly72Hours(t *testing.T) {
	e, x := requested(t).Authorize(requested(t).References(), tokenKey, command("authorize", 1, now))
	if x != nil {
		t.Fatal(x)
	}
	if e.Authorization().ExpiresAt.Sub(e.Authorization().AuthorizedAt) != AuthorizationTTL {
		t.Fatal("wrong ttl")
	}
	used, x := e.Use(tokenKey, command("use", 2, now.Add(AuthorizationTTL-time.Second)))
	if x != nil || used.Status() != StatusUsed {
		t.Fatalf("used=%s err=%v", used.Status(), x)
	}
	if _, x = used.Use(tokenKey, command("again", 3, now.Add(time.Hour))); x == nil {
		t.Fatal("token reused")
	}
	e, _ = requested(t).Authorize(requested(t).References(), tokenKey, command("authorize", 1, now))
	revoked, x := e.Revoke(command("revoke", 2, now.Add(time.Hour)))
	if x != nil {
		t.Fatal(x)
	}
	if _, x = revoked.Use(tokenKey, command("use", 3, now.Add(2*time.Hour))); x == nil {
		t.Fatal("revoked token used")
	}
}
func TestAuthorizationTTLProperty(t *testing.T) {
	r := rand.New(rand.NewSource(828))
	for i := 0; i < 1000; i++ {
		at := now.Add(time.Duration(r.Intn(100000)) * time.Second)
		e, _ := Request(fmt.Sprintf("export:%d", i+2), member, PurposeLegalSupport, []Reference{{Kind: KindMessageMetadata, RefKey: refKey, RedactionKey: redact}}, command(fmt.Sprintf("request:%d", i), 0, at))
		e, x := e.Authorize(e.References(), tokenKey, command(fmt.Sprintf("auth:%d", i), 1, at))
		if x != nil || e.Authorization().ExpiresAt.Sub(at) != AuthorizationTTL {
			t.Fatalf("case %d err=%v", i, x)
		}
	}
}
func FuzzRequestReferenceCountBound(f *testing.F) {
	f.Add(1)
	f.Add(21)
	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 || n > 100 {
			return
		}
		refs := make([]Reference, n)
		for i := range refs {
			refs[i] = Reference{Kind: KindIncidentSummary, RefKey: fmt.Sprintf("%064x", i+1), RedactionKey: fmt.Sprintf("%064x", i+100)}
		}
		_, e := Request("export:fuzz", member, PurposePersonalRecord, refs, command("request:fuzz", 0, now))
		valid := n > 0 && n <= MaxReferences
		if valid != (e == nil) {
			t.Fatalf("n=%d valid=%v err=%v", n, valid, e)
		}
	})
}
