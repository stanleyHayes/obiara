package atlasconfig

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func root(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}
func TestRepositoryAtlasSpecifications(t *testing.T) {
	base := root(t)
	for _, name := range []string{"staging.yaml", "production.yaml"} {
		t.Run(name, func(t *testing.T) {
			if err := ValidateFile(base, filepath.Join(base, "deploy", "atlas", name)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestSecurityAndRecoveryRegressionsFailClosed(t *testing.T) {
	base := root(t)
	production, raw, err := Load(filepath.Join(base, "deploy", "atlas", "production.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(*Spec)
	}{
		{"public", func(s *Spec) { s.Network.PublicAccess = true }},
		{"wildcard", func(s *Spec) { s.Network.AllowlistRefs = []string{"0.0.0.0/0", "worker"} }},
		{"shared-credential", func(s *Spec) { s.Identities[1].CredentialRef = s.Identities[0].CredentialRef }},
		{"no-c4-encryption", func(s *Spec) { s.Encryption.C4FieldEncryption = "optional" }},
		{"cross-region-backup", func(s *Spec) { s.Backup.Region = "EU_WEST_1" }},
		{"weak-rpo", func(s *Spec) { s.Backup.RPOMinutes = 6 }},
		{"destructive-restore", func(s *Spec) { s.Restore.DestructiveInPlace = true }},
		{"false-production-ready", func(s *Spec) { s.Activation.State = "ready" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := production
			candidate.Network.AllowlistRefs = append([]string(nil), production.Network.AllowlistRefs...)
			candidate.Identities = append([]Identity(nil), production.Identities...)
			tc.mutate(&candidate)
			if err := Validate(candidate, string(raw)); err == nil {
				t.Fatal("unsafe mutation accepted")
			}
		})
	}
}
func TestCommittedSpecsContainReferencesNotSecrets(t *testing.T) {
	base := root(t)
	for _, name := range []string{"staging.yaml", "production.yaml"} {
		_, raw, err := Load(filepath.Join(base, "deploy", "atlas", name))
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(raw))
		for _, bad := range []string{"mongodb+srv://", "mongodb://", "0.0.0.0/0", "password:", "-----begin"} {
			if strings.Contains(lower, bad) {
				t.Fatalf("%s contains %q", name, bad)
			}
		}
	}
}
