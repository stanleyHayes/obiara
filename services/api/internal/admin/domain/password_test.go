package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const goodPassword = "Correct-Horse-Battery-9"

func TestHashPasswordRoundTrips(t *testing.T) {
	encoded, err := HashPassword(goodPassword)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$") {
		t.Errorf("digest %q is not argon2id PHC", encoded)
	}
	if strings.Contains(encoded, goodPassword) {
		t.Fatal("digest contains the plaintext password")
	}
	if !VerifyPassword(encoded, goodPassword) {
		t.Error("VerifyPassword rejected the correct password")
	}
	if VerifyPassword(encoded, goodPassword+"x") {
		t.Error("VerifyPassword accepted a wrong password")
	}
}

func TestHashPasswordSaltsEachDigest(t *testing.T) {
	first, err := HashPassword(goodPassword)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	second, err := HashPassword(goodPassword)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	// Equal digests for equal passwords would let anyone with read access to
	// the collection spot operators who share a password.
	if first == second {
		t.Error("two digests of the same password are identical; the salt is not random")
	}
}

func TestCheckPasswordPolicy(t *testing.T) {
	cases := map[string]struct {
		password string
		want     error
	}{
		"good":               {goodPassword, nil},
		"too short":          {"Ab1!", ErrPasswordTooShort},
		"no upper":           {"correct-horse-battery-9", ErrPasswordWeak},
		"no lower":           {"CORRECT-HORSE-BATTERY-9", ErrPasswordWeak},
		"letters only":       {"CorrectHorseBattery", ErrPasswordWeak},
		"too long":           {strings.Repeat("aB1", 100), ErrPasswordTooLong},
		"unicode passphrase": {"Akwaaba-Obiara-2026-ɛ", nil},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			err := CheckPasswordPolicy(testCase.password)
			if !errors.Is(err, testCase.want) {
				t.Fatalf("CheckPasswordPolicy = %v, want %v", err, testCase.want)
			}
		})
	}
}

func TestVerifyPasswordRejectsMalformedDigests(t *testing.T) {
	for name, encoded := range map[string]string{
		"empty":            "",
		"not phc":          "plaintext",
		"wrong algorithm":  "$argon2i$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA",
		"wrong version":    "$argon2id$v=16$m=65536,t=3,p=4$c2FsdA$aGFzaA",
		"missing segments": "$argon2id$v=19$m=65536,t=3,p=4",
		"zero cost":        "$argon2id$v=19$m=0,t=0,p=0$c2FsdA$aGFzaA",
		"bad base64":       "$argon2id$v=19$m=65536,t=3,p=4$!!!$!!!",
	} {
		t.Run(name, func(t *testing.T) {
			if VerifyPassword(encoded, goodPassword) {
				t.Error("VerifyPassword accepted a malformed digest")
			}
		})
	}
}

func TestPrincipalPasswordLifecycle(t *testing.T) {
	principal, err := NewPrincipal("adm_1", "root@obiara.test", []Role{RoleAdmin}, time.Now())
	if err != nil {
		t.Fatalf("NewPrincipal: %v", err)
	}

	// A freshly enrolled principal has no password and must not be treated
	// as accepting any password.
	if principal.HasPassword() {
		t.Error("new principal reports a password")
	}
	if principal.VerifyPassword("") || principal.VerifyPassword(goodPassword) {
		t.Error("passwordless principal accepted a password")
	}

	before := principal.Version()
	if err := principal.SetPassword(goodPassword); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if !principal.HasPassword() {
		t.Error("principal does not report a password after SetPassword")
	}
	if principal.Version() != before+1 {
		t.Errorf("version = %d, want %d: a credential change must bump the version", principal.Version(), before+1)
	}
	if !principal.VerifyPassword(goodPassword) {
		t.Error("VerifyPassword rejected the password just set")
	}
	if principal.VerifyPassword("Wrong-Horse-Battery-9") {
		t.Error("VerifyPassword accepted a wrong password")
	}

	if err := principal.SetPassword("weak"); !errors.Is(err, ErrPasswordTooShort) {
		t.Errorf("SetPassword(weak) = %v, want ErrPasswordTooShort", err)
	}
	if !principal.VerifyPassword(goodPassword) {
		t.Error("a rejected SetPassword must leave the existing password intact")
	}
}

func TestReconstitutePrincipalWithPassword(t *testing.T) {
	encoded, err := HashPassword(goodPassword)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	principal := ReconstitutePrincipalWithPassword(
		"adm_1", "root@obiara.test", []Role{RoleAdmin}, StatusActive, encoded, 4, time.Now())

	if !principal.HasPassword() || !principal.VerifyPassword(goodPassword) {
		t.Error("a reconstituted principal lost its password")
	}
	if principal.PasswordHash() != encoded {
		t.Error("PasswordHash did not round-trip for persistence")
	}
	// Principals stored before password support must keep working.
	legacy := ReconstitutePrincipal("adm_2", "old@obiara.test", []Role{RoleAdmin}, StatusActive, 1, time.Now())
	if legacy.HasPassword() {
		t.Error("a legacy principal reports a password")
	}
}
