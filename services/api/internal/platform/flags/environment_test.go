package flags

import (
	"strings"
	"testing"
)

func TestParseEnvironment(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"FEATURE_SOW_ENABLED":      " TRUE ",
		"KILL_SWITCH_AI":           "true",
		"FEATURE_PAYMENTS_ENABLED": "false",
	}
	configuration, err := ParseEnvironment(func(name string) string {
		return values[name]
	})
	if err != nil {
		t.Fatal(err)
	}
	registry := New(configuration, nil, nil)

	if got := registry.Evaluate(FlagSow); !got.Enabled || got.Source != SourceEnvironment {
		t.Fatalf("Sow override not parsed: %+v", got)
	}
	if got := registry.Evaluate(FlagAI); !got.Killed || got.Enabled {
		t.Fatalf("AI kill switch not parsed: %+v", got)
	}
	if got := registry.Evaluate(FlagPayments); got.Enabled || got.Source != SourceEnvironment {
		t.Fatalf("explicit false override not retained: %+v", got)
	}
	if got := registry.Evaluate(FlagGate); got.Enabled || got.Source != SourceDefault {
		t.Fatalf("unset flag did not use default: %+v", got)
	}
}

func TestParseEnvironmentRejectsInvalidValueWithoutEchoingIt(t *testing.T) {
	t.Parallel()

	const secretLikeValue = "token=do-not-log"
	_, err := ParseEnvironment(func(name string) string {
		if name == "KILL_SWITCH_GATE" {
			return secretLikeValue
		}
		return ""
	})
	if err == nil {
		t.Fatal("expected invalid value error")
	}
	if !strings.Contains(err.Error(), "KILL_SWITCH_GATE") {
		t.Fatalf("error did not identify safe variable name: %v", err)
	}
	if strings.Contains(err.Error(), secretLikeValue) {
		t.Fatalf("error echoed raw configuration: %v", err)
	}
}

func TestParseEnvironmentRejectsAmbiguousBoolean(t *testing.T) {
	t.Parallel()

	for _, invalid := range []string{"1", "yes", "on", "enabled"} {
		invalid := invalid
		t.Run(invalid, func(t *testing.T) {
			t.Parallel()
			_, err := ParseEnvironment(func(name string) string {
				if name == "FEATURE_FIRES_ENABLED" {
					return invalid
				}
				return ""
			})
			if err == nil {
				t.Fatalf("expected %q to be rejected", invalid)
			}
		})
	}
}
