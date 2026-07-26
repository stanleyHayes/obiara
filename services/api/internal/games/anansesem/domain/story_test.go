package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func k(n int) string                { return fmt.Sprintf("%064x", n) }
func c(id string, r uint64) Command { return Command{ID: id, ExpectedRevision: r, At: now} }
func TestAlternationEditInvalidatesAndEditionRedacts(t *testing.T) {
	s, e := Create("story-1", k(3), "spider-path", []string{k(1), k(2)}, c("create", 0))
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Add("p1", k(1), "Once.", now, c("add1", 1))
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Add("p2", k(2), "Then.", now, c("add2", 2))
	if e != nil {
		t.Fatal(e)
	}
	s, _ = s.Grant(k(1), c("g1", 3))
	s, _ = s.Grant(k(2), c("g2", 4))
	s, e = s.Edit("p1", k(1), "Once again.", now, c("edit", 5))
	if e != nil || len(s.Grants()) != 0 {
		t.Fatal("edit retained grants")
	}
	s, _ = s.Grant(k(1), c("g3", 6))
	s, _ = s.Grant(k(2), c("g4", 7))
	_, edition, e := s.Publish(now, c("publish", 8))
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(edition)
	for _, bad := range []string{k(1), k(2), k(3), "author", "room", "contact"} {
		if strings.Contains(strings.ToLower(string(raw)), strings.ToLower(bad)) {
			t.Fatalf("edition leak %q: %s", bad, raw)
		}
	}
}
func TestAlternationProperty(t *testing.T) {
	property := func(count uint8) bool {
		s, e := Create("story-1", k(3), "spider-path", []string{k(1), k(2)}, c("create", 0))
		if e != nil {
			return false
		}
		limit := int(count % 20)
		for i := 0; i < limit; i++ {
			actor := k(i%2 + 1)
			s, e = s.Add(fmt.Sprintf("p%d", i), actor, "bounded", now, Command{ID: fmt.Sprintf("c%d", i), ExpectedRevision: uint64(i + 1), At: now})
			if e != nil {
				return false
			}
		}
		for i, p := range s.Passages() {
			if p.AuthorKey != k(i%2+1) {
				return false
			}
		}
		return true
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 1000}); e != nil {
		t.Fatal(e)
	}
}
