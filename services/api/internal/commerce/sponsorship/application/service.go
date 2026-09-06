// Package application is the sponsorship fund's use cases.
//
// An organization pre-funds a balance and its sponsored codes draw from it.
// Deposits are recorded by an operator rather than collected by a rail: an
// organization pays by whatever means it and Obiara agreed, and somebody
// records what arrived (agent_plan.md §78).
package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/domain"
)

var (
	ErrNotFound    = errors.New("sponsorship fund not found")
	ErrConflict    = errors.New("sponsorship fund changed concurrently")
	ErrUnavailable = errors.New("sponsorship service unavailable")
	// ErrIssuerNotIssuing refuses a fund for a suspended organization.
	ErrIssuerNotIssuing = errors.New("that organization is not issuing")
)

//go:generate mockgen -source=service.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Fund) error
	FindByOrganization(ctx context.Context, organizationID string) (domain.Fund, error)
	Append(ctx context.Context, fund domain.Fund, expected uint64, commandID string) error
	List(ctx context.Context, limit int) ([]domain.Fund, error)
}

// Issuers answers whether an organization is still a live relationship.
type Issuers interface {
	Issuing(ctx context.Context, organizationID string) (bool, error)
}

// Ledger books the money.
//
// A deposit is a liability: the platform is holding somebody else's money
// against seats nobody has taken. A draw turns that liability into revenue,
// because at that point a seat was delivered. Booking a deposit as revenue
// would recognise income for something not yet given.
type Ledger interface {
	RecordDeposit(ctx context.Context, reference string, minor int64, at time.Time) error
	RecordDraw(ctx context.Context, reference string, minor int64, at time.Time) error
}

type IDSource interface{ NewID() string }

type Service struct {
	repository Repository
	issuers    Issuers
	ledger     Ledger
	ids        IDSource
	now        func() time.Time
}

func New(
	repository Repository, issuers Issuers, ledger Ledger, ids IDSource, now func() time.Time,
) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: repository, issuers: issuers, ledger: ledger, ids: ids, now: now}
}

type DepositCommand struct {
	CommandID, OperatorKey, ReasonCode string
	OrganizationID                     string
	AmountPesewas                      int64
}

// Deposit records money an organization has paid, opening its fund if this is
// the first time.
func (service Service) Deposit(
	ctx context.Context, command DepositCommand,
) (domain.Fund, error) {
	if !service.ready() {
		return domain.Fund{}, ErrUnavailable
	}
	issuing, err := service.issuers.Issuing(ctx, strings.TrimSpace(command.OrganizationID))
	if err != nil {
		return domain.Fund{}, ErrUnavailable
	}
	if !issuing {
		return domain.Fund{}, ErrIssuerNotIssuing
	}
	applied := domain.Command{
		ID: strings.TrimSpace(command.CommandID), ActorKey: command.OperatorKey,
		ReasonCode: strings.TrimSpace(command.ReasonCode), At: service.now().UTC(),
	}
	fund, err := service.repository.FindByOrganization(ctx, command.OrganizationID)
	if errors.Is(err, ErrNotFound) {
		opened, openErr := domain.Open(
			service.ids.NewID(), command.OrganizationID,
			domain.Command{
				ID: applied.ID + ":open", ActorKey: applied.ActorKey,
				ReasonCode: applied.ReasonCode, At: applied.At,
			})
		if openErr != nil {
			return domain.Fund{}, openErr
		}
		if createErr := service.repository.Create(ctx, opened); createErr != nil {
			return domain.Fund{}, createErr
		}
		fund = opened
	} else if err != nil {
		return domain.Fund{}, err
	}

	next, err := fund.Deposit(command.AmountPesewas, applied)
	if err != nil {
		return domain.Fund{}, err
	}
	if next.Revision() == fund.Revision() {
		return next, nil
	}
	if err := service.repository.Append(ctx, next, fund.Revision(), applied.ID); err != nil {
		return domain.Fund{}, err
	}
	if service.ledger != nil {
		// A liability, not revenue. The platform is holding money against
		// seats nobody has taken yet.
		_ = service.ledger.RecordDeposit(
			ctx, next.ID()+":"+applied.ID, command.AmountPesewas, applied.At)
	}
	return next, nil
}

// Draw takes the price of one seat from an organization's fund.
//
// Answers (drawn, nil) when it worked and (false, nil) when the fund cannot
// cover it. A fund that is short is not an error: the sponsorship simply does
// not apply, and the member buys their own membership rather than being
// blocked because their employer's balance ran out.
func (service Service) Draw(
	ctx context.Context, organizationID, seatRef string, amountPesewas int64,
) (bool, error) {
	if !service.ready() {
		return false, ErrUnavailable
	}
	fund, err := service.repository.FindByOrganization(ctx, strings.TrimSpace(organizationID))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, ErrUnavailable
	}
	commandID := "draw:" + strings.TrimSpace(seatRef)
	next, err := fund.Draw(seatRef, amountPesewas, domain.Command{
		// Drawn by the system on a member's purchase, so the actor is the
		// fund itself rather than a person. It is still keyed, because an
		// audit trail with a readable actor in one row and digests in the
		// rest is a trail nobody can search consistently.
		ID: commandID, ActorKey: systemActorKey, ReasonCode: "seat_taken",
		At: service.now().UTC(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrInsufficient) || errors.Is(err, domain.ErrFundClosed) {
			return false, nil
		}
		return false, err
	}
	if next.Revision() == fund.Revision() {
		// Already drawn for this seat. Treated as a success, because it is:
		// the seat is paid for.
		return true, nil
	}
	if err := service.repository.Append(ctx, next, fund.Revision(), commandID); err != nil {
		// Somebody else drew between the read and the write. The member pays
		// their own way rather than getting a seat nothing recorded.
		return false, nil
	}
	if service.ledger != nil {
		// The liability becomes revenue here, because this is where a seat is
		// actually delivered.
		_ = service.ledger.RecordDraw(ctx, seatRef, amountPesewas, service.now().UTC())
	}
	return true, nil
}

// Refund returns a drawn seat, for a membership that was reversed.
func (service Service) Refund(ctx context.Context, organizationID, seatRef string) error {
	if !service.ready() {
		return ErrUnavailable
	}
	fund, err := service.repository.FindByOrganization(ctx, strings.TrimSpace(organizationID))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return ErrUnavailable
	}
	commandID := "refund:" + strings.TrimSpace(seatRef)
	next, err := fund.Refund(seatRef, domain.Command{
		ID: commandID, ActorKey: systemActorKey, ReasonCode: "membership_reversed",
		At: service.now().UTC(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotDrawn) {
			return nil
		}
		return err
	}
	if next.Revision() == fund.Revision() {
		return nil
	}
	return service.repository.Append(ctx, next, fund.Revision(), commandID)
}

// FindByOrganization reads a fund, for the operator surface.
func (service Service) FindByOrganization(
	ctx context.Context, organizationID string,
) (domain.Fund, error) {
	if !service.ready() {
		return domain.Fund{}, ErrUnavailable
	}
	return service.repository.FindByOrganization(ctx, strings.TrimSpace(organizationID))
}

func (service Service) List(ctx context.Context, limit int) ([]domain.Fund, error) {
	if !service.ready() {
		return nil, ErrUnavailable
	}
	return service.repository.List(ctx, limit)
}

func (service Service) ready() bool {
	return service.repository != nil && service.issuers != nil &&
		service.ids != nil && service.now != nil
}

// systemActorKey is the actor recorded for a movement nobody made by hand — a
// seat drawn when a member completes a purchase.
//
// A fixed digest rather than a person's, because there is no person, and a
// blank actor in a money trail reads as a bug rather than as a system act.
const systemActorKey = "0000000000000000000000000000000000000000000000000000000000000000"
