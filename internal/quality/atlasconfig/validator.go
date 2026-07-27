package atlasconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Spec struct {
	APIVersion      string          `yaml:"apiVersion"`
	Kind            string          `yaml:"kind"`
	Metadata        Metadata        `yaml:"metadata"`
	Activation      Activation      `yaml:"activation"`
	Topology        Topology        `yaml:"topology"`
	Network         Network         `yaml:"network"`
	Encryption      Encryption      `yaml:"encryption"`
	Identities      []Identity      `yaml:"identities"`
	RoleDefinitions map[string]Role `yaml:"roleDefinitions"`
	Backup          Backup          `yaml:"backup"`
	Restore         Restore         `yaml:"restore"`
}
type Metadata struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	DataPolicy  string `yaml:"dataPolicy"`
}
type Activation struct {
	State               string `yaml:"state"`
	ResidencyDecision   string `yaml:"residencyDecision"`
	DPIAApproval        string `yaml:"dpiaApproval"`
	ProcurementApproval string `yaml:"procurementApproval"`
	RestoreEvidence     string `yaml:"restoreEvidence"`
}
type Topology struct {
	Provider              string `yaml:"provider"`
	Region                string `yaml:"region"`
	MongoMajor            string `yaml:"mongoMajor"`
	MinimumTier           string `yaml:"minimumTier"`
	ElectableNodes        int    `yaml:"electableNodes"`
	AvailabilityZones     int    `yaml:"availabilityZones"`
	TerminationProtection bool   `yaml:"terminationProtection"`
}
type Network struct {
	PublicAccess                         bool     `yaml:"publicAccess"`
	MinimumTLS                           string   `yaml:"minimumTLS"`
	AllowlistRefs                        []string `yaml:"allowlistRefs"`
	PrivateEndpointRequiredWhenColocated bool     `yaml:"privateEndpointRequiredWhenColocated"`
}
type Encryption struct {
	AtRest                string `yaml:"atRest"`
	Transport             string `yaml:"transport"`
	CustomerManagedKeyRef string `yaml:"customerManagedKeyRef"`
	C4FieldEncryption     string `yaml:"c4FieldEncryption"`
	C4KeyRef              string `yaml:"c4KeyRef"`
}
type Identity struct {
	Name          string   `yaml:"name"`
	CredentialRef string   `yaml:"credentialRef"`
	Roles         []string `yaml:"roles"`
}
type Role struct {
	Actions     []string `yaml:"actions"`
	Collections []string `yaml:"collections"`
}
type Backup struct {
	Continuous          bool      `yaml:"continuous"`
	PointInTimeRecovery bool      `yaml:"pointInTimeRecovery"`
	Region              string    `yaml:"region"`
	RPOMinutes          int       `yaml:"rpoMinutes"`
	RTOMinutes          int       `yaml:"rtoMinutes"`
	Schedules           Schedules `yaml:"schedules"`
}
type Schedules struct {
	DailyDays     int `yaml:"dailyDays"`
	WeeklyWeeks   int `yaml:"weeklyWeeks"`
	MonthlyMonths int `yaml:"monthlyMonths"`
}
type Restore struct {
	DestructiveInPlace     bool   `yaml:"destructiveInPlace"`
	IsolatedTargetRequired bool   `yaml:"isolatedTargetRequired"`
	EvidenceSchema         string `yaml:"evidenceSchema"`
	EvidenceRef            string `yaml:"evidenceRef"`
}
type Evidence struct {
	SchemaVersion                                                       int       `json:"schemaVersion"`
	Environment                                                         string    `json:"environment"`
	SyntheticOnly                                                       bool      `json:"syntheticOnly"`
	RehearsalID                                                         string    `json:"rehearsalId"`
	SourceSnapshotAt                                                    time.Time `json:"sourceSnapshotAt"`
	RestoreStartedAt                                                    time.Time `json:"restoreStartedAt"`
	RestoreCompletedAt                                                  time.Time `json:"restoreCompletedAt"`
	Target                                                              string    `json:"target"`
	PointInTime                                                         bool      `json:"pointInTime"`
	ValidationDigest, DataOwnerApprovalRef, SecurityApprovalRef, Result string
	RPOMinutesObserved                                                  int `json:"rpoMinutesObserved"`
	RTOMinutesObserved                                                  int `json:"rtoMinutesObserved"`
}

func Load(path string) (Spec, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, nil, err
	}
	var spec Spec
	if err = yaml.Unmarshal(raw, &spec); err != nil {
		return Spec{}, nil, err
	}
	return spec, raw, nil
}
func ValidateFile(root, path string) error {
	spec, raw, err := Load(path)
	if err != nil {
		return err
	}
	if err = Validate(spec, string(raw)); err != nil {
		return err
	}
	if spec.Restore.EvidenceRef != "" {
		evidencePath := spec.Restore.EvidenceRef
		if !filepath.IsAbs(evidencePath) {
			evidencePath = filepath.Join(root, evidencePath)
		}
		return ValidateEvidenceFile(evidencePath, spec.Metadata.Environment)
	}
	return nil
}
func Validate(s Spec, raw string) error {
	var problems []string
	add := func(ok bool, message string) {
		if !ok {
			problems = append(problems, message)
		}
	}
	add(s.APIVersion == "obiara.platform/v1" && s.Kind == "AtlasEnvironment", "apiVersion/kind must be pinned")
	add(s.Metadata.Environment == "staging" || s.Metadata.Environment == "production", "environment must be staging or production")
	add(s.Topology.Provider == "AWS" && s.Topology.Region == "AF_SOUTH_1", "topology must use reviewed Africa-region candidate")
	add(s.Topology.MongoMajor == "8.0" && s.Topology.MinimumTier == "M10", "MongoDB and minimum backup-capable tier must be pinned")
	add(s.Topology.ElectableNodes >= 3 && s.Topology.AvailabilityZones >= 3 && s.Topology.TerminationProtection, "three-zone protected replica topology required")
	add(!s.Network.PublicAccess && s.Network.MinimumTLS == "1.2" && s.Network.PrivateEndpointRequiredWhenColocated, "TLS, no public access and colocated private endpoint required")
	add(len(s.Network.AllowlistRefs) >= 2 && unique(s.Network.AllowlistRefs), "at least two distinct allowlist references required")
	for _, ref := range s.Network.AllowlistRefs {
		add(ref != "" && !strings.Contains(ref, "0.0.0.0") && !strings.ContainsAny(ref, "*/"), "wildcard or empty allowlist reference forbidden")
	}
	add(s.Encryption.AtRest == "AES-256" && s.Encryption.Transport == "TLS" && s.Encryption.CustomerManagedKeyRef != "" && s.Encryption.C4FieldEncryption == "required" && s.Encryption.C4KeyRef != "", "at-rest, transport and C4 field encryption required")
	add(len(s.Identities) >= 6, "separate API, worker, C4, backup and restore identities required")
	credentials := []string{}
	names := []string{}
	assignments := map[string]int{}
	for _, identity := range s.Identities {
		credentials = append(credentials, identity.CredentialRef)
		names = append(names, identity.Name)
		add(identity.Name != "" && identity.CredentialRef != "" && len(identity.Roles) == 1, "identity must have one explicit role")
		for _, role := range identity.Roles {
			assignments[role]++
			_, exists := s.RoleDefinitions[role]
			add(exists, "identity references unknown role "+role)
		}
	}
	add(unique(credentials) && unique(names), "identity names and credential references must be unique")
	for _, required := range []string{"app-api", "app-worker", "c4-identity", "c4-safety", "backup-operator", "restore-operator"} {
		_, ok := s.RoleDefinitions[required]
		add(ok, "missing role "+required)
		add(assignments[required] == 1, "role must be assigned to exactly one identity: "+required)
	}
	for name, role := range s.RoleDefinitions {
		add(len(role.Actions) > 0 && len(role.Collections) > 0, "empty role "+name)
		for _, action := range role.Actions {
			add(!slices.Contains([]string{"atlasAdmin", "readWriteAnyDatabase", "dbAdminAnyDatabase", "root"}, action), "broad role action forbidden")
		}
		if name == "app-api" || name == "app-worker" {
			for _, collection := range role.Collections {
				add(!slices.Contains([]string{"identity_verifications", "identity_bindings", "device_risk", "reports", "safety_cases", "safety_evidence", "suban_events"}, collection), "C4 collection leaked into general application role")
			}
		}
	}
	add(slices.Equal(s.RoleDefinitions["backup-operator"].Actions, []string{"backup-read"}), "backup operator must be read-only")
	add(slices.Equal(s.RoleDefinitions["restore-operator"].Actions, []string{"restore-to-isolated-target"}), "restore operator must target isolation only")
	add(s.Backup.Continuous && s.Backup.PointInTimeRecovery && s.Backup.Region == s.Topology.Region, "localized continuous PITR required")
	add(s.Backup.RPOMinutes > 0 && s.Backup.RPOMinutes <= 5 && s.Backup.RTOMinutes > 0 && s.Backup.RTOMinutes <= 60, "RPO/RTO exceed ADR limits")
	add(s.Backup.Schedules.DailyDays >= 7 && s.Backup.Schedules.WeeklyWeeks >= 4 && s.Backup.Schedules.MonthlyMonths >= 13, "backup retention schedule too short")
	add(!s.Restore.DestructiveInPlace && s.Restore.IsolatedTargetRequired && s.Restore.EvidenceSchema == "deploy/atlas/restore-evidence.schema.json", "isolated non-destructive restore contract required")
	lower := strings.ToLower(raw)
	for _, forbidden := range []string{"mongodb://", "mongodb+srv://", "0.0.0.0/0", "password:", "private_key", "-----begin"} {
		add(!strings.Contains(lower, forbidden), "secret or wildcard shape forbidden: "+forbidden)
	}
	if s.Metadata.Environment == "staging" {
		add(s.Metadata.DataPolicy == "synthetic-only" && s.Activation.State == "synthetic-only", "staging must be synthetic-only")
		add(s.Restore.EvidenceRef == "deploy/atlas/evidence/staging.synthetic.json", "staging synthetic restore evidence required")
	} else {
		add(s.Activation.State == "blocked", "production must remain blocked")
		add(strings.HasPrefix(s.Activation.ResidencyDecision, "pending-") && strings.HasPrefix(s.Activation.DPIAApproval, "pending-") && strings.HasPrefix(s.Activation.ProcurementApproval, "pending-") && strings.HasPrefix(s.Activation.RestoreEvidence, "pending-"), "production gates must remain explicitly pending")
		add(s.Restore.EvidenceRef == "", "production must not claim restore evidence")
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}
func ValidateEvidenceFile(path, environment string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var evidence Evidence
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&evidence); err != nil {
		return err
	}
	hex64 := regexp.MustCompile(`^[a-f0-9]{64}$`)
	if evidence.SchemaVersion != 1 || evidence.Environment != environment || environment == "staging" && !evidence.SyntheticOnly || !strings.HasPrefix(evidence.Target, "isolated-") || !evidence.PointInTime || evidence.Result != "pass" ||
		!hex64.MatchString(evidence.ValidationDigest) || !hex64.MatchString(evidence.DataOwnerApprovalRef) || !hex64.MatchString(evidence.SecurityApprovalRef) || evidence.DataOwnerApprovalRef == evidence.SecurityApprovalRef ||
		evidence.SourceSnapshotAt.IsZero() || !evidence.RestoreStartedAt.After(evidence.SourceSnapshotAt) || !evidence.RestoreCompletedAt.After(evidence.RestoreStartedAt) ||
		evidence.RPOMinutesObserved < 0 || evidence.RPOMinutesObserved > 5 || evidence.RTOMinutesObserved < 0 || evidence.RTOMinutesObserved > 60 {
		return fmt.Errorf("invalid restore evidence")
	}
	return nil
}
func unique(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if value == "" || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}
