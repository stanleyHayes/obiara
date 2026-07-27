package domain

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

var now = time.Date(2026, 7, 27, 3, 0, 0, 0, time.UTC)

func master(t *testing.T) Master {
	t.Helper()
	m, e := NewMaster(MasterSpec{ID: "ghana.master", Version: 4, Entries: []MasterEntry{{Key: "room.drum", Text: "The drum is with {name}."}, {Key: "welcome", Text: "Akwaaba to Obiara."}}, Terms: []Term{{Value: "Obiara", DoNotTranslate: true}, {Value: "Akwaaba"}}})
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func translations() []Translation {
	return []Translation{{Key: "room.drum", Text: "Drum no dey {name} nkyɛn."}, {Key: "welcome", Text: "Akwaaba ba Obiara."}}
}
func pack(t *testing.T) Pack {
	t.Helper()
	p, e := Propose(key(1), "GH", "tw-GH", key(2), 1, master(t), translations(), "propose-1", now)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func review(stage ReviewStage, reviewer int) Review {
	checks := map[ReviewStage][]Check{StageProfessional: {CheckMeaning, CheckVoice}, StageCommunity: {CheckCulturalFit, CheckDignity}, StageInContext: {CheckScreenshot, CheckTruncation}}[stage]
	return Review{Stage: stage, ReviewerKey: key(reviewer), Checks: checks, EvidenceRef: key(reviewer + 10), ReviewedAt: now}
}
func TestParityPlaceholdersAndTerminology(t *testing.T) {
	bad := translations()
	bad[0].Text = "Drum no dey {person} nkyɛn."
	if _, e := Propose(key(1), "GH", "tw-GH", key(2), 1, master(t), bad, "propose-1", now); e == nil {
		t.Fatal("placeholder drift")
	}
	bad = translations()
	bad[1].Text = "Akwaaba ba ha."
	if _, e := Propose(key(1), "GH", "tw-GH", key(2), 1, master(t), bad, "propose-1", now); e == nil {
		t.Fatal("do-not-translate term lost")
	}
	bad = translations()[:1]
	if _, e := Propose(key(1), "GH", "tw-GH", key(2), 1, master(t), bad, "propose-1", now); e == nil {
		t.Fatal("missing key")
	}
}
func TestDistinctHumanPipelineBeforePublishReady(t *testing.T) {
	p := pack(t)
	if _, e := p.Approve(key(9), key(19), "approve-1", now); e == nil {
		t.Fatal("approved without reviews")
	}
	for index, stage := range []ReviewStage{StageProfessional, StageCommunity, StageInContext} {
		var e error
		p, e = p.AddReview(review(stage, index+3), fmt.Sprintf("review-%d", index))
		if e != nil {
			t.Fatal(e)
		}
	}
	if _, e := p.Approve(key(3), key(19), "approve-1", now); e == nil {
		t.Fatal("reviewer self-approved")
	}
	p, e := p.Approve(key(9), key(19), "approve-1", now)
	if e != nil || !p.PublishReady() {
		t.Fatal(e)
	}
	state := p.State()
	state.Translations[0].Text = "mutated"
	if p.State().Translations[0].Text == "mutated" {
		t.Fatal("mutable version")
	}
}
func TestTranslationOrderDoesNotChangeVersion(t *testing.T) {
	base := translations()
	want := pack(t).State()
	random := rand.New(rand.NewSource(42))
	for range 1000 {
		x := append([]Translation(nil), base...)
		random.Shuffle(len(x), func(i, j int) { x[i], x[j] = x[j], x[i] })
		got, e := Propose(key(1), "GH", "tw-GH", key(2), 1, master(t), x, "propose-1", now)
		if e != nil || !reflect.DeepEqual(got.State(), want) {
			t.Fatal("order changed version")
		}
	}
}
func FuzzPlaceholderDriftNeverAccepted(f *testing.F) {
	f.Add("name")
	f.Add("person")
	f.Fuzz(func(t *testing.T, placeholder string) {
		if placeholder == "name" || !keyPattern.MatchString("x."+placeholder) {
			t.Skip()
		}
		x := translations()
		x[0].Text = "Drum no dey {" + placeholder + "} nkyɛn."
		if _, e := Propose(key(1), "GH", "tw-GH", key(2), 1, master(t), x, "propose-1", now); e == nil {
			t.Fatalf("accepted %q", placeholder)
		}
	})
}
