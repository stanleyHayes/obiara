package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/domain"
	"go.uber.org/mock/gomock"
)

var now = time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)

type fixedID string

func (i fixedID) NewID() string { return string(i) }

func operator() string { return strings.Repeat("a", 64) }

func fundFor(t *testing.T, pesewas int64) domain.Fund {
	t.Helper()
	opened, err := domain.Open("fund_1", "org_1", domain.Command{
		ID: "cmd_1:open", ActorKey: operator(), ReasonCode: "partner_agreement", At: now})
	if err != nil {
		t.Fatal(err)
	}
	if pesewas == 0 {
		return opened
	}
	funded, err := opened.Deposit(pesewas, domain.Command{
		ID: "cmd_1", ActorKey: operator(), ReasonCode: "bank_transfer_received", At: now})
	if err != nil {
		t.Fatal(err)
	}
	return funded
}

func service(ctrl *gomock.Controller, repository *MockRepository, issuing bool) Service {
	issuers := NewMockIssuers(ctrl)
	issuers.EXPECT().Issuing(gomock.Any(), gomock.Any()).Return(issuing, nil).AnyTimes()
	return New(repository, issuers, nil, fixedID("fund_1"), func() time.Time { return now })
}

func TestAFirstDepositOpensTheFund(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByOrganization(gomock.Any(), "org_1").Return(domain.Fund{}, ErrNotFound)
	repository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "cmd_2").DoAndReturn(
		func(_ context.Context, fund domain.Fund, _ uint64, _ string) error {
			if fund.Balance() != 100_000 {
				t.Fatalf("balance = %d", fund.Balance())
			}
			return nil
		})

	fund, err := service(ctrl, repository, true).Deposit(context.Background(), DepositCommand{
		CommandID: "cmd_2", OperatorKey: operator(), ReasonCode: "bank_transfer_received",
		OrganizationID: "org_1", AmountPesewas: 100_000,
	})
	if err != nil || fund.Balance() != 100_000 {
		t.Fatalf("fund = %#v, err = %v", fund.Balance(), err)
	}
}

func TestNoFundIsOpenedForASuspendedOrganization(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	// No Create and no Append: nothing may be written.

	if _, err := service(ctrl, repository, false).Deposit(
		context.Background(), DepositCommand{
			CommandID: "cmd_2", OperatorKey: operator(), ReasonCode: "bank_transfer_received",
			OrganizationID: "org_1", AmountPesewas: 100_000,
		}); !errors.Is(err, ErrIssuerNotIssuing) {
		t.Fatalf("err = %v, want ErrIssuerNotIssuing", err)
	}
}

func TestASeatIsDrawnFromTheBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	fund := fundFor(t, 100_000)
	repository.EXPECT().FindByOrganization(gomock.Any(), "org_1").Return(fund, nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), fund.Revision(), "draw:purchase_1").
		DoAndReturn(func(_ context.Context, drawn domain.Fund, _ uint64, _ string) error {
			if drawn.Balance() != 95_000 || drawn.Seats() != 1 {
				t.Fatalf("balance %d over %d seats", drawn.Balance(), drawn.Seats())
			}
			return nil
		})

	drawn, err := service(ctrl, repository, true).
		Draw(context.Background(), "org_1", "purchase_1", 5_000)
	if err != nil || !drawn {
		t.Fatalf("drawn = %v, err = %v", drawn, err)
	}
}

func TestAFundThatIsShortDoesNotBlockTheMember(t *testing.T) {
	// The member buys their own membership rather than being stopped because
	// their employer's balance ran out. Not an error — the sponsorship simply
	// does not apply.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByOrganization(gomock.Any(), "org_1").Return(fundFor(t, 1_000), nil)
	// No Append expectation: nothing is written.

	drawn, err := service(ctrl, repository, true).
		Draw(context.Background(), "org_1", "purchase_1", 5_000)
	if err != nil {
		t.Fatalf("a short fund refused the purchase: %v", err)
	}
	if drawn {
		t.Fatal("a short fund reported paying for a seat")
	}
}

func TestAnOrganizationWithNoFundAtAllDoesNotBlockTheMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByOrganization(gomock.Any(), "org_1").Return(domain.Fund{}, ErrNotFound)

	drawn, err := service(ctrl, repository, true).
		Draw(context.Background(), "org_1", "purchase_1", 5_000)
	if err != nil || drawn {
		t.Fatalf("drawn = %v, err = %v", drawn, err)
	}
}

func TestLosingTheRaceForTheLastSeatMeansTheMemberPays(t *testing.T) {
	// Two members draw the last seat at once and only one write lands. The
	// loser pays their own way rather than getting a seat nothing recorded.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	fund := fundFor(t, 5_000)
	repository.EXPECT().FindByOrganization(gomock.Any(), "org_1").Return(fund, nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), fund.Revision(), gomock.Any()).
		Return(ErrConflict)

	drawn, err := service(ctrl, repository, true).
		Draw(context.Background(), "org_1", "purchase_1", 5_000)
	if err != nil {
		t.Fatalf("a lost race refused the purchase: %v", err)
	}
	if drawn {
		t.Fatal("a lost race reported paying for a seat")
	}
}

func TestDrawingTheSameSeatTwiceIsStillOneSeat(t *testing.T) {
	// A retried purchase reaches the same seat reference. Reporting success
	// is correct: the seat is paid for.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	drawn, err := fundFor(t, 100_000).Draw("purchase_1", 5_000, domain.Command{
		ID: "draw:purchase_1", ActorKey: systemActorKey, ReasonCode: "seat_taken", At: now})
	if err != nil {
		t.Fatal(err)
	}
	repository.EXPECT().FindByOrganization(gomock.Any(), "org_1").Return(drawn, nil)
	// No Append expectation: nothing more is written.

	paid, err := service(ctrl, repository, true).
		Draw(context.Background(), "org_1", "purchase_1", 5_000)
	if err != nil || !paid {
		t.Fatalf("paid = %v, err = %v", paid, err)
	}
}

func TestARefundReturnsTheSeatToTheBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	drawn, err := fundFor(t, 100_000).Draw("purchase_1", 5_000, domain.Command{
		ID: "draw:purchase_1", ActorKey: systemActorKey, ReasonCode: "seat_taken", At: now})
	if err != nil {
		t.Fatal(err)
	}
	repository.EXPECT().FindByOrganization(gomock.Any(), "org_1").Return(drawn, nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), drawn.Revision(), "refund:purchase_1").
		DoAndReturn(func(_ context.Context, fund domain.Fund, _ uint64, _ string) error {
			if fund.Balance() != 100_000 {
				t.Fatalf("balance = %d after a refund", fund.Balance())
			}
			return nil
		})

	if err := service(ctrl, repository, true).
		Refund(context.Background(), "org_1", "purchase_1"); err != nil {
		t.Fatal(err)
	}
}

func TestAnUncomposedServiceRefuses(t *testing.T) {
	if _, err := (Service{}).List(context.Background(), 10); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if _, err := (Service{}).Draw(context.Background(), "org_1", "p_1", 1); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
