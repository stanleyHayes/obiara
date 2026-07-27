package domain

import "testing"

func TestScoreWeightsDistinctKinds(t *testing.T) {
	if score := Score(nil); score != 0 {
		t.Fatalf("empty = %v", score)
	}
	if score := Score([]SignalKind{SignalAffectionCadence}); score != 1.0 {
		t.Fatalf("single = %v, want 1.0", score)
	}
	// Same kind twice counts once.
	if score := Score([]SignalKind{SignalAskPattern, SignalAskPattern}); score != 3.0 {
		t.Fatalf("duplicate kind = %v, want 3.0", score)
	}
	// Two kinds: (1.0+3.0) × 1.25 = 5.0.
	if score := Score([]SignalKind{SignalAffectionCadence, SignalAskPattern}); score != 5.0 {
		t.Fatalf("two kinds = %v, want 5.0", score)
	}
	// Three kinds: (1.0+2.0+2.5) × 1.5 = 8.25.
	if score := Score([]SignalKind{SignalAffectionCadence, SignalEmergencyNarrative, SignalOffPlatformPull}); score != 8.25 {
		t.Fatalf("three kinds = %v, want 8.25", score)
	}
}

func TestLadderThresholds(t *testing.T) {
	for score, want := range map[float64]LadderState{
		0:    LadderNone,
		2.0:  LadderWatch,
		4.0:  LadderEducation,
		5.99: LadderEducation,
		6.0:  LadderFriction,
		8.0:  LadderCase,
		99.0: LadderCase,
	} {
		if got := LadderFor(score); got != want {
			t.Fatalf("LadderFor(%v) = %v, want %v", score, got, want)
		}
	}
}

func TestSignalValidation(t *testing.T) {
	if !(Signal{RoomID: "r", ActorID: "a", Kind: SignalAskPattern}).Valid() {
		t.Fatal("well-formed signal rejected")
	}
	if (Signal{RoomID: "r", ActorID: "a", Kind: SignalKind("other")}).Valid() {
		t.Fatal("unknown kind accepted")
	}
	if (Signal{ActorID: "a", Kind: SignalAskPattern}).Valid() {
		t.Fatal("missing room accepted")
	}
}
