package secrets_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	platformsecrets "github.com/stanleyHayes/obiara/internal/platform/secrets"
	"go.yaml.in/yaml/v3"
)

type inventorySecret struct {
	Name                                       string   `yaml:"name"`
	Services                                   []string `yaml:"services"`
	RotatedAt                                  string   `yaml:"rotated_at_variable"`
	MaxAgeDays                                 int      `yaml:"maximum_age_days"`
	OverlapMinutes                             int      `yaml:"overlap_minutes"`
	Owner, Custodian, Reload, Revoke, Evidence string
}
type inventory struct {
	Version int               `yaml:"version"`
	Secrets []inventorySecret `yaml:"secrets"`
}
type renderVariable struct {
	Key, Value string
	Sync       *bool
}
type renderService struct {
	Name    string
	EnvVars []renderVariable `yaml:"envVars"`
}
type renderFile struct {
	Projects []struct {
		Environments []struct {
			Services []renderService
		}
	}
}

func root(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}
func load(t *testing.T) inventory {
	t.Helper()
	raw, e := os.ReadFile(filepath.Join(root(t), "deploy/secrets/inventory.yaml"))
	if e != nil {
		t.Fatal(e)
	}
	var got inventory
	if e = yaml.Unmarshal(raw, &got); e != nil {
		t.Fatal(e)
	}
	return got
}

func TestInventoryMatchesRuntimeAndHasCompleteRotationControls(t *testing.T) {
	got := load(t)
	if got.Version != 1 || len(got.Secrets) != len(platformsecrets.Inventory()) {
		t.Fatalf("inventory shape %#v", got)
	}
	byName := map[string]inventorySecret{}
	for _, s := range got.Secrets {
		byName[s.Name] = s
	}
	for _, definition := range platformsecrets.Inventory() {
		row, ok := byName[definition.Name]
		if !ok {
			t.Fatalf("%s missing", definition.Name)
		}
		services := make([]string, len(definition.Services))
		for i, v := range definition.Services {
			services[i] = string(v)
		}
		slices.Sort(services)
		slices.Sort(row.Services)
		if !slices.Equal(services, row.Services) || row.RotatedAt != definition.RotatedAtVariable || row.MaxAgeDays != int(definition.MaxAge/(24*time.Hour)) || row.OverlapMinutes <= 0 || row.Owner == "" || row.Custodian == "" || row.Reload == "" || row.Revoke == "" || row.Evidence == "" {
			t.Fatalf("incomplete policy for %s: %#v", definition.Name, row)
		}
	}
}

func TestRenderDeclaresValueFreeRuntimeSlots(t *testing.T) {
	raw, e := os.ReadFile(filepath.Join(root(t), "render.yaml"))
	if e != nil {
		t.Fatal(e)
	}
	var render renderFile
	if e = yaml.Unmarshal(raw, &render); e != nil {
		t.Fatal(e)
	}
	found := map[string]int{}
	for _, project := range render.Projects {
		for _, environment := range project.Environments {
			for _, service := range environment.Services {
				for _, variable := range service.EnvVars {
					for _, definition := range platformsecrets.Inventory() {
						if variable.Key == definition.Name || variable.Key == definition.RotatedAtVariable {
							found[variable.Key]++
							if variable.Value != "" || variable.Sync == nil || *variable.Sync {
								t.Fatalf("%s %s must be value-free sync:false", service.Name, variable.Key)
							}
						}
					}
				}
			}
		}
	}
	for _, definition := range platformsecrets.Inventory() {
		if found[definition.Name] != len(definition.Services) || found[definition.RotatedAtVariable] != len(definition.Services) {
			t.Fatalf("wrong Render slot counts for %s: %#v", definition.Name, found)
		}
	}
}

func TestTrackedFilesContainNoHighConfidenceSecretValues(t *testing.T) {
	repo := root(t)
	cmd := exec.Command("git", "ls-files", "-z")
	cmd.Dir = repo
	out, e := cmd.Output()
	if e != nil {
		t.Fatal(e)
	}
	patterns := []struct {
		name string
		re   *regexp.Regexp
	}{
		{"credentialed MongoDB URI", regexp.MustCompile(`mongodb(?:\+srv)?://[^\s:/]+:[^\s@/]+@`)},
		{"Resend API key", regexp.MustCompile(`\bre_[A-Za-z0-9]{20,}\b`)},
		{"webhook secret", regexp.MustCompile(`\bwhsec_[A-Za-z0-9_+/=-]{16,}\b`)},
		{"private key", regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`)},
		{"GitHub token", regexp.MustCompile(`\bgh[oprsu]_[A-Za-z0-9]{30,}\b`)},
	}
	for _, name := range bytes.Split(out, []byte{0}) {
		if len(name) == 0 {
			continue
		}
		path := string(name)
		if strings.HasSuffix(path, ".docx") || strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".ico") {
			continue
		}
		raw, e := os.ReadFile(filepath.Join(repo, path))
		if e != nil {
			t.Fatal(e)
		}
		for _, pattern := range patterns {
			if pattern.re.Match(raw) {
				t.Fatalf("%s contains a %s shape", path, pattern.name)
			}
		}
	}
}
