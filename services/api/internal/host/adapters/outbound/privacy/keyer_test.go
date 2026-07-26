package privacy

import (
	"strings"
	"testing"
)

func TestKeyerSeparatesNamespacesAndHidesInput(t *testing.T) {
	keyer, _ := NewHMACKeyer([]byte(strings.Repeat("k", 32)))
	applicant, _ := keyer.Key("applicant", "member:1")
	issuer, _ := keyer.Key("issuer", "member:1")
	if len(applicant) != 64 || applicant == issuer || strings.Contains(applicant, "member") {
		t.Fatal("privacy key did not separate or hide identifiers")
	}
}
