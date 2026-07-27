package domain

import (
	"fmt"
	suban "github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

var now = time.Date(2026, 7, 27, 1, 0, 0, 0, time.UTC)

func events() []suban.Event {
	return []suban.Event{{ID: "event-1", SubjectID: "member-1", Kind: suban.KindMeetingFollowThrough, Provenance: suban.Provenance{Source: "meeting", Ref: "private-ref"}, OccurredAt: now}, {ID: "event-2", SubjectID: "member-1", Kind: suban.KindHarassmentFinding, Provenance: suban.Provenance{Source: "panel", Ref: "raw-evidence"}, OccurredAt: now.Add(time.Minute)}}
}
func TestExplanationIsMemberBoundedAndRedacted(t *testing.T) {
	x, e := Explain("member-1", events(), now.Add(time.Hour))
	if e != nil {
		t.Fatal(e)
	}
	if len(x.Events) != 2 || len(x.Marks) != 0 {
		t.Fatal(x)
	}
	render := fmt.Sprintf("%+v", x)
	for _, bad := range []string{"private-ref", "raw-evidence", "score", "credit"} {
		if strings.Contains(render, bad) {
			t.Fatalf("leak %q: %s", bad, render)
		}
	}
	wrong := events()
	wrong[0].SubjectID = "member-2"
	if _, e = Explain("member-1", wrong, now); e == nil {
		t.Fatal("other member event exposed")
	}
}
func TestAppealPreservesEventAndRequiresDistinctReviewer(t *testing.T) {
	a, e := File(key(1), "member-1", "event-2", ReasonEventInaccurate, "file-1", now)
	if e != nil {
		t.Fatal(e)
	}
	before := a.State().EventID
	if _, e = a.Resolve(StatusOverturned, "member-1", key(3), "resolve-1", now); e == nil {
		t.Fatal("self review")
	}
	a, e = a.Resolve(StatusOverturned, key(2), key(3), "resolve-1", now)
	if e != nil || a.State().EventID != before || a.State().ReasoningRef != key(3) {
		t.Fatal(e)
	}
	audit := a.State().Audit
	audit[0].Kind = "deleted"
	if a.State().Audit[0].Kind != "filed" {
		t.Fatal("mutable audit")
	}
}
func TestEventOrderingIsDeterministic(t *testing.T) {
	base := events()
	want, _ := Explain("member-1", base, now.Add(time.Hour))
	random := rand.New(rand.NewSource(42))
	for range 1000 {
		x := append([]suban.Event(nil), base...)
		random.Shuffle(len(x), func(i, j int) { x[i], x[j] = x[j], x[i] })
		got, e := Explain("member-1", x, now.Add(time.Hour))
		if e != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("non deterministic")
		}
	}
}
func FuzzOnlyAdverseEventsAppealable(f *testing.F) {
	for _, kind := range []string{"meeting_follow_through", "harassment_finding", "fraud_finding", "unknown"} {
		f.Add(kind)
	}
	f.Fuzz(func(t *testing.T, value string) {
		kind := suban.Kind(value)
		got := Appealable(kind)
		want := kind == suban.KindGhostPattern || kind == suban.KindHarassmentFinding || kind == suban.KindFraudFinding || kind == suban.KindVouchStakeLoss
		if got != want {
			t.Fatalf("%q", value)
		}
	})
}
