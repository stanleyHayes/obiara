package drrehearsal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	MaxRPO = 5 * time.Minute
	MaxRTO = 60 * time.Minute
)

var (
	ErrUnsafePlan        = errors.New("unsafe disaster-recovery rehearsal plan")
	ErrIntegrityMismatch = errors.New("restored data does not match source snapshot")
	ErrApprovalConflict  = errors.New("data-owner and security approvals must be distinct")
	ErrObjectiveMissed   = errors.New("recovery objective missed")
)

type Plan struct {
	RehearsalID string
	Environment string
	Source      string
	Target      string
	PointInTime time.Time
}

func (p Plan) Validate(now time.Time) error {
	if len(p.RehearsalID) < 8 || p.Environment != "staging" ||
		p.Source == "" || !strings.HasPrefix(p.Target, "isolated-") ||
		p.Source == p.Target || p.PointInTime.IsZero() ||
		p.PointInTime.After(now) || now.Sub(p.PointInTime) > MaxRPO {
		return ErrUnsafePlan
	}
	return nil
}

type CollectionFact struct {
	Name        string   `json:"name"`
	Count       int64    `json:"count"`
	Digest      string   `json:"digest"`
	Indexes     []string `json:"indexes"`
	Transaction bool     `json:"transaction"`
	Audit       bool     `json:"audit"`
}

type Snapshot struct {
	Watermark   time.Time        `json:"watermark"`
	Collections []CollectionFact `json:"collections"`
}

func (s Snapshot) CanonicalDigest() (string, error) {
	facts := append([]CollectionFact(nil), s.Collections...)
	sort.Slice(facts, func(i, j int) bool { return facts[i].Name < facts[j].Name })
	for i := range facts {
		facts[i].Indexes = append([]string(nil), facts[i].Indexes...)
		sort.Strings(facts[i].Indexes)
		if facts[i].Name == "" || facts[i].Count < 0 || len(facts[i].Digest) != 64 {
			return "", ErrIntegrityMismatch
		}
	}
	raw, err := json.Marshal(struct {
		Watermark   time.Time        `json:"watermark"`
		Collections []CollectionFact `json:"collections"`
	}{s.Watermark.UTC(), facts})
	if err != nil {
		return "", fmt.Errorf("marshal snapshot: %w", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func Verify(source, restored Snapshot) (string, error) {
	if source.Watermark.IsZero() || !source.Watermark.Equal(restored.Watermark) {
		return "", ErrIntegrityMismatch
	}
	left, err := source.CanonicalDigest()
	if err != nil {
		return "", err
	}
	right, err := restored.CanonicalDigest()
	if err != nil || left != right {
		return "", ErrIntegrityMismatch
	}
	for _, fact := range restored.Collections {
		if !fact.Transaction || !fact.Audit {
			return "", ErrIntegrityMismatch
		}
	}
	return right, nil
}

type Approval struct {
	PrincipalRef string
	EvidenceRef  string
}

type Evidence struct {
	SchemaVersion        int       `json:"schemaVersion"`
	Environment          string    `json:"environment"`
	SyntheticOnly        bool      `json:"syntheticOnly"`
	RehearsalID          string    `json:"rehearsalId"`
	SourceSnapshotAt     time.Time `json:"sourceSnapshotAt"`
	RestoreStartedAt     time.Time `json:"restoreStartedAt"`
	RestoreCompletedAt   time.Time `json:"restoreCompletedAt"`
	Target               string    `json:"target"`
	PointInTime          bool      `json:"pointInTime"`
	ValidationDigest     string    `json:"validationDigest"`
	DataOwnerApprovalRef string    `json:"dataOwnerApprovalRef"`
	SecurityApprovalRef  string    `json:"securityApprovalRef"`
	Result               string    `json:"result"`
	RPOMinutesObserved   int64     `json:"rpoMinutesObserved"`
	RTOMinutesObserved   int64     `json:"rtoMinutesObserved"`
}

func (e Evidence) Digest() (string, error) {
	raw, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
