package domain

import "testing"

func TestDefaults(t *testing.T) {
	if on, err := State(PurposeIdentitySafety, nil); err != nil || !on {
		t.Fatalf("identity safety default = %v, %v", on, err)
	}
	if on, _ := State(PurposeMatching, nil); on {
		t.Fatal("matching defaults off (opt-in)")
	}
	if on, _ := State(PurposeScamArc, nil); !on {
		t.Fatal("scam-arc defaults on (opt-out)")
	}
	if on, _ := State(PurposePlayPortraits, nil); on {
		t.Fatal("play-portraits defaults off (opt-in)")
	}
	if on, _ := State(PurposeProductAnalytics, nil); !on {
		t.Fatal("analytics defaults on (opt-out)")
	}
	if _, err := State(Purpose("made_up"), nil); err != ErrInvalidPurpose {
		t.Fatalf("unknown purpose = %v", err)
	}
}

func TestExplicitChoiceWins(t *testing.T) {
	off := false
	if on, _ := State(PurposeScamArc, &off); on {
		t.Fatal("explicit opt-out must win over default-on")
	}
	on := true
	if state, _ := State(PurposeMatching, &on); !state {
		t.Fatal("explicit opt-in must win over default-off")
	}
}

func TestControlRules(t *testing.T) {
	if err := ValidateChange(PurposeIdentitySafety, false); err != ErrPurposeLocked {
		t.Fatalf("disabling required purpose = %v, want locked", err)
	}
	if err := ValidateChange(PurposeIdentitySafety, true); err != ErrPurposeLocked {
		t.Fatalf("changing required purpose at all = %v, want locked", err)
	}
	if err := ValidateChange(PurposeMatching, false); err != ErrWrongDirection {
		t.Fatalf("disabling opt-in purpose = %v, want wrong direction", err)
	}
	if err := ValidateChange(PurposeMatching, true); err != nil {
		t.Fatalf("enabling opt-in purpose = %v", err)
	}
	if err := ValidateChange(PurposeScamArc, true); err != ErrWrongDirection {
		t.Fatalf("enabling opt-out purpose = %v, want wrong direction", err)
	}
	if err := ValidateChange(PurposeScamArc, false); err != nil {
		t.Fatalf("disabling opt-out purpose = %v", err)
	}
}
