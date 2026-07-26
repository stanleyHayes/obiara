package flags

import (
	"fmt"
	"strings"
)

var environmentNames = map[Flag]struct {
	enabled string
	killed  string
}{
	FlagSow:      {enabled: "FEATURE_SOW_ENABLED", killed: "KILL_SWITCH_SOW"},
	FlagFires:    {enabled: "FEATURE_FIRES_ENABLED", killed: "KILL_SWITCH_FIRES"},
	FlagAI:       {enabled: "FEATURE_AI_ENABLED", killed: "KILL_SWITCH_AI"},
	FlagPayments: {enabled: "FEATURE_PAYMENTS_ENABLED", killed: "KILL_SWITCH_PAYMENTS"},
	FlagGate:     {enabled: "FEATURE_GATE_ENABLED", killed: "KILL_SWITCH_GATE"},
}

// ParseEnvironment reads only the fixed, documented variables. Empty values
// mean no override; accepted values are exactly true or false, ignoring case
// and surrounding whitespace.
func ParseEnvironment(getenv func(string) string) (Configuration, error) {
	enabled := make(map[Flag]bool)
	killed := make(map[Flag]bool)

	for _, flag := range canonicalFlags {
		names := environmentNames[flag]
		value, set, err := parseBoolean(getenv(names.enabled))
		if err != nil {
			return Configuration{}, fmt.Errorf("%s: %w", names.enabled, err)
		}
		if set {
			enabled[flag] = value
		}

		value, set, err = parseBoolean(getenv(names.killed))
		if err != nil {
			return Configuration{}, fmt.Errorf("%s: %w", names.killed, err)
		}
		if set {
			killed[flag] = value
		}
	}
	return NewConfiguration(enabled, killed)
}

func parseBoolean(raw string) (value bool, set bool, err error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case "":
		return false, false, nil
	case "true":
		return true, true, nil
	case "false":
		return false, true, nil
	default:
		return false, false, fmt.Errorf("must be true or false")
	}
}
