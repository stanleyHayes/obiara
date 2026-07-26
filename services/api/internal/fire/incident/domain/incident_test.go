package domain

import (
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func k(n int) string { return fmt.Sprintf("%064x", n) }
func TestCategoryPropertyAndNeutralProjection(t *testing.T) {
	cats := []Category{CategoryThreat, CategoryHarassment, CategoryIdentity, CategoryMedical, CategoryOther}
	property := func(index uint8, evidence bool) bool {
		ref := ""
		if evidence {
			ref = k(3)
		}
		i, e := Create("case-1", k(1), k(2), cats[int(index)%len(cats)], ref, now, Command{ID: "trigger", At: now})
		if e != nil {
			return false
		}
		p := i.Project()
		return p.Reference == "case-1" && !p.Routed
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 1000}); e != nil {
		t.Fatal(e)
	}
}
func TestRouteIsExplicitAndImmutable(t *testing.T) {
	i, _ := Create("case-1", k(1), k(2), CategoryThreat, "", now, Command{ID: "trigger", At: now})
	r, e := i.Route(now, Command{ID: "trigger:route", ExpectedRevision: 1, At: now})
	if e != nil || !r.Project().Routed || len(r.Events()) != 2 {
		t.Fatal(e)
	}
}
