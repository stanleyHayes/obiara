package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/domain"
	"go.uber.org/mock/gomock"
)

var (
	opened = time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	during = time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC)
	closed = time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
)

type keyerStub struct{ err error }

func (k keyerStub) MemberKey(string) (string, error) {
	if k.err != nil {
		return "", k.err
	}
	return strings.Repeat("a", 64), nil
}

type fixedID string

func (i fixedID) NewID() string { return string(i) }

func promotionFor(t *testing.T, shape domain.Shape, amount int64, redemptionCap uint32) domain.Promotion {
	t.Helper()
	promotion, err := domain.Issue(
		"promo_1", "ASHESI26", "org_1", "sku_membership", shape, amount,
		opened, closed, redemptionCap, domain.Command{ID: "cmd_1", At: opened})
	if err != nil {
		t.Fatal(err)
	}
	return promotion
}

func service(t *testing.T, ctrl *gomock.Controller, repository *MockRepository, issuing bool) Service {
	t.Helper()
	issuers := NewMockIssuers(ctrl)
	issuers.EXPECT().Issuing(gomock.Any(), gomock.Any()).Return(issuing, nil).AnyTimes()
	return New(repository, issuers, keyerStub{}, fixedID("promo_1"), func() time.Time { return during })
}

func TestNoCodeIsIssuedInASuspendedOrganizationsName(t *testing.T) {
	// A suspended organization is not a live relationship, so nothing new is
	// minted in its name. Codes already issued are untouched — that is the
	// promotion's business, not the organization's.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	// No Create expectation: nothing may be written.

	_, err := service(t, ctrl, repository, false).Issue(context.Background(), IssueCommand{
		CommandID: "cmd_1", Code: "ashesi26", IssuerID: "org_1", SKUID: "sku_membership",
		Shape: domain.ShapePercentage, Amount: 50, StartsAt: opened, EndsAt: closed, Cap: 50,
	})
	if !errors.Is(err, ErrIssuerNotIssuing) {
		t.Fatalf("err = %v, want ErrIssuerNotIssuing", err)
	}
}

func TestAMistypedCodeStillBuysTheThing(t *testing.T) {
	// A member who mistyped a code came to buy a membership. Refusing the
	// purchase would punish them for a typo, so it goes ahead at full price.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByCode(gomock.Any(), "WRONG").Return(domain.Promotion{}, ErrNotFound)

	applied, err := service(t, ctrl, repository, true).
		Apply(context.Background(), "wrong", "member-1", "sku_membership", 5000, "cmd_2")
	if err != nil {
		t.Fatalf("a mistyped code refused the purchase: %v", err)
	}
	if applied.DiscountMinor != 0 {
		t.Fatalf("discount = %d", applied.DiscountMinor)
	}
}

func TestNoCodeAtAllIsTheOrdinaryCase(t *testing.T) {
	// Most purchases carry no code. That is not a failure and must not reach
	// the store at all.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	// No expectations: nothing is looked up.

	applied, err := service(t, ctrl, repository, true).
		Apply(context.Background(), "", "member-1", "sku_membership", 5000, "cmd_2")
	if err != nil || applied.DiscountMinor != 0 {
		t.Fatalf("applied = %#v, err = %v", applied, err)
	}
}

func TestACodeForAnotherThingIsNotADiscountOnThisOne(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByCode(gomock.Any(), "ASHESI26").
		Return(promotionFor(t, domain.ShapePercentage, 50, 50), nil)

	applied, err := service(t, ctrl, repository, true).
		Apply(context.Background(), "ASHESI26", "member-1", "sku_something_else", 5000, "cmd_2")
	if err != nil || applied.DiscountMinor != 0 {
		t.Fatalf("a code for another SKU discounted this one: %#v", applied)
	}
}

func TestApplyingACodeSpendsARedemptionAndReportsTheDiscount(t *testing.T) {
	// Redeemed here rather than at settlement, so a discounted price is never
	// charged without a redemption recorded against it.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByCode(gomock.Any(), "ASHESI26").
		Return(promotionFor(t, domain.ShapePercentage, 50, 50), nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "cmd_2").DoAndReturn(
		func(_ context.Context, promotion domain.Promotion, _ uint64, _ string) error {
			if promotion.Redeemed() != 1 {
				t.Fatalf("redeemed = %d", promotion.Redeemed())
			}
			// Keyed, so the row says how many and never who.
			for _, key := range promotion.RedeemedKeys() {
				if len(key) != 64 {
					t.Fatalf("a redemption was recorded unkeyed: %q", key)
				}
			}
			return nil
		})

	applied, err := service(t, ctrl, repository, true).
		Apply(context.Background(), "ashesi26", "member-1", "sku_membership", 5000, "cmd_2")
	if err != nil {
		t.Fatal(err)
	}
	if applied.DiscountMinor != 2500 || applied.Code != "ASHESI26" {
		t.Fatalf("applied = %#v", applied)
	}
}

func TestLosingTheRaceForTheLastSlotChargesFullPrice(t *testing.T) {
	// Two members read the same revision and only one write lands. The loser
	// pays full price rather than a discount nothing recorded — which would
	// be money the books cannot account for.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByCode(gomock.Any(), "ASHESI26").
		Return(promotionFor(t, domain.ShapeFixed, 1000, 1), nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "cmd_2").Return(ErrConflict)

	applied, err := service(t, ctrl, repository, true).
		Apply(context.Background(), "ASHESI26", "member-1", "sku_membership", 5000, "cmd_2")
	if err != nil {
		t.Fatalf("a lost race refused the purchase: %v", err)
	}
	if applied.DiscountMinor != 0 {
		t.Fatalf("a discount was given that nothing recorded: %d", applied.DiscountMinor)
	}
}

func TestAnUnkeyableMemberGetsNoDiscountAndNoPurchase(t *testing.T) {
	// If the member cannot be keyed, the redemption cannot be recorded
	// against anybody. That is a fault rather than a quiet full-price sale,
	// because it means the keyer is down and the next thing to fail is worse.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByCode(gomock.Any(), "ASHESI26").
		Return(promotionFor(t, domain.ShapeFixed, 1000, 5), nil)
	issuers := NewMockIssuers(ctrl)
	issuers.EXPECT().Issuing(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	broken := New(repository, issuers, keyerStub{err: errors.New("hmac down")},
		fixedID("promo_1"), func() time.Time { return during })
	if _, err := broken.Apply(
		context.Background(), "ASHESI26", "member-1", "sku_membership", 5000, "cmd_2",
	); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestAnUncomposedServiceRefuses(t *testing.T) {
	if _, err := (Service{}).Issue(context.Background(), IssueCommand{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if _, err := (Service{}).Apply(
		context.Background(), "CODE", "member-1", "sku", 100, "cmd",
	); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
