package releasebundle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"
)

var reference = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]{2,79}$`)
var sha = regexp.MustCompile(`^[0-9a-f]{40}$`)

type Bundle struct {
	SchemaVersion string `json:"schemaVersion"`
	Environment   string `json:"environment"`
	Candidate     struct {
		CommitSHA          string `json:"commitSha"`
		ReleaseEvidenceRef string `json:"releaseEvidenceRef"`
	} `json:"candidate"`
	Notes struct {
		Summary    string   `json:"summary"`
		ChangeRefs []string `json:"changeRefs"`
	} `json:"notes"`
	UAT struct {
		Invited      int    `json:"invited"`
		Consented    int    `json:"consented"`
		Trained      int    `json:"trained"`
		Completed    int    `json:"completed"`
		CriticalOpen int    `json:"criticalOpen"`
		EvidenceRef  string `json:"evidenceRef"`
	} `json:"uat"`
	Rollback struct {
		RehearsalRef string `json:"rehearsalRef"`
		KnownGoodSHA string `json:"knownGoodSha"`
		RTOMinutes   int    `json:"rtoMinutes"`
	} `json:"rollback"`
	Hypercare struct {
		PlatformOwner  string   `json:"platformOwner"`
		SafetyOwner    string   `json:"safetyOwner"`
		UATOwner       string   `json:"uatOwner"`
		CadenceMinutes int      `json:"cadenceMinutes"`
		Blockers       []string `json:"blockers"`
	} `json:"hypercare"`
	Approvals struct {
		PreparedBy         string `json:"preparedBy"`
		ReviewedBy         string `json:"reviewedBy"`
		ProductionApproved bool   `json:"productionApproved"`
	} `json:"approvals"`
	GeneratedAt time.Time `json:"generatedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Disposition string    `json:"disposition"`
}

func Load(path string, now time.Time) (Bundle, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Bundle{}, err
	}
	var bundle Bundle
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bundle); err != nil {
		return Bundle{}, fmt.Errorf("parse release bundle: %w", err)
	}
	return bundle, Validate(bundle, now)
}

func Validate(bundle Bundle, now time.Time) error {
	if bundle.SchemaVersion != "obiara.release-bundle.v1" ||
		(bundle.Environment != "preview" && bundle.Environment != "staging" && bundle.Environment != "production") {
		return errors.New("unsupported release bundle")
	}
	if !sha.MatchString(bundle.Candidate.CommitSHA) || !reference.MatchString(bundle.Candidate.ReleaseEvidenceRef) ||
		len(bundle.Notes.Summary) < 12 || len(bundle.Notes.Summary) > 240 ||
		len(bundle.Notes.ChangeRefs) == 0 || len(bundle.Notes.ChangeRefs) > 20 {
		return errors.New("candidate or notes are incomplete")
	}
	if bundle.UAT.Invited < 1 || bundle.UAT.Consented < 0 ||
		bundle.UAT.Consented > bundle.UAT.Invited ||
		bundle.UAT.Trained > bundle.UAT.Consented ||
		bundle.UAT.Completed > bundle.UAT.Trained ||
		bundle.UAT.CriticalOpen < 0 || !reference.MatchString(bundle.UAT.EvidenceRef) {
		return errors.New("invalid UAT aggregate")
	}
	if !reference.MatchString(bundle.Rollback.RehearsalRef) ||
		!sha.MatchString(bundle.Rollback.KnownGoodSHA) ||
		bundle.Rollback.RTOMinutes < 1 || bundle.Rollback.RTOMinutes > 60 {
		return errors.New("rollback evidence is incomplete")
	}
	if !reference.MatchString(bundle.Hypercare.PlatformOwner) ||
		!reference.MatchString(bundle.Hypercare.SafetyOwner) ||
		!reference.MatchString(bundle.Hypercare.UATOwner) ||
		bundle.Hypercare.CadenceMinutes < 15 || bundle.Hypercare.CadenceMinutes > 1440 {
		return errors.New("hypercare ownership is incomplete")
	}
	if bundle.Approvals.PreparedBy == bundle.Approvals.ReviewedBy ||
		!reference.MatchString(bundle.Approvals.PreparedBy) ||
		!reference.MatchString(bundle.Approvals.ReviewedBy) {
		return errors.New("distinct bounded approvals are required")
	}
	if bundle.ExpiresAt.Sub(bundle.GeneratedAt) > 7*24*time.Hour ||
		!bundle.ExpiresAt.After(bundle.GeneratedAt) || now.After(bundle.ExpiresAt) {
		return errors.New("release bundle is stale")
	}
	blocked := bundle.UAT.CriticalOpen > 0 || len(bundle.Hypercare.Blockers) > 0
	if bundle.Environment == "production" {
		if !bundle.Approvals.ProductionApproved || blocked || bundle.Disposition != "production-approved" {
			return errors.New("production release is not approved")
		}
		return errors.New("production topology and repository gates remain blocked")
	}
	if blocked && bundle.Disposition != "blocked" {
		return errors.New("blocking evidence must fail closed")
	}
	if !blocked && bundle.Environment != "production" && bundle.Disposition != "qualified-non-production" {
		return errors.New("invalid non-production disposition")
	}
	return nil
}
