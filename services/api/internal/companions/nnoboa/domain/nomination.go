// Package domain defines the Nnoboa nomination aggregate (E13-S06, FR-1302).
// A member nominates a trusted kin — aunt, uncle, mother, father, or elder —
// as their Nnoboa companion. The kin receives a consent invite; only after
// explicit consent do they become an active companion. Decline is always
// respected without consequence. The kin sees nothing about the member's
// romantic life: a nomination carries contact and relationship only.
package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

// NominationExpiry is how long a kin has to respond before the nomination lapses.
const NominationExpiry = 30 * 24 * time.Hour

var (
	ErrInvalidNomination = errors.New("invalid nomination")
	ErrNotPending        = errors.New("nomination is not pending")
	e164                 = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	opaque               = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:-]{7,127}$`)
)

// Relationship is the member's kin relation to the nominee.
type Relationship string

const (
	Aunt   Relationship = "aunt"
	Uncle  Relationship = "uncle"
	Mother Relationship = "mother"
	Father Relationship = "father"
	Elder  Relationship = "elder"
)

var allowedRelationship = map[Relationship]bool{
	Aunt: true, Uncle: true, Mother: true, Father: true, Elder: true,
}

// Status is the nomination lifecycle state.
type Status string

const (
	StatusPending   Status = "pending"
	StatusConsented Status = "consented"
	StatusDeclined  Status = "declined"
	StatusExpired   Status = "expired"
)

// Nomination is one member's nomination of one kin as Nnoboa companion.
type Nomination struct {
	ID           string
	MemberID     string
	KinName      string
	KinPhone     string
	Relationship Relationship
	Status       Status
	Version      int64
	CreatedAt    time.Time
	RespondedAt  *time.Time
}

// NewNomination validates inputs and opens a pending nomination.
func NewNomination(memberID, kinName, kinPhone, relationship string, now time.Time) (Nomination, error) {
	rel := Relationship(relationship)
	if !opaque.MatchString(memberID) || strings.TrimSpace(kinName) == "" ||
		!e164.MatchString(kinPhone) || !allowedRelationship[rel] || now.IsZero() {
		return Nomination{}, ErrInvalidNomination
	}
	var idBytes [16]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		return Nomination{}, err
	}
	return Nomination{
		ID:           "nom_" + hex.EncodeToString(idBytes[:]),
		MemberID:     memberID,
		KinName:      strings.TrimSpace(kinName),
		KinPhone:     kinPhone,
		Relationship: rel,
		Status:       StatusPending,
		Version:      1,
		CreatedAt:    now.UTC(),
	}, nil
}

// ExpiredAt reports whether a pending nomination has lapsed at now.
func (n Nomination) ExpiredAt(now time.Time) bool {
	return n.Status == StatusPending && !now.Before(n.CreatedAt.Add(NominationExpiry))
}

// Consent records the kin's explicit consent.
func (n *Nomination) Consent(now time.Time) error {
	return n.respond(StatusConsented, now)
}

// Decline records the kin's decline. Always respected, without consequence.
func (n *Nomination) Decline(now time.Time) error {
	return n.respond(StatusDeclined, now)
}

// Expire lapses a pending nomination past its window.
func (n *Nomination) Expire(now time.Time) error {
	if n.Status != StatusPending {
		return ErrNotPending
	}
	if now.Before(n.CreatedAt.Add(NominationExpiry)) {
		return ErrInvalidNomination
	}
	return n.respond(StatusExpired, now)
}

func (n *Nomination) respond(status Status, now time.Time) error {
	if n.Status != StatusPending {
		return ErrNotPending
	}
	if now.IsZero() {
		return ErrInvalidNomination
	}
	n.Status = status
	t := now.UTC()
	n.RespondedAt = &t
	n.Version++
	return nil
}
