// Package domain models embers: the one warm introduction each attendee
// may give after a fire (FR-402: one per attendee per fire, co-attendees
// only, redeemable within 24 h). A mutual ember opens a doorway instantly
// (Doc 06 S-65).
package domain

import (
	"errors"
	"time"
)

// EmberLifetime is the redemption window (FR-402).
const EmberLifetime = 24 * time.Hour

// EmberStatus is the ember lifecycle.
type EmberStatus string

const (
	StatusIssued   EmberStatus = "issued"
	StatusMutual   EmberStatus = "mutual"
	StatusRedeemed EmberStatus = "redeemed"
	StatusExpired  EmberStatus = "expired"
)

var (
	ErrEmberIDRequired = errors.New("ember id is required")
	ErrSelfEmber       = errors.New("embers cannot be given to oneself")
	ErrEmberExpired    = errors.New("ember redemption window has closed")
	ErrEmberNotOpen    = errors.New("ember is not open for redemption")
)

// Ember is one directed warm introduction between co-attendees.
type Ember struct {
	id         string
	fireID     string
	fromID     string
	toID       string
	status     EmberStatus
	expiresAt  time.Time
	version    int64
	createdAt  time.Time
	redeemedAt *time.Time
}

func NewEmber(id, fireID, fromID, toID string, now time.Time) (Ember, error) {
	if id == "" {
		return Ember{}, ErrEmberIDRequired
	}
	if fromID == "" || toID == "" || fromID == toID {
		return Ember{}, ErrSelfEmber
	}
	now = now.UTC()
	return Ember{
		id:        id,
		fireID:    fireID,
		fromID:    fromID,
		toID:      toID,
		status:    StatusIssued,
		expiresAt: now.Add(EmberLifetime),
		version:   1,
		createdAt: now,
	}, nil
}

// ReconstituteEmber rebuilds a stored ember without policy checks.
func ReconstituteEmber(id, fireID, fromID, toID string, status EmberStatus, expiresAt time.Time, version int64, createdAt time.Time, redeemedAt *time.Time) Ember {
	return Ember{id: id, fireID: fireID, fromID: fromID, toID: toID, status: status, expiresAt: expiresAt, version: version, createdAt: createdAt, redeemedAt: redeemedAt}
}

// MarkMutual records that the reverse ember also exists (Doc 06 S-65:
// mutual embers open a doorway instantly).
func (ember *Ember) MarkMutual() {
	if ember.status == StatusIssued {
		ember.status = StatusMutual
		ember.version++
	}
}

// Redeem closes the ember within its window.
func (ember *Ember) Redeem(now time.Time) error {
	if ember.status != StatusIssued && ember.status != StatusMutual {
		return ErrEmberNotOpen
	}
	if ember.Expired(now) {
		ember.status = StatusExpired
		ember.version++
		return ErrEmberExpired
	}
	ember.status = StatusRedeemed
	redeemed := now.UTC()
	ember.redeemedAt = &redeemed
	ember.version++
	return nil
}

func (ember Ember) Expired(now time.Time) bool {
	return !now.UTC().Before(ember.expiresAt)
}

func (ember Ember) ID() string             { return ember.id }
func (ember Ember) FireID() string         { return ember.fireID }
func (ember Ember) FromID() string         { return ember.fromID }
func (ember Ember) ToID() string           { return ember.toID }
func (ember Ember) Status() EmberStatus    { return ember.status }
func (ember Ember) ExpiresAt() time.Time   { return ember.expiresAt }
func (ember Ember) Version() int64         { return ember.version }
func (ember Ember) CreatedAt() time.Time   { return ember.createdAt }
func (ember Ember) RedeemedAt() *time.Time { return ember.redeemedAt }
