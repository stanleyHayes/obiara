// Package providerdiligence validates metadata-only procurement evidence.
// It cannot contact, select, purchase from, or configure a provider.
package providerdiligence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"time"
)

const SchemaVersion = "obiara.provider-diligence.v1"

var (
	hex64   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	roleRef = regexp.MustCompile(`^role/[a-z][a-z0-9-]{2,63}$`)
)

type Registry struct {
	SchemaVersion string     `json:"schemaVersion"`
	GeneratedAt   time.Time  `json:"generatedAt"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	Environment   string     `json:"environment"`
	Market        string     `json:"market"`
	Providers     []Provider `json:"providers"`
}

type Provider struct {
	ID          string     `json:"id"`
	Outcome     string     `json:"outcome"`
	Synthetic   bool       `json:"synthetic"`
	IssuerRef   string     `json:"issuerRef"`
	ReviewerRef string     `json:"reviewerRef"`
	Evidence    []Evidence `json:"evidence"`
}

type Evidence struct {
	Subject     string    `json:"subject"`
	EvidenceRef string    `json:"evidenceRef"`
	CollectedAt time.Time `json:"collectedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Outcome     string    `json:"outcome"`
	Synthetic   bool      `json:"synthetic"`
}

type ProviderResult struct {
	ID       string   `json:"id"`
	Ready    bool     `json:"ready"`
	Blockers []string `json:"blockers"`
}

type Decision struct {
	Ready     bool             `json:"ready"`
	Providers []ProviderResult `json:"providers"`
	Blockers  []string         `json:"blockers"`
}

func ProviderIDs() []string {
	return []string{"atlas", "communications", "livekit", "storage-cdn"}
}

func Subjects() []string {
	return []string{"breach", "cost", "deletion", "dpa", "exit", "keys", "locations", "retention", "subprocessors", "support"}
}

func Load(path string, now time.Time) (Registry, Decision, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, Decision{}, err
	}
	var registry Registry
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&registry); err != nil {
		return Registry{}, Decision{}, fmt.Errorf("parse provider diligence registry: %w", err)
	}
	var extra any
	if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Registry{}, Decision{}, errors.New("provider diligence registry contains trailing data")
	}
	decision, err := Evaluate(registry, now)
	return registry, decision, err
}

func Evaluate(registry Registry, now time.Time) (Decision, error) {
	decision := Decision{}
	if registry.SchemaVersion != SchemaVersion || registry.Environment != "production" || registry.Market != "GH" ||
		now.IsZero() || registry.GeneratedAt.IsZero() || !registry.ExpiresAt.After(registry.GeneratedAt) ||
		registry.ExpiresAt.Sub(registry.GeneratedAt) > 24*time.Hour || now.Before(registry.GeneratedAt.Add(-5*time.Minute)) || now.After(registry.ExpiresAt) {
		return decision, errors.New("invalid or stale provider diligence registry")
	}
	providers := make(map[string]Provider, len(registry.Providers))
	refs := map[string]bool{}
	for _, provider := range registry.Providers {
		if _, exists := providers[provider.ID]; exists {
			return decision, fmt.Errorf("duplicate provider %q", provider.ID)
		}
		if !slices.Contains(ProviderIDs(), provider.ID) {
			return decision, fmt.Errorf("unknown provider %q", provider.ID)
		}
		if !roleRef.MatchString(provider.IssuerRef) || !roleRef.MatchString(provider.ReviewerRef) || provider.IssuerRef == provider.ReviewerRef {
			return decision, fmt.Errorf("invalid authority separation for %q", provider.ID)
		}
		seenSubjects := map[string]bool{}
		for _, evidence := range provider.Evidence {
			if seenSubjects[evidence.Subject] {
				return decision, fmt.Errorf("duplicate subject %q for %q", evidence.Subject, provider.ID)
			}
			if !slices.Contains(Subjects(), evidence.Subject) {
				return decision, fmt.Errorf("unknown subject %q for %q", evidence.Subject, provider.ID)
			}
			if !hex64.MatchString(evidence.EvidenceRef) || refs[evidence.EvidenceRef] {
				return decision, errors.New("invalid or duplicate evidence reference")
			}
			seenSubjects[evidence.Subject], refs[evidence.EvidenceRef] = true, true
		}
		providers[provider.ID] = provider
	}
	for _, id := range ProviderIDs() {
		result := ProviderResult{ID: id}
		provider, exists := providers[id]
		if !exists {
			result.Blockers = append(result.Blockers, "provider-missing")
		} else {
			if provider.Synthetic {
				result.Blockers = append(result.Blockers, "provider-synthetic")
			}
			if provider.Outcome != "approved" {
				result.Blockers = append(result.Blockers, "provider-outcome-"+provider.Outcome)
			}
			bySubject := map[string]Evidence{}
			for _, evidence := range provider.Evidence {
				bySubject[evidence.Subject] = evidence
			}
			for _, subject := range Subjects() {
				evidence, ok := bySubject[subject]
				if !ok {
					result.Blockers = append(result.Blockers, "evidence-missing-"+subject)
					continue
				}
				if evidence.Synthetic {
					result.Blockers = append(result.Blockers, "evidence-synthetic-"+subject)
				}
				if evidence.Outcome != "accepted" {
					result.Blockers = append(result.Blockers, "evidence-outcome-"+subject+"-"+evidence.Outcome)
				}
				if evidence.CollectedAt.IsZero() || !evidence.ExpiresAt.After(evidence.CollectedAt) ||
					evidence.ExpiresAt.Sub(evidence.CollectedAt) > 90*24*time.Hour ||
					evidence.CollectedAt.After(now.Add(5*time.Minute)) || now.After(evidence.ExpiresAt) {
					result.Blockers = append(result.Blockers, "evidence-stale-"+subject)
				}
			}
		}
		result.Ready = len(result.Blockers) == 0
		if !result.Ready {
			decision.Blockers = append(decision.Blockers, id)
		}
		decision.Providers = append(decision.Providers, result)
	}
	decision.Ready = len(decision.Blockers) == 0
	return decision, nil
}
