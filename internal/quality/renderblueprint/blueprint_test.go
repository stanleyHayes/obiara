package renderblueprint_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type blueprint struct {
	Previews struct {
		Generation      string `yaml:"generation"`
		ExpireAfterDays int    `yaml:"expireAfterDays"`
	} `yaml:"previews"`
	Projects []project `yaml:"projects"`
}

type project struct {
	Name         string        `yaml:"name"`
	Environments []environment `yaml:"environments"`
}

type environment struct {
	Name       string `yaml:"name"`
	Networking struct {
		Isolation string `yaml:"isolation"`
	} `yaml:"networking"`
	Permissions struct {
		Protection string `yaml:"protection"`
	} `yaml:"permissions"`
	Services []service `yaml:"services"`
}

type service struct {
	Name              string   `yaml:"name"`
	Type              string   `yaml:"type"`
	Runtime           string   `yaml:"runtime"`
	Region            string   `yaml:"region"`
	BuildCommand      string   `yaml:"buildCommand"`
	StartCommand      string   `yaml:"startCommand"`
	HealthCheckPath   string   `yaml:"healthCheckPath"`
	AutoDeployTrigger string   `yaml:"autoDeployTrigger"`
	NumInstances      int      `yaml:"numInstances"`
	EnvVars           []envVar `yaml:"envVars"`
}

type envVar struct {
	Key       string `yaml:"key"`
	Value     string `yaml:"value"`
	Sync      *bool  `yaml:"sync"`
	FromGroup string `yaml:"fromGroup"`
}

func load(t *testing.T) (blueprint, string) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
	path := filepath.Join(root, "render.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var parsed blueprint
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse render.yaml: %v", err)
	}
	return parsed, string(raw)
}

func TestProtectedBackendOnlyProductionTopology(t *testing.T) {
	parsed, _ := load(t)
	if parsed.Previews.Generation != "manual" || parsed.Previews.ExpireAfterDays != 7 {
		t.Fatalf("preview policy = %#v", parsed.Previews)
	}
	if len(parsed.Projects) != 1 || parsed.Projects[0].Name != "obiara" {
		t.Fatalf("projects = %#v", parsed.Projects)
	}
	environments := parsed.Projects[0].Environments
	if len(environments) != 1 || environments[0].Name != "production" {
		t.Fatalf("production environment = %#v", environments)
	}
	env := environments[0]
	if env.Networking.Isolation != "enabled" || env.Permissions.Protection != "enabled" {
		t.Fatal("production must be network-isolated and protected")
	}
	if len(env.Services) != 2 {
		t.Fatalf("services = %d, want backend api and worker only", len(env.Services))
	}
}

func TestServicesArePinnedStatelessAndCheckGated(t *testing.T) {
	parsed, _ := load(t)
	services := parsed.Projects[0].Environments[0].Services
	seen := map[string]service{}
	for _, candidate := range services {
		if _, exists := seen[candidate.Name]; exists {
			t.Fatalf("duplicate service %q", candidate.Name)
		}
		seen[candidate.Name] = candidate
		if candidate.Region != "frankfurt" {
			t.Fatalf("%s region = %q", candidate.Name, candidate.Region)
		}
		if candidate.AutoDeployTrigger != "off" {
			t.Fatalf("%s deploy trigger = %q", candidate.Name, candidate.AutoDeployTrigger)
		}
		if candidate.NumInstances != 1 || candidate.BuildCommand == "" || candidate.StartCommand == "" {
			t.Fatalf("%s missing pinned build/start/scale", candidate.Name)
		}
	}
	api := seen["obiara-api-production"]
	if api.Type != "web" || api.Runtime != "go" || api.HealthCheckPath != "/live" {
		t.Fatalf("api = %#v", api)
	}
	worker := seen["obiara-worker-production"]
	if worker.Type != "worker" || worker.Runtime != "go" || worker.HealthCheckPath != "" {
		t.Fatalf("worker = %#v", worker)
	}
	for name, service := range seen {
		if service.Runtime == "node" {
			t.Fatalf("frontend service %s remains in Render: %#v", name, service)
		}
	}
}

func TestSecretsArePromptedAndNeverCommitted(t *testing.T) {
	parsed, raw := load(t)
	lower := strings.ToLower(raw)
	for _, forbidden := range []string{"mongodb+srv://", "mongodb://", "sk-", "-----begin"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("render.yaml contains forbidden secret shape %q", forbidden)
		}
	}
	if regexp.MustCompile(`\bre_[A-Za-z0-9]{20,}\b`).MatchString(raw) {
		t.Fatal("render.yaml contains a Resend API key shape")
	}
	services := parsed.Projects[0].Environments[0].Services
	for _, candidate := range services {
		for _, variable := range candidate.EnvVars {
			if variable.Key == "MONGODB_URI" || variable.Key == "RESEND_WEBHOOK_SECRET" ||
				variable.Key == "LIVENESS_HMAC_SECRET" ||
				variable.Key == "COMMERCE_HMAC_SECRET" ||
				variable.Key == "ADMIN_HMAC_SECRET" ||
				variable.Key == "OTEL_EXPORTER_OTLP_ENDPOINT" {
				if variable.Sync == nil || *variable.Sync {
					t.Fatalf("%s %s must use sync:false", candidate.Name, variable.Key)
				}
				if variable.Value != "" {
					t.Fatalf("%s %s contains a committed value", candidate.Name, variable.Key)
				}
			}
		}
	}
}

func TestApiAndWorkerUseIndependentCredentialSlots(t *testing.T) {
	parsed, _ := load(t)
	services := parsed.Projects[0].Environments[0].Services
	var apiURI, workerURI *bool
	for _, candidate := range services {
		for index := range candidate.EnvVars {
			variable := &candidate.EnvVars[index]
			if variable.Key != "MONGODB_URI" {
				continue
			}
			switch candidate.Name {
			case "obiara-api-production":
				apiURI = variable.Sync
			case "obiara-worker-production":
				workerURI = variable.Sync
			}
		}
	}
	if apiURI == nil || workerURI == nil || *apiURI || *workerURI {
		t.Fatal("api and worker each need an independent sync:false MongoDB credential slot")
	}
}
