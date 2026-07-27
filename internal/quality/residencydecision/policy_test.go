package residencydecision

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

var decisionReviewTime = time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve repository root")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func fixturePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repositoryRoot(t), "deploy", "legal", "examples", "residency-dpia.synthetic.json")
}

func completeRecord() Record {
	issued := decisionReviewTime.Add(-time.Hour)
	expires := decisionReviewTime.Add(30 * 24 * time.Hour)
	signed := issued.Add(5 * time.Minute)
	record := Record{
		SchemaVersion:         SchemaVersion,
		DecisionID:            "decision/external-residency-review",
		Environment:           "production",
		Market:                "GH",
		CandidateSHA:          strings.Repeat("1", 40),
		Interpretation:        "ghana-only",
		IssuerKind:            "external-authority",
		IssuerRef:             "role/external-issuer",
		ProcessingScopeRefs:   ProcessingScopes(),
		TransferAssessmentRef: strings.Repeat("a", 64),
		DPIARef:               strings.Repeat("b", 64),
		ResidualRiskOutcome:   "accepted",
		IssuedAt:              &issued,
		ExpiresAt:             &expires,
		Outcome:               "approved",
	}
	record.AuthorityRefs = AuthorityRefs{
		Founder: Authority{
			ActorRef:             "role/founder-authority",
			SignatureEvidenceRef: strings.Repeat("c", 64),
			SignedAt:             &signed,
		},
		DPOLegal: Authority{
			ActorRef:             "role/dpo-legal-authority",
			SignatureEvidenceRef: strings.Repeat("d", 64),
			SignedAt:             &signed,
		},
		IndependentReviewer: Authority{
			ActorRef:             "role/independent-reviewer",
			SignatureEvidenceRef: strings.Repeat("e", 64),
			SignedAt:             &signed,
		},
	}
	for index, boundary := range LocationBoundaries() {
		classes := []string{"C2"}
		if boundary != "logs-observability" {
			classes = []string{"C2", "C3", "C4"}
		}
		record.ScopeLocations = append(record.ScopeLocations, ScopeLocation{
			Boundary:    boundary,
			ProviderRef: "provider/external-evidence",
			CountryCode: "GH",
			RegionCode:  "gh-accra-" + string(rune('a'+index)),
			DataClasses: classes,
		})
	}
	return record
}

func TestBlockedSyntheticFixtureCannotSupplyLegalDecision(t *testing.T) {
	record, decision, err := Load(fixturePath(t), decisionReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Eligible || !record.Synthetic {
		t.Fatalf("synthetic fixture passed: %#v", decision)
	}
	for _, required := range []string{
		"repository-issued",
		"synthetic",
		"outcome-pending",
		"residual-risk-pending",
		"transfer-assessment-missing",
		"dpia-missing",
		"decision-unsigned-or-undated",
		"founder-unsigned",
		"dpo-legal-unsigned",
		"independent-reviewer-unsigned",
		"processing-scope-missing-commerce-and-financial-ledger",
		"location-missing-cdn-cache",
		"location-missing-ai-vendor-processing",
		"location-missing-identity-provider-processing",
	} {
		if !slicesContain(decision.Blockers, required) {
			t.Errorf("missing blocker %q in %#v", required, decision.Blockers)
		}
	}
}

func TestCompleteExternalDecisionMetadataCanSatisfyOnlyThisPolicyGate(t *testing.T) {
	decision, err := Evaluate(completeRecord(), decisionReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Eligible || len(decision.Blockers) != 0 {
		t.Fatalf("complete metadata blocked: %#v", decision)
	}
}

func TestUnsafeDecisionEvidenceFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Record)
	}{
		{"unsigned founder", func(record *Record) { record.AuthorityRefs.Founder = Authority{} }},
		{"self approved", func(record *Record) {
			record.AuthorityRefs.IndependentReviewer.ActorRef = record.AuthorityRefs.Founder.ActorRef
		}},
		{"stale", func(record *Record) {
			expired := decisionReviewTime.Add(-time.Second)
			record.ExpiresAt = &expired
		}},
		{"synthetic", func(record *Record) { record.Synthetic = true }},
		{"incomplete processing scope", func(record *Record) {
			record.ProcessingScopeRefs = record.ProcessingScopeRefs[1:]
		}},
		{"repository issued", func(record *Record) { record.IssuerKind = "repository" }},
		{"incomplete location map", func(record *Record) { record.ScopeLocations = record.ScopeLocations[1:] }},
		{"ghana only mismatch", func(record *Record) { record.ScopeLocations[0].CountryCode = "ZA" }},
		{"ambiguous africa region", func(record *Record) {
			record.Interpretation = "africa-region"
			record.ScopeLocations[0].CountryCode = "ZA"
			record.ScopeLocations[0].RegionCode = "africa"
		}},
		{"out of africa primary compute", func(record *Record) {
			record.Interpretation = "africa-region"
			record.ScopeLocations[0].CountryCode = "DE"
			record.ScopeLocations[0].RegionCode = "eu-central-1"
		}},
		{"missing transfer assessment", func(record *Record) { record.TransferAssessmentRef = "" }},
		{"missing dpia", func(record *Record) { record.DPIARef = "" }},
		{"pending residual risk", func(record *Record) { record.ResidualRiskOutcome = "pending" }},
		{"pending decision", func(record *Record) { record.Outcome = "pending" }},
		{"replayed signature evidence", func(record *Record) {
			record.AuthorityRefs.IndependentReviewer.SignatureEvidenceRef =
				record.AuthorityRefs.Founder.SignatureEvidenceRef
		}},
		{"replayed DPIA as signature", func(record *Record) {
			record.AuthorityRefs.IndependentReviewer.SignatureEvidenceRef = record.DPIARef
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record := completeRecord()
			test.mutate(&record)
			decision, err := Evaluate(record, decisionReviewTime)
			if err == nil && decision.Eligible {
				t.Fatal("unsafe decision evidence passed")
			}
		})
	}
}

func TestMalformedPrivateAndAmbiguousMetadataIsRejected(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Record)
	}{
		{"wrong environment", func(record *Record) { record.Environment = "staging" }},
		{"ambiguous interpretation", func(record *Record) { record.Interpretation = "ghana-or-africa" }},
		{"secret actor", func(record *Record) { record.IssuerRef = "role/secret-owner" }},
		{"email actor", func(record *Record) { record.IssuerRef = "person@example.com" }},
		{"URL provider", func(record *Record) {
			record.ScopeLocations[0].ProviderRef = "https://provider.invalid"
		}},
		{"unknown data class", func(record *Record) {
			record.ScopeLocations[0].DataClasses = []string{"C5"}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record := completeRecord()
			test.mutate(&record)
			decision, err := Evaluate(record, decisionReviewTime)
			if err == nil && decision.Eligible {
				t.Fatal("malformed or private metadata passed")
			}
		})
	}
}

func TestStrictLoaderRejectsUnknownPrivateFields(t *testing.T) {
	raw, err := os.ReadFile(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document["legalOpinion"] = "must never be stored here"
	mutated, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "unsafe.json")
	if err := os.WriteFile(path, mutated, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(path, decisionReviewTime); err == nil {
		t.Fatal("unknown legal/private field was accepted")
	}
}

func TestLocationAndScopeOrderCannotChangeDecision(t *testing.T) {
	record := completeRecord()
	baseline, err := Evaluate(record, decisionReviewTime)
	if err != nil {
		t.Fatal(err)
	}
	if err := quick.Check(func(seed uint64) bool {
		candidate := record
		candidate.ProcessingScopeRefs = append([]string(nil), record.ProcessingScopeRefs...)
		candidate.ScopeLocations = append([]ScopeLocation(nil), record.ScopeLocations...)
		shuffle(seed, candidate.ProcessingScopeRefs)
		shuffle(seed^0x9e3779b97f4a7c15, candidate.ScopeLocations)
		decision, err := Evaluate(candidate, decisionReviewTime)
		return err == nil && reflect.DeepEqual(decision, baseline)
	}, &quick.Config{MaxCount: 1000}); err != nil {
		t.Fatal(err)
	}
}

func TestSchemaDeclaresStableDecisionPacketFields(t *testing.T) {
	path := filepath.Join(repositoryRoot(t), "deploy", "legal", "residency-dpia.schema.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{
		"decisionId", "interpretation", "authorityRefs", "scopeLocations",
		"transferAssessmentRef", "dpiaRef", "residualRiskOutcome",
		"issuedAt", "expiresAt", "synthetic", "outcome",
	} {
		if !slicesContain(schema.Required, field) {
			t.Fatalf("schema missing required field %q", field)
		}
	}
}

func slicesContain(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func shuffle[T any](seed uint64, values []T) {
	for index := len(values) - 1; index > 0; index-- {
		other := int(seed % uint64(index+1))
		values[index], values[other] = values[other], values[index]
		seed = seed*6364136223846793005 + 1
	}
}
