// Package domain models trust-and-safety intake (E12-S01): the one-tap
// report sheet on every surface and member blocks. Reporter identity is
// never disclosed to the reported party (Doc 09 §3) — it is stored only
// for the T&S desk under least-exposure access. Categories map to the
// conduct tiers in Doc 09 §2.
package domain

import (
	"errors"
	"strings"
	"time"
)

// Category is a report class from the universal sheet.
type Category string

const (
	CategoryFraud         Category = "fraud"
	CategoryHarassment    Category = "harassment"
	CategorySexualContent Category = "sexual_content"
	CategoryMinorSafety   Category = "minor_safety"
	CategorySpam          Category = "spam"
	CategoryOther         Category = "other"
)

// Tier is the conduct severity from Doc 09 §2.
type Tier string

const (
	TierA Tier = "A" // account-ending
	TierB Tier = "B" // severe
	TierC Tier = "C" // conduct
	TierD Tier = "D" // care
)

// TierFor maps a report category to its initial triage tier. Panels may
// re-tier on review; fraud operations and minor safety start at Tier A.
func TierFor(category Category) Tier {
	switch category {
	case CategoryFraud, CategoryMinorSafety, CategorySexualContent:
		return TierA
	case CategoryHarassment:
		return TierB
	case CategorySpam, CategoryOther:
		return TierC
	default:
		return TierC
	}
}

// Surface is where the report originated.
type Surface string

const (
	SurfaceRoom    Surface = "room"
	SurfaceDoorway Surface = "doorway"
	SurfacePod     Surface = "pod"
	SurfaceCircle  Surface = "circle"
	SurfaceFire    Surface = "fire"
	SurfaceGame    Surface = "game"
	SurfaceProfile Surface = "profile"
)

var (
	ErrReporterRequired = errors.New("reporter id is required")
	ErrSubjectRequired  = errors.New("reported member id is required")
	ErrSelfReport       = errors.New("members cannot report themselves")
	ErrInvalidCategory  = errors.New("unknown report category")
	ErrInvalidSurface   = errors.New("unknown report surface")
	ErrReasonTooLong    = errors.New("report reason must be 500 characters or fewer")
)

// Report is one filed report.
type Report struct {
	id         string
	reporterID string
	subjectID  string
	category   Category
	tier       Tier
	surface    Surface
	contextRef string
	reason     string
	status     Status
	version    int64
	createdAt  time.Time
}

// Status is the intake lifecycle. Triage queues land with E12-S02.
type Status string

const (
	StatusReceived Status = "received"
)

func NewReport(id, reporterID, subjectID string, category Category, surface Surface, contextRef, reason string, now time.Time) (Report, error) {
	if strings.TrimSpace(reporterID) == "" {
		return Report{}, ErrReporterRequired
	}
	if strings.TrimSpace(subjectID) == "" {
		return Report{}, ErrSubjectRequired
	}
	if reporterID == subjectID {
		return Report{}, ErrSelfReport
	}
	switch category {
	case CategoryFraud, CategoryHarassment, CategorySexualContent, CategoryMinorSafety, CategorySpam, CategoryOther:
	default:
		return Report{}, ErrInvalidCategory
	}
	switch surface {
	case SurfaceRoom, SurfaceDoorway, SurfacePod, SurfaceCircle, SurfaceFire, SurfaceGame, SurfaceProfile:
	default:
		return Report{}, ErrInvalidSurface
	}
	if len(reason) > 500 {
		return Report{}, ErrReasonTooLong
	}
	return Report{
		id:         id,
		reporterID: reporterID,
		subjectID:  subjectID,
		category:   category,
		tier:       TierFor(category),
		surface:    surface,
		contextRef: strings.TrimSpace(contextRef),
		reason:     strings.TrimSpace(reason),
		status:     StatusReceived,
		version:    1,
		createdAt:  now.UTC(),
	}, nil
}

// ReconstituteReport rebuilds a stored report without policy checks.
func ReconstituteReport(id, reporterID, subjectID string, category Category, tier Tier, surface Surface, contextRef, reason string, status Status, version int64, createdAt time.Time) Report {
	return Report{
		id: id, reporterID: reporterID, subjectID: subjectID, category: category, tier: tier,
		surface: surface, contextRef: contextRef, reason: reason, status: status, version: version, createdAt: createdAt,
	}
}

// Acknowledgement is the reporter-safe view: no internals, no subject-facing
// information beyond the reference.
func (report Report) Acknowledgement() (id string, tier Tier, createdAt time.Time) {
	return report.id, report.tier, report.createdAt
}

func (report Report) ID() string           { return report.id }
func (report Report) ReporterID() string   { return report.reporterID }
func (report Report) SubjectID() string    { return report.subjectID }
func (report Report) Category() Category   { return report.category }
func (report Report) Tier() Tier           { return report.tier }
func (report Report) Surface() Surface     { return report.surface }
func (report Report) ContextRef() string   { return report.contextRef }
func (report Report) Reason() string       { return report.reason }
func (report Report) Status() Status       { return report.status }
func (report Report) Version() int64       { return report.version }
func (report Report) CreatedAt() time.Time { return report.createdAt }

// Block is a member-to-member block. Enforcement (hiding, sow prevention)
// composes at the affected contexts; intake stores the edge.
type Block struct {
	blockerID string
	blockedID string
	createdAt time.Time
}

func NewBlock(blockerID, blockedID string, now time.Time) (Block, error) {
	if strings.TrimSpace(blockerID) == "" || strings.TrimSpace(blockedID) == "" || blockerID == blockedID {
		return Block{}, ErrSelfReport
	}
	return Block{blockerID: blockerID, blockedID: blockedID, createdAt: now.UTC()}, nil
}

func (block Block) BlockerID() string    { return block.blockerID }
func (block Block) BlockedID() string    { return block.blockedID }
func (block Block) CreatedAt() time.Time { return block.createdAt }
