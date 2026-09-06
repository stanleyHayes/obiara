// Package application is the affiliate's use cases.
//
// Commission accrues on a qualified conversion and never on a signup. The
// owner set the rule (agent_plan.md §41): the referred member reaches Tier 1,
// is still there after a waiting period, and has no upheld safety finding.
// Getting that wrong in the generous direction is not a rounding error — it is
// paying people to bring accounts the safety model exists to keep out.
package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/domain"
)

// WaitingPeriod is how long a referral is pending before it can qualify.
//
// Thirty days, because "signed up" and "stayed" are different things and only
// the second is worth paying for. It is a constant rather than configuration:
// a rate is a commercial lever, but the shape of what counts as a conversion
// is a safety property, and one that can be turned down to a day is not one.
const WaitingPeriod = 30 * 24 * time.Hour

const (
	outcomeQualified = "qualified"
	outcomeRefused   = "refused"
)

type Service struct {
	repository    Repository
	referrals     Referrals
	qualification Qualification
	ledger        Ledger
	keyer         Keyer
	ids           IDSource
	// commissionPesewas is a flat amount per qualified conversion.
	//
	// Flat rather than a percentage, by the owner's decision: it is
	// predictable to budget, cannot compound with any other take rate into a
	// loss, and — unlike a percentage — tells an affiliate nothing about what
	// any individual member paid.
	commissionPesewas int64
	now               func() time.Time
}

func New(
	repository Repository, referrals Referrals, qualification Qualification,
	ledger Ledger, keyer Keyer, ids IDSource, commissionPesewas int64, now func() time.Time,
) Service {
	if now == nil {
		now = time.Now
	}
	return Service{
		repository: repository, referrals: referrals, qualification: qualification,
		ledger: ledger, keyer: keyer, ids: ids,
		commissionPesewas: commissionPesewas, now: now,
	}
}

type RegisterCommand struct {
	CommandID, ReasonCode string
	Name, Code, Email     string
}

// Register records an affiliate.
//
// MemberIDs is the check that keeps members out. It is passed rather than
// looked up because whether a given identifier is a member is the member
// context's question, and this one has no business browsing it.
func (service Service) Register(
	ctx context.Context, command RegisterCommand, isMember func(context.Context, string) (bool, error),
) (domain.Affiliate, error) {
	if !service.ready() {
		return domain.Affiliate{}, ErrUnavailable
	}
	if isMember == nil {
		// A missing check is not permission. Without it there is no way to
		// know an affiliate is not a member, and that is the one rule this
		// context exists to hold.
		return domain.Affiliate{}, ErrUnavailable
	}
	member, err := isMember(ctx, strings.TrimSpace(command.Email))
	if err != nil {
		return domain.Affiliate{}, ErrUnavailable
	}
	if member {
		return domain.Affiliate{}, ErrMemberAffiliate
	}
	affiliate, err := domain.Register(
		service.ids.NewID(), command.Name, command.Code, command.Email,
		domain.Command{
			ID:         strings.TrimSpace(command.CommandID),
			ReasonCode: strings.TrimSpace(command.ReasonCode), At: service.now().UTC(),
		},
	)
	if err != nil {
		return domain.Affiliate{}, err
	}
	if err := service.repository.Create(ctx, affiliate); err != nil {
		return domain.Affiliate{}, err
	}
	return affiliate, nil
}

// Attribute records that a code brought a member.
//
// Nothing accrues here. The referral waits out its period and is then asked
// the qualification questions — which is the whole difference between paying
// for a signup and paying for somebody who stayed.
//
// A code that does not resolve is not an error. Somebody mistyping a code at
// signup must not be stopped from joining.
func (service Service) Attribute(ctx context.Context, code, memberID string) error {
	if strings.TrimSpace(code) == "" {
		return nil
	}
	if !service.ready() {
		return ErrUnavailable
	}
	affiliate, err := service.repository.FindByCode(
		ctx, strings.ToUpper(strings.TrimSpace(code)))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return ErrUnavailable
	}
	if !affiliate.Earning() {
		return nil
	}
	memberKey, err := service.keyer.MemberKey(memberID)
	if err != nil {
		return ErrUnavailable
	}
	// If this member was already attributed to somebody, the first one keeps
	// them. Otherwise a second code typed later would move a referral that
	// somebody has already been waiting thirty days on.
	if _, err := service.referrals.Find(ctx, memberKey); err == nil {
		return nil
	}
	now := service.now().UTC()
	return service.referrals.Record(ctx, Referral{
		MemberKey: memberKey, MemberID: strings.TrimSpace(memberID),
		AffiliateID: affiliate.ID(), Code: affiliate.Code(),
		RecordedAt: now, QualifiesAt: now.Add(WaitingPeriod),
	})
}

// QualifyDue settles every referral whose waiting period has elapsed.
//
// Run as a sweep rather than triggered by the member, because the thing being
// checked is that thirty days passed and nothing went wrong in them — which is
// an absence, and nothing emits an event for an absence.
//
// Returns how many accrued and how many were refused.
func (service Service) QualifyDue(ctx context.Context, limit int) (accrued int, refused int, err error) {
	if !service.ready() || service.qualification == nil {
		return 0, 0, ErrUnavailable
	}
	now := service.now().UTC()
	due, err := service.referrals.DueForQualification(ctx, now, limit)
	if err != nil {
		return 0, 0, ErrUnavailable
	}
	for _, referral := range due {
		qualified, checkErr := service.qualifies(ctx, referral.MemberID)
		if checkErr != nil {
			// A question that cannot be answered is not a conversion, and it
			// is not a refusal either: leaving it pending means the next
			// sweep asks again rather than writing down a guess.
			continue
		}
		if !qualified {
			if err := service.referrals.MarkSettled(ctx, referral.MemberKey, outcomeRefused); err == nil {
				refused++
			}
			continue
		}
		if err := service.accrue(ctx, referral); err != nil {
			continue
		}
		if err := service.referrals.MarkSettled(ctx, referral.MemberKey, outcomeQualified); err != nil {
			// The accrual landed and the mark did not. The aggregate refuses
			// a second accrual for the same referral, so the next sweep is
			// safe: it will try, be told it is already counted, and settle.
			continue
		}
		accrued++
	}
	return accrued, refused, nil
}

// qualifies asks the three questions. All of them, and any "no" is a no.
func (service Service) qualifies(ctx context.Context, memberID string) (bool, error) {
	verified, err := service.qualification.Verified(ctx, memberID)
	if err != nil {
		return false, err
	}
	if !verified {
		return false, nil
	}
	clean, err := service.qualification.Clean(ctx, memberID)
	if err != nil {
		return false, err
	}
	return clean, nil
}

func (service Service) accrue(ctx context.Context, referral Referral) error {
	affiliate, err := service.repository.FindByID(ctx, referral.AffiliateID)
	if err != nil {
		return err
	}
	commandID := "accrue:" + referral.MemberKey
	next, err := affiliate.Accrue(referral.MemberKey, service.commissionPesewas, domain.Command{
		ID: commandID, ReasonCode: "qualified_conversion", At: service.now().UTC(),
	})
	if err != nil {
		return err
	}
	if next.Revision() == affiliate.Revision() {
		return nil
	}
	if err := service.repository.Append(ctx, next, affiliate.Revision(), commandID); err != nil {
		return err
	}
	if service.ledger != nil {
		// Owed the moment it accrues, whether or not anybody has asked to be
		// paid. A liability that only appears at payout time is a liability
		// the books did not know about.
		_ = service.ledger.RecordAccrual(
			ctx, commandID, service.commissionPesewas, service.now().UTC())
	}
	return nil
}

// ClawBack reverses an accrual for a referral that should not have counted —
// a refunded membership, or a member removed for conduct.
func (service Service) ClawBack(ctx context.Context, memberID, reasonCode string) error {
	if !service.ready() {
		return ErrUnavailable
	}
	memberKey, err := service.keyer.MemberKey(memberID)
	if err != nil {
		return ErrUnavailable
	}
	referral, err := service.referrals.Find(ctx, memberKey)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Nobody was paid for this member. Nothing to reverse.
			return nil
		}
		return ErrUnavailable
	}
	if referral.Settled != outcomeQualified {
		return nil
	}
	affiliate, err := service.repository.FindByID(ctx, referral.AffiliateID)
	if err != nil {
		return ErrUnavailable
	}
	commandID := "clawback:" + memberKey
	next, err := affiliate.ClawBack(memberKey, service.commissionPesewas, domain.Command{
		ID: commandID, ReasonCode: strings.TrimSpace(reasonCode), At: service.now().UTC(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotCounted) {
			return nil
		}
		return err
	}
	if next.Revision() == affiliate.Revision() {
		return nil
	}
	if err := service.repository.Append(ctx, next, affiliate.Revision(), commandID); err != nil {
		return err
	}
	if service.ledger != nil {
		_ = service.ledger.RecordClawback(
			ctx, commandID, service.commissionPesewas, service.now().UTC())
	}
	return nil
}

// Find reads one affiliate, for the operator surface and the payout desk.
func (service Service) Find(ctx context.Context, id string) (domain.Affiliate, error) {
	if !service.ready() {
		return domain.Affiliate{}, ErrUnavailable
	}
	return service.repository.FindByID(ctx, strings.TrimSpace(id))
}

func (service Service) List(ctx context.Context, limit int) ([]domain.Affiliate, error) {
	if !service.ready() {
		return nil, ErrUnavailable
	}
	return service.repository.List(ctx, limit)
}

// Commission is the flat amount a qualified conversion earns.
func (service Service) Commission() int64 { return service.commissionPesewas }

func (service Service) ready() bool {
	return service.repository != nil && service.referrals != nil &&
		service.keyer != nil && service.ids != nil && service.now != nil &&
		service.commissionPesewas > 0
}
