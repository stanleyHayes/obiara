package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	whatsappdomain "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/domain"
)

// Sentinel errors.
var (
	ErrDuplicateNomination = errors.New("a pending nomination already exists for this kin")
	ErrNominationNotFound  = errors.New("nomination not found")
)

// Clock supplies the decision time.
type Clock func() time.Time

// NominationRepository persists nominations.
type NominationRepository interface {
	Create(ctx context.Context, n domain.Nomination) error
	FindByID(ctx context.Context, id string) (domain.Nomination, error)
	Update(ctx context.Context, n domain.Nomination) error
	ListByMember(ctx context.Context, memberID string) ([]domain.Nomination, error)
	HasPendingForKin(ctx context.Context, memberID, kinPhone string) (bool, error)
}

// NotificationSender sends the kin consent invite.
type NotificationSender interface {
	Send(ctx context.Context, msg whatsappdomain.Message) error
}

// NominationService implements FR-1302: a member nominates a trusted kin
// (aunt/uncle/mother/father/elder) to be their Nnoboa companion. The kin
// receives a consent invite over WhatsApp; only after explicit consent do
// they become an active companion. Decline is always respected without
// consequence. Nominations expire after 30 days. The kin sees nothing about
// the member's romantic life — only contact and consent.
type NominationService struct {
	repo   NominationRepository
	sender NotificationSender
	clock  Clock
}

// NewNominationService constructs the service.
func NewNominationService(repo NominationRepository, sender NotificationSender, clock Clock) *NominationService {
	return &NominationService{repo: repo, sender: sender, clock: clock}
}

// NominateInput carries a new nomination.
type NominateInput struct {
	MemberID     string
	KinName      string
	KinPhone     string
	Relationship string
}

// Nominate validates and persists a nomination, then sends the consent invite.
func (s *NominationService) Nominate(ctx context.Context, in NominateInput) (domain.Nomination, error) {
	n, err := domain.NewNomination(in.MemberID, in.KinName, in.KinPhone, in.Relationship, s.clock().UTC())
	if err != nil {
		return domain.Nomination{}, err
	}
	exists, err := s.repo.HasPendingForKin(ctx, n.MemberID, n.KinPhone)
	if err != nil {
		return domain.Nomination{}, err
	}
	if exists {
		return domain.Nomination{}, ErrDuplicateNomination
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return domain.Nomination{}, err
	}
	// Invite text is deliberately free of any mention of dating, romance, or
	// the member's journey — the kin consents to be a companion, nothing more.
	msg, err := whatsappdomain.NewNnoboaConsentMessage(n.KinPhone, n.KinName)
	if err != nil {
		return domain.Nomination{}, fmt.Errorf("build consent invite: %w", err)
	}
	err = s.sender.Send(ctx, msg)
	if err != nil {
		return domain.Nomination{}, fmt.Errorf("send consent invite: %w", err)
	}
	return n, nil
}

// Get returns a nomination, applying lazy expiry.
func (s *NominationService) Get(ctx context.Context, id string) (domain.Nomination, error) {
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Nomination{}, err
	}
	return s.applyExpiry(ctx, n)
}

// ListForMember lists a member's nominations (latest first), applying lazy expiry.
func (s *NominationService) ListForMember(ctx context.Context, memberID string) ([]domain.Nomination, error) {
	ns, err := s.repo.ListByMember(ctx, memberID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Nomination, 0, len(ns))
	for _, n := range ns {
		n, err = s.applyExpiry(ctx, n)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

// Consent records kin consent (FR-1302 consent gate).
func (s *NominationService) Consent(ctx context.Context, id string) (domain.Nomination, error) {
	return s.transition(ctx, id, func(n *domain.Nomination, now time.Time) error { return n.Consent(now) })
}

// Decline records kin decline. Always respected without consequence.
func (s *NominationService) Decline(ctx context.Context, id string) (domain.Nomination, error) {
	return s.transition(ctx, id, func(n *domain.Nomination, now time.Time) error { return n.Decline(now) })
}

func (s *NominationService) transition(ctx context.Context, id string, f func(*domain.Nomination, time.Time) error) (domain.Nomination, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return domain.Nomination{}, err
	}
	if err := f(&n, s.clock().UTC()); err != nil {
		return domain.Nomination{}, err
	}
	if err := s.repo.Update(ctx, n); err != nil {
		return domain.Nomination{}, err
	}
	return n, nil
}

func (s *NominationService) applyExpiry(ctx context.Context, n domain.Nomination) (domain.Nomination, error) {
	if n.Status != domain.StatusPending || !n.ExpiredAt(s.clock().UTC()) {
		return n, nil
	}
	if err := n.Expire(s.clock().UTC()); err != nil {
		return domain.Nomination{}, err
	}
	if err := s.repo.Update(ctx, n); err != nil {
		return domain.Nomination{}, err
	}
	return n, nil
}
