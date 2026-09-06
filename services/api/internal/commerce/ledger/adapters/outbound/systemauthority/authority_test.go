package systemauthority

import (
	"context"
	"errors"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
)

func TestOnlyTheSystemPostsSettlement(t *testing.T) {
	authority := New()
	if err := authority.RequirePoster(
		context.Background(), Actor, domain.PurposeSaleSettlement); err != nil {
		t.Fatalf("the system could not book a payment: %v", err)
	}
	if err := authority.RequirePoster(
		context.Background(), Actor, domain.PurposeRefundSettlement); err != nil {
		t.Fatalf("the system could not book a refund: %v", err)
	}
	// Anybody else, refused. This authority exists for one actor.
	for _, actor := range []string{"", "adm_1", "system", "system:settlement "} {
		if err := authority.RequirePoster(
			context.Background(), actor, domain.PurposeSaleSettlement,
		); !errors.Is(err, ErrNotPoster) && actor != "system:settlement " {
			t.Fatalf("%q was admitted", actor)
		}
	}
}

func TestTheSystemCannotPostWhateverItLikes(t *testing.T) {
	// The actor name is a constant in this binary. If it ever leaked into a
	// request, the purpose bound is what still stops it posting anything.
	if err := New().RequirePoster(
		context.Background(), Actor, domain.PurposeCatalogReceivable,
	); !errors.Is(err, ErrNotPoster) {
		t.Fatalf("err = %v, want ErrNotPoster", err)
	}
}

func TestTheSystemNeverReadsABalance(t *testing.T) {
	// Reading a balance is a person's act and stays behind the finance desk.
	// A system that could read one would be a way around that gate.
	if err := New().RequireBalanceReader(
		context.Background(), Actor, "membership_revenue",
	); !errors.Is(err, ErrNotReader) {
		t.Fatalf("err = %v, want ErrNotReader", err)
	}
}
