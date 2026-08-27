package main

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func completeEnv() map[string]string {
	return map[string]string{
		"MONGODB_URI":              "mongodb://localhost:27017",
		"MONGODB_DATABASE":         "obiara_test",
		"BOOTSTRAP_ADMIN_EMAIL":    "root@obiara.test",
		"BOOTSTRAP_ADMIN_PASSWORD": "Correct-Horse-Battery-9",
	}
}

func TestLoadSettingsDefaultsToEveryRole(t *testing.T) {
	loaded, err := loadSettings(env(completeEnv()))
	if err != nil {
		t.Fatalf("loadSettings: %v", err)
	}
	if len(loaded.roles) != len(allRoles()) {
		t.Errorf("roles = %v, want every role", roleNames(loaded.roles))
	}
	if !slices.Contains(loaded.roles, domain.RoleAdmin) {
		t.Error("default grant omits the admin role")
	}
}

func TestLoadSettingsRequiresEveryCredential(t *testing.T) {
	for _, missing := range []string{
		"MONGODB_URI", "MONGODB_DATABASE", "BOOTSTRAP_ADMIN_EMAIL", "BOOTSTRAP_ADMIN_PASSWORD",
	} {
		t.Run("without "+missing, func(t *testing.T) {
			values := completeEnv()
			delete(values, missing)
			_, err := loadSettings(env(values))
			if err == nil {
				t.Fatal("loadSettings succeeded, want error")
			}
			if !strings.Contains(err.Error(), missing) {
				t.Errorf("error %v should name %s", err, missing)
			}
		})
	}
}

// TestLoadSettingsEnforcesThePasswordPolicy checks the policy runs before we
// connect, so a weak password cannot leave a half-finished run.
func TestLoadSettingsEnforcesThePasswordPolicy(t *testing.T) {
	values := completeEnv()
	values["BOOTSTRAP_ADMIN_PASSWORD"] = "short"

	_, err := loadSettings(env(values))
	if err == nil {
		t.Fatal("loadSettings accepted a weak password")
	}
	if !strings.Contains(err.Error(), "BOOTSTRAP_ADMIN_PASSWORD") {
		t.Errorf("error %v should name the offending variable", err)
	}
	if strings.Contains(err.Error(), "short") && strings.Count(err.Error(), "short") > 1 {
		t.Errorf("error may name the variable but must not echo the password: %v", err)
	}
}

func TestLoadSettingsRejectsAnInvalidEmail(t *testing.T) {
	// An invalid address would only fail later, at principal construction.
	values := completeEnv()
	values["BOOTSTRAP_ADMIN_EMAIL"] = "not-an-email"
	loaded, err := loadSettings(env(values))
	if err != nil {
		t.Fatalf("loadSettings: %v", err)
	}
	if _, err := domain.NewPrincipal("adm_x", loaded.email, loaded.roles, time.Now()); err == nil {
		t.Fatal("an invalid bootstrap email produced a valid principal")
	}
}

func TestParseRoles(t *testing.T) {
	t.Run("explicit subset", func(t *testing.T) {
		roles, err := parseRoles("admin, verifier ,admin", false)
		if err != nil {
			t.Fatalf("parseRoles: %v", err)
		}
		want := []domain.Role{domain.RoleAdmin, domain.RoleVerifier}
		if !slices.Equal(roles, want) {
			t.Errorf("roles = %v, want %v (deduplicated, order preserved)", roles, want)
		}
	})

	t.Run("unknown role", func(t *testing.T) {
		if _, err := parseRoles("admin,superuser", false); err == nil {
			t.Fatal("parseRoles accepted an unknown role")
		}
	})

	// A grant without the admin role could not enroll anyone, leaving the
	// console just as unreachable as an empty database.
	t.Run("without admin", func(t *testing.T) {
		_, err := parseRoles("verifier,finance", false)
		if err == nil {
			t.Fatal("parseRoles accepted a grant with no admin role")
		}
		if !strings.Contains(err.Error(), "admin") {
			t.Errorf("error %v should explain the admin requirement", err)
		}
	})

	t.Run("blank falls back to every role", func(t *testing.T) {
		roles, err := parseRoles("   ", false)
		if err != nil {
			t.Fatalf("parseRoles: %v", err)
		}
		if !slices.Equal(roles, allRoles()) {
			t.Errorf("roles = %v, want every role", roles)
		}
	})
}

func TestNewIDIsPrefixedAndUnique(t *testing.T) {
	first, second := newID(), newID()
	if !strings.HasPrefix(first, "adm_") {
		t.Errorf("id %q lacks the adm_ prefix", first)
	}
	if first == second {
		t.Error("newID returned the same id twice")
	}
}

func TestParseRolesRequiresAdminUnlessAllowed(t *testing.T) {
	if _, err := parseRoles("verifier", false); err == nil {
		t.Fatal("parseRoles accepted a non-admin role set without the opt-out")
	}
	roles, err := parseRoles("verifier", true)
	if err != nil {
		t.Fatalf("parseRoles with opt-out: %v", err)
	}
	if len(roles) != 1 || roles[0] != domain.RoleVerifier {
		t.Fatalf("roles = %v, want [verifier]", roles)
	}
}

func TestLoadSettingsHonoursNonAdminOptOut(t *testing.T) {
	values := completeEnv()
	values["BOOTSTRAP_ADMIN_ROLES"] = "finance"
	if _, err := loadSettings(env(values)); err == nil {
		t.Fatal("loadSettings accepted a non-admin role set without the opt-out")
	}
	values["BOOTSTRAP_ALLOW_NON_ADMIN"] = "1"
	loaded, err := loadSettings(env(values))
	if err != nil {
		t.Fatalf("loadSettings: %v", err)
	}
	if len(loaded.roles) != 1 || loaded.roles[0] != domain.RoleFinance {
		t.Fatalf("roles = %v, want [finance]", loaded.roles)
	}
}
