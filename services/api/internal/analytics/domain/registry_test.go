package domain

import (
	"testing"
)

func TestRegisteredEventValidates(t *testing.T) {
	err := ValidateProps("epono.pod_heard", map[string]any{
		"durationPct":   82.5,
		"trustPathType": "circle",
	})
	if err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}
}

func TestUnregisteredEventRejected(t *testing.T) {
	if err := ValidateProps("evil.custom_event", map[string]any{}); err == nil {
		t.Fatal("unregistered event must fail")
	}
}

func TestUnknownPropRejected(t *testing.T) {
	err := ValidateProps("epono.pod_heard", map[string]any{
		"durationPct":   82.5,
		"trustPathType": "circle",
		"rawTranscript": "member said something private",
	})
	if err == nil {
		t.Fatal("undeclared prop (content smuggling) must fail")
	}
}

func TestFreeTextCannotValidate(t *testing.T) {
	// Enum prop with arbitrary text.
	if err := ValidateProps("danmu.room_closed", map[string]any{"mode": "she said she loved him"}); err == nil {
		t.Fatal("free text in enum prop must fail")
	}
	// OpaqueID with spaces (free text).
	if err := ValidateProps("danmu.theme_completed", map[string]any{"themeId": "my secret answer"}); err == nil {
		t.Fatal("free text in opaque id prop must fail")
	}
	// OpaqueID over length.
	if err := ValidateProps("danmu.theme_completed", map[string]any{"themeId": string(make([]byte, 0))}); err == nil {
		t.Fatal("empty opaque id must fail")
	}
}

func TestRequiredPropsEnforced(t *testing.T) {
	if err := ValidateProps("epono.pod_heard", map[string]any{"durationPct": 50}); err == nil {
		t.Fatal("missing required trustPathType must fail")
	}
	if err := ValidateProps("gyaase.fire_attended", map[string]any{"type": "weekly", "minutes": 90}); err != nil {
		t.Fatalf("optional gamesPlayed may be omitted: %v", err)
	}
}

func TestPropTypesEnforced(t *testing.T) {
	if err := ValidateProps("epono.pod_heard", map[string]any{"durationPct": "lots", "trustPathType": "circle"}); err == nil {
		t.Fatal("string in number prop must fail")
	}
	if err := ValidateProps("gyaase.fire_attended", map[string]any{"type": "weekly", "minutes": 45, "gamesPlayed": true}); err == nil {
		t.Fatal("bool in number prop must fail")
	}
}

func TestRegistryCoversDoc08Taxonomy(t *testing.T) {
	for _, name := range []string{
		"epono.pod_heard", "epono.seed_sown", "epono.sprout_opened", "epono.room_opened",
		"danmu.drum_passed", "danmu.theme_completed", "danmu.room_closed",
		"gyaase.fire_attended", "gyaase.ember_converted",
		"ceremony.gate_crossed", "ceremony.aseda_declared",
		"wellbeing.regret_reported", "commerce.order_completed",
	} {
		if _, ok := find(name); !ok {
			t.Errorf("taxonomy event %q missing from registry", name)
		}
	}
}
