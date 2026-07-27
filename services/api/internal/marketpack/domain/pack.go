// Package domain models market packs and their governance (E16-S06;
// plan §8: market_packs + configuration_changes). Packs are market-level
// configuration; publishing requires four eyes — a proposer and a
// different approver — with an immutable configuration change audit.
package domain

import (
	"errors"
	"strings"
	"time"
)

// PackStatus is the pack lifecycle.
type PackStatus string

const (
	StatusDraft     PackStatus = "draft"
	StatusPublished PackStatus = "published"
	StatusRetired   PackStatus = "retired"
)

var (
	ErrPackIDRequired      = errors.New("market pack id is required")
	ErrInvalidMarket       = errors.New("unknown market")
	ErrPackNotDraft        = errors.New("market pack is not a draft")
	ErrPackNotPublished    = errors.New("market pack is not published")
	ErrSelfApproval        = errors.New("approver must differ from proposer (four eyes)")
	ErrActorRequired       = errors.New("governance actor is required")
	ErrTerminologyRequired = errors.New("terminology registry reference is required")
)

// Market is a supported market.
type Market string

const (
	MarketGhanaEnglish Market = "gh_en"
	MarketGhanaTwi     Market = "gh_tw"
	MarketGhanaPidgin  Market = "gh_pidgin"
	MarketGhanaGa      Market = "gh_ga"
)

// MarketPack is one market's configuration bundle.
type MarketPack struct {
	id             string
	market         Market
	terminologyRef string
	features       map[string]bool
	status         PackStatus
	version        int64
	proposedBy     string
	approvedBy     string
	createdAt      time.Time
	publishedAt    *time.Time
}

// NewPack drafts a pack.
func NewPack(id string, market Market, terminologyRef string, features map[string]bool, proposedBy string, now time.Time) (MarketPack, error) {
	if strings.TrimSpace(id) == "" {
		return MarketPack{}, ErrPackIDRequired
	}
	switch market {
	case MarketGhanaEnglish, MarketGhanaTwi, MarketGhanaPidgin, MarketGhanaGa:
	default:
		return MarketPack{}, ErrInvalidMarket
	}
	if strings.TrimSpace(terminologyRef) == "" {
		return MarketPack{}, ErrTerminologyRequired
	}
	if strings.TrimSpace(proposedBy) == "" {
		return MarketPack{}, ErrActorRequired
	}
	return MarketPack{
		id: id, market: market, terminologyRef: terminologyRef, features: features,
		status: StatusDraft, version: 1, proposedBy: proposedBy, createdAt: now.UTC(),
	}, nil
}

// ReconstitutePack rebuilds a stored pack without policy checks.
func ReconstitutePack(id string, market Market, terminologyRef string, features map[string]bool, status PackStatus, version int64, proposedBy, approvedBy string, createdAt time.Time, publishedAt *time.Time) MarketPack {
	return MarketPack{
		id: id, market: market, terminologyRef: terminologyRef, features: features,
		status: status, version: version, proposedBy: proposedBy, approvedBy: approvedBy,
		createdAt: createdAt, publishedAt: publishedAt,
	}
}

// Publish releases a draft pack under four-eyes control (E16-S08 pattern:
// sensitive configuration requires a second approver).
func (pack *MarketPack) Publish(approvedBy string, now time.Time) error {
	if pack.status != StatusDraft {
		return ErrPackNotDraft
	}
	if strings.TrimSpace(approvedBy) == "" {
		return ErrActorRequired
	}
	if approvedBy == pack.proposedBy {
		return ErrSelfApproval
	}
	pack.status = StatusPublished
	pack.approvedBy = approvedBy
	published := now.UTC()
	pack.publishedAt = &published
	pack.version++
	return nil
}

// Retire withdraws a published pack.
func (pack *MarketPack) Retire(actorID string) error {
	if pack.status != StatusPublished {
		return ErrPackNotPublished
	}
	if strings.TrimSpace(actorID) == "" {
		return ErrActorRequired
	}
	pack.status = StatusRetired
	pack.version++
	return nil
}

func (pack MarketPack) ID() string                { return pack.id }
func (pack MarketPack) Market() Market            { return pack.market }
func (pack MarketPack) TerminologyRef() string    { return pack.terminologyRef }
func (pack MarketPack) Features() map[string]bool { return pack.features }
func (pack MarketPack) Status() PackStatus        { return pack.status }
func (pack MarketPack) Version() int64            { return pack.version }
func (pack MarketPack) ProposedBy() string        { return pack.proposedBy }
func (pack MarketPack) ApprovedBy() string        { return pack.approvedBy }
func (pack MarketPack) CreatedAt() time.Time      { return pack.createdAt }
func (pack MarketPack) PublishedAt() *time.Time   { return pack.publishedAt }
