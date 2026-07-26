package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidPurpose  = errors.New("invalid consent purpose")
	ErrPurposeInactive = errors.New("consent purpose is not active")
)

type PurposeKind string

const (
	PurposePromise PurposeKind = "promise"
	PurposeTerms   PurposeKind = "terms"
	PurposeAge     PurposeKind = "age"
)

type PurposeStatus string

const (
	PurposeActive  PurposeStatus = "active"
	PurposeRetired PurposeStatus = "retired"
)

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[._-][a-z0-9]+)*$`)

// Purpose is an immutable, versioned declaration of what a subject is asked
// to consent to. Text is represented by a content reference so translated or
// regulated copy remains outside the consent record.
type Purpose struct {
	id             string
	kind           PurposeKind
	version        uint64
	contentRef     string
	status         PurposeStatus
	effectiveSince time.Time
}

type NewPurposeParams struct {
	ID             string
	Kind           PurposeKind
	Version        uint64
	ContentRef     string
	Status         PurposeStatus
	EffectiveSince time.Time
}

func NewPurpose(params NewPurposeParams) (Purpose, error) {
	params.ID = strings.TrimSpace(params.ID)
	params.ContentRef = strings.TrimSpace(params.ContentRef)
	if !slugPattern.MatchString(params.ID) || !validOpaqueReference(params.ContentRef) ||
		params.Version == 0 || params.EffectiveSince.IsZero() {
		return Purpose{}, ErrInvalidPurpose
	}
	switch params.Kind {
	case PurposePromise, PurposeTerms, PurposeAge:
	default:
		return Purpose{}, ErrInvalidPurpose
	}
	switch params.Status {
	case PurposeActive, PurposeRetired:
	default:
		return Purpose{}, ErrInvalidPurpose
	}
	return Purpose{
		id:             params.ID,
		kind:           params.Kind,
		version:        params.Version,
		contentRef:     params.ContentRef,
		status:         params.Status,
		effectiveSince: params.EffectiveSince.UTC(),
	}, nil
}

func (purpose Purpose) ID() string                { return purpose.id }
func (purpose Purpose) Kind() PurposeKind         { return purpose.kind }
func (purpose Purpose) Version() uint64           { return purpose.version }
func (purpose Purpose) ContentRef() string        { return purpose.contentRef }
func (purpose Purpose) Status() PurposeStatus     { return purpose.status }
func (purpose Purpose) EffectiveSince() time.Time { return purpose.effectiveSince }
func (purpose Purpose) IsActive(at time.Time) bool {
	return purpose.status == PurposeActive && !at.UTC().Before(purpose.effectiveSince)
}
