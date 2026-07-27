package observabilitypolicy

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Policy struct {
	SchemaVersion string `yaml:"schemaVersion"`
	Telemetry     struct {
		Exporter            string   `yaml:"exporter"`
		Transport           string   `yaml:"transport"`
		AllowedDimensions   []string `yaml:"allowedDimensions"`
		ForbiddenDimensions []string `yaml:"forbiddenDimensions"`
	} `yaml:"telemetry"`
	Objectives []Objective `yaml:"objectives"`
	Alerts     []Alert     `yaml:"alerts"`
	Dashboards []Dashboard `yaml:"dashboards"`
}

type Objective struct {
	ID              string  `yaml:"id"`
	Owner           string  `yaml:"owner"`
	Indicator       string  `yaml:"indicator"`
	Target          float64 `yaml:"target"`
	WindowDays      int     `yaml:"windowDays"`
	ReleaseBlocking bool    `yaml:"releaseBlocking"`
}

type Alert struct {
	ID                 string  `yaml:"id"`
	Objective          string  `yaml:"objective"`
	Signal             string  `yaml:"signal"`
	Severity           string  `yaml:"severity"`
	Owner              string  `yaml:"owner"`
	BurnRate           float64 `yaml:"burnRate"`
	LongWindowMinutes  int     `yaml:"longWindowMinutes"`
	ShortWindowMinutes int     `yaml:"shortWindowMinutes"`
	ForMinutes         int     `yaml:"forMinutes"`
	Runbook            string  `yaml:"runbook"`
}

type Dashboard struct {
	ID     string   `yaml:"id"`
	Owner  string   `yaml:"owner"`
	Panels []string `yaml:"panels"`
}

func Load(path string) (Policy, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	var policy Policy
	if err := yaml.Unmarshal(raw, &policy); err != nil {
		return Policy{}, fmt.Errorf("parse observability policy: %w", err)
	}
	return policy, Validate(policy)
}

func Validate(policy Policy) error {
	if policy.SchemaVersion != "obiara.observability.v1" ||
		policy.Telemetry.Exporter != "otlp-http" ||
		policy.Telemetry.Transport != "https" {
		return errors.New("unsupported telemetry contract")
	}
	forbidden := make(map[string]struct{}, len(policy.Telemetry.ForbiddenDimensions))
	for _, dimension := range policy.Telemetry.ForbiddenDimensions {
		forbidden[dimension] = struct{}{}
	}
	for _, dimension := range policy.Telemetry.AllowedDimensions {
		if dimension == "" {
			return errors.New("empty telemetry dimension")
		}
		if _, exists := forbidden[dimension]; exists {
			return fmt.Errorf("forbidden telemetry dimension %q", dimension)
		}
	}
	if len(policy.Objectives) < 4 {
		return errors.New("core, latency, fire and safety objectives are required")
	}
	objectives := make(map[string]Objective, len(policy.Objectives))
	for _, objective := range policy.Objectives {
		if objective.ID == "" || objective.Owner == "" || objective.Indicator == "" ||
			objective.Target <= 0 || objective.Target >= 1 ||
			objective.WindowDays != 30 || !objective.ReleaseBlocking {
			return fmt.Errorf("invalid objective %q", objective.ID)
		}
		if _, exists := objectives[objective.ID]; exists {
			return fmt.Errorf("duplicate objective %q", objective.ID)
		}
		objectives[objective.ID] = objective
	}
	if len(policy.Alerts) < 5 {
		return errors.New("multi-window, dependency, safety and worker alerts are required")
	}
	for _, alert := range policy.Alerts {
		if alert.ID == "" || alert.Owner == "" || alert.Runbook == "" ||
			(alert.Severity != "page" && alert.Severity != "ticket") {
			return fmt.Errorf("invalid alert %q", alert.ID)
		}
		if alert.Objective != "" {
			if _, exists := objectives[alert.Objective]; !exists {
				return fmt.Errorf("alert %q references unknown objective", alert.ID)
			}
			if alert.BurnRate <= 0 || alert.LongWindowMinutes <= alert.ShortWindowMinutes ||
				alert.ShortWindowMinutes <= 0 {
				return fmt.Errorf("alert %q has invalid burn windows", alert.ID)
			}
		} else if alert.Signal == "" || alert.ForMinutes <= 0 {
			return fmt.Errorf("signal alert %q is incomplete", alert.ID)
		}
	}
	for _, dashboard := range policy.Dashboards {
		if dashboard.ID == "" || dashboard.Owner == "" || len(dashboard.Panels) < 3 {
			return fmt.Errorf("dashboard %q is incomplete", dashboard.ID)
		}
	}
	return nil
}

func ErrorBudgetMinutes(target float64, windowDays int) int {
	if target <= 0 || target >= 1 || windowDays <= 0 {
		return 0
	}
	return int((1 - target) * float64(windowDays*24*60))
}
