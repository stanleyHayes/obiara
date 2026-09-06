// Package application is the discount code's use cases.
package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/domain"
)

type Service struct {
	repository Repository
	issuers    Issuers
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func New(
	repository Repository, issuers Issuers, keyer Keyer, ids IDSource, now func() time.Time,
) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: repository, issuers: issuers, keyer: keyer, ids: ids, now: now}
}

type IssueCommand struct {
	CommandID        string
	Code             string
	IssuerID         string
	SKUID            string
	Shape            domain.Shape
	Amount           int64
	StartsAt, EndsAt time.Time
	Cap              uint32
}

// Issue creates a code in an organization's name.
func (service Service) Issue(
	ctx context.Context, command IssueCommand,
) (domain.Promotion, error) {
	if !service.ready() {
		return domain.Promotion{}, ErrUnavailable
	}
	issuing, err := service.issuers.Issuing(ctx, strings.TrimSpace(command.IssuerID))
	if err != nil {
		return domain.Promotion{}, ErrUnavailable
	}
	if !issuing {
		return domain.Promotion{}, ErrIssuerNotIssuing
	}
	promotion, err := domain.Issue(
		service.ids.NewID(), command.Code, command.IssuerID, command.SKUID,
		command.Shape, command.Amount, command.StartsAt, command.EndsAt, command.Cap,
		domain.Command{ID: strings.TrimSpace(command.CommandID), At: service.now().UTC()},
	)
	if err != nil {
		return domain.Promotion{}, err
	}
	if err := service.repository.Create(ctx, promotion); err != nil {
		return domain.Promotion{}, err
	}
	return promotion, nil
}

// Withdraw stops a code being used. What it already paid for is untouched.
func (service Service) Withdraw(
	ctx context.Context, code, commandID string,
) (domain.Promotion, error) {
	if !service.ready() {
		return domain.Promotion{}, ErrUnavailable
	}
	current, err := service.repository.FindByCode(ctx, normalise(code))
	if err != nil {
		return domain.Promotion{}, err
	}
	next, err := current.Withdraw(domain.Command{
		ID: strings.TrimSpace(commandID), At: service.now().UTC(),
	})
	if err != nil {
		return domain.Promotion{}, err
	}
	if next.Revision() == current.Revision() {
		return next, nil
	}
	if err := service.repository.Append(ctx, next, current.Revision(), commandID); err != nil {
		return domain.Promotion{}, err
	}
	return next, nil
}

// Applied is a code that has been spent on one purchase.
type Applied struct {
	Code string
	// DiscountMinor is what came off. Zero means no code was given, which is
	// the ordinary case and not a failure.
	DiscountMinor int64
}

// Apply spends a redemption and reports what comes off the price.
//
// Redeemed here rather than when the payment settles, on purpose. Waiting
// would let two members both be quoted the last slot and both pay, and the
// second would be charged a discounted price with no redemption recorded
// against it — a discount the books cannot account for. Reserving instead
// means an abandoned payment burns a slot, which is bounded by the cap the
// issuer chose and is the smaller problem.
//
// A code that does not apply is not an error to the caller. The purchase goes
// ahead at full price, because a member who mistyped a code should be able to
// buy the thing they came to buy.
func (service Service) Apply(
	ctx context.Context, code, memberID, skuID string, priceMinor int64, commandID string,
) (Applied, error) {
	if strings.TrimSpace(code) == "" {
		return Applied{}, nil
	}
	if !service.ready() {
		return Applied{}, ErrUnavailable
	}
	promotion, err := service.repository.FindByCode(ctx, normalise(code))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Applied{}, nil
		}
		return Applied{}, ErrUnavailable
	}
	// Membership passes only, and only the one this code names. A code for
	// another SKU is not a discount on this one.
	if promotion.SKUID() != strings.TrimSpace(skuID) {
		return Applied{}, nil
	}
	memberKey, err := service.keyer.MemberKey(memberID)
	if err != nil {
		return Applied{}, ErrUnavailable
	}
	now := service.now().UTC()
	if err := promotion.Redeemable(memberKey, now); err != nil {
		return Applied{}, nil
	}
	redeemed, err := promotion.Redeem(memberKey, domain.Command{
		ID: strings.TrimSpace(commandID), At: now,
	})
	if err != nil {
		return Applied{}, nil
	}
	if redeemed.Revision() != promotion.Revision() {
		if err := service.repository.Append(
			ctx, redeemed, promotion.Revision(), commandID,
		); err != nil {
			// Somebody else took the slot between the read and the write. The
			// purchase goes ahead at full price rather than at a discount
			// nothing recorded.
			return Applied{}, nil
		}
	}
	return Applied{Code: promotion.Code(), DiscountMinor: promotion.Discount(priceMinor)}, nil
}

// ListByIssuer is the operator view: how many of an organization's codes are
// in use. It never says which members used them.
func (service Service) ListByIssuer(
	ctx context.Context, issuerID string, limit int,
) ([]domain.Promotion, error) {
	if !service.ready() {
		return nil, ErrUnavailable
	}
	return service.repository.ListByIssuer(ctx, strings.TrimSpace(issuerID), limit)
}

func (service Service) ready() bool {
	return service.repository != nil && service.issuers != nil &&
		service.keyer != nil && service.ids != nil && service.now != nil
}

func normalise(code string) string { return strings.ToUpper(strings.TrimSpace(code)) }
