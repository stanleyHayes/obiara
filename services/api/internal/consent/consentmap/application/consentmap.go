// Package application runs the consent map (Doc 08 §8): member toggles
// within each purpose's control level, immutable receipts for every
// change, and the enforcement ports other contexts consult.
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/domain"
)

var ErrMemberIDRequired = errors.New("member id is required")

// StateStore persists explicit per-member consent choices.
type StateStore interface {
	Get(ctx context.Context, memberID string, purpose domain.Purpose) (*bool, error)
	Set(ctx context.Context, memberID string, purpose domain.Purpose, enabled bool) error
	AllForMember(ctx context.Context, memberID string) (map[domain.Purpose]bool, error)
}

// ReceiptStore persists immutable consent receipts.
type ReceiptStore interface {
	Append(context.Context, domain.Receipt) error
}

// ConsentMapService resolves and manages consent states.
type ConsentMapService struct {
	states   StateStore
	receipts ReceiptStore
	now      func() time.Time
	newID    func() string
}

func NewConsentMapService(states StateStore, receipts ReceiptStore, now func() time.Time, newID func() string) ConsentMapService {
	return ConsentMapService{states: states, receipts: receipts, now: now, newID: newID}
}

// StateFor resolves the member's effective consent for a purpose
// (explicit choice, else the purpose default).
func (service ConsentMapService) StateFor(ctx context.Context, memberID string, purpose domain.Purpose) (bool, error) {
	explicit, err := service.states.Get(ctx, memberID, purpose)
	if err != nil {
		return false, err
	}
	return domain.State(purpose, explicit)
}

// Set applies a member's change, validating the purpose's control level
// and writing an immutable receipt.
func (service ConsentMapService) Set(ctx context.Context, memberID string, purpose domain.Purpose, enable bool) (bool, error) {
	if memberID == "" {
		return false, ErrMemberIDRequired
	}
	if err := domain.ValidateChange(purpose, enable); err != nil {
		return false, err
	}
	if err := service.states.Set(ctx, memberID, purpose, enable); err != nil {
		return false, err
	}
	if err := service.receipts.Append(ctx, domain.Receipt{
		ID: service.newID(), MemberID: memberID, Purpose: purpose, Enabled: enable, CreatedAt: service.now().UTC(),
	}); err != nil {
		return false, err
	}
	return enable, nil
}

// Switchboard returns the member's full effective state for the
// switchboard view.
func (service ConsentMapService) Switchboard(ctx context.Context, memberID string) (map[domain.Purpose]bool, error) {
	explicit, err := service.states.AllForMember(ctx, memberID)
	if err != nil {
		return nil, err
	}
	board := domain.Purposes()
	for purpose, enabled := range explicit {
		board[purpose] = enabled
	}
	return board, nil
}
