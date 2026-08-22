package apihttp

import (
	"errors"
	"strings"
	"testing"
)

// TestSanitizeFaultMasksIdentifiers keeps the fault key loggable. The runtime
// logger redacts "error" outright because an arbitrary chain can quote member
// data, so this must remove the same class of identifiers itself.
func TestSanitizeFaultMasksIdentifiers(t *testing.T) {
	cases := map[string]struct {
		cause   error
		absent  []string
		present []string
	}{
		"email address": {
			cause:   errors.New(`duplicate key: {email: "member@example.test"}`),
			absent:  []string{"member@example.test"},
			present: []string{"[email]", "duplicate key"},
		},
		"phone number": {
			cause:   errors.New(`no account for +233544919953`),
			absent:  []string{"233544919953"},
			present: []string{"[digits]"},
		},
		"opaque bearer token": {
			// Deliberately not shaped like any real provider key: the repo's
			// secret scanner treats those shapes as findings even in tests.
			cause:   errors.New(`rejected token ` + strings.Repeat("z", 32)),
			absent:  []string{strings.Repeat("z", 32)},
			present: []string{"[token]"},
		},
		"safe provider fault survives": {
			cause:   errors.New("resend: delivery failed: status 403, provider error validation_error"),
			present: []string{"403", "validation_error"},
		},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			got := sanitizeFault(testCase.cause)
			for _, absent := range testCase.absent {
				if strings.Contains(got, absent) {
					t.Errorf("sanitizeFault leaked %q: %s", absent, got)
				}
			}
			for _, present := range testCase.present {
				if !strings.Contains(got, present) {
					t.Errorf("sanitizeFault dropped %q, leaving nothing to triage: %s", present, got)
				}
			}
		})
	}
}

func TestSanitizeFaultIsBounded(t *testing.T) {
	got := sanitizeFault(errors.New(strings.Repeat("x", 5000)))
	if len(got) > maxFaultLength+4 {
		t.Errorf("sanitizeFault returned %d characters, want it bounded", len(got))
	}
}

func TestSanitizeFaultStripsControlCharacters(t *testing.T) {
	// A newline in a log line would let a fault forge a second entry.
	got := sanitizeFault(errors.New("first\nsecond\r\tthird"))
	if strings.ContainsAny(got, "\n\r\t") {
		t.Errorf("sanitizeFault kept control characters: %q", got)
	}
}
