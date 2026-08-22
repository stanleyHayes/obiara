package adminauthority

import (
	"context"
	"errors"
	"testing"
	"time"

	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
)

type stubAdmin struct {
	principal admindomain.Principal
	err       error
}

func (stub stubAdmin) Authenticate(context.Context, string) (admindomain.Session, admindomain.Principal, error) {
	return admindomain.Session{}, stub.principal, stub.err
}

func principalWith(t *testing.T, roles ...admindomain.Role) admindomain.Principal {
	t.Helper()
	principal, err := admindomain.NewPrincipal("adm_1", "op@obiara.test", roles, time.Now())
	if err != nil {
		t.Fatalf("NewPrincipal: %v", err)
	}
	return principal
}

// TestFinanceDeskMayPostAndRead pins who touches the record of what the
// platform owes and is owed.
func TestFinanceDeskMayPostAndRead(t *testing.T) {
	for _, role := range []admindomain.Role{admindomain.RoleFinance, admindomain.RoleAdmin} {
		authority := New(stubAdmin{principal: principalWith(t, role)})
		if err := authority.RequirePoster(context.Background(), "sess_1", domain.PurposeSaleSettlement); err != nil {
			t.Errorf("%s could not post: %v", role, err)
		}
		if err := authority.RequireBalanceReader(context.Background(), "sess_1", "acct_1"); err != nil {
			t.Errorf("%s could not read a balance: %v", role, err)
		}
	}
}

// TestOtherDesksAreRefused keeps least privilege real: a verifier has no
// reason to see settlement positions.
func TestOtherDesksAreRefused(t *testing.T) {
	for _, role := range []admindomain.Role{admindomain.RoleVerifier, admindomain.RoleTSAgent, admindomain.RoleHost} {
		authority := New(stubAdmin{principal: principalWith(t, role)})
		if err := authority.RequirePoster(context.Background(), "sess_1", domain.PurposeSaleSettlement); !errors.Is(err, ErrNotPoster) {
			t.Errorf("%s was allowed to post: %v", role, err)
		}
		if err := authority.RequireBalanceReader(context.Background(), "sess_1", "acct_1"); !errors.Is(err, ErrNotBalanceReader) {
			t.Errorf("%s was allowed to read a balance: %v", role, err)
		}
	}
}

func TestUnreadableSessionIsRefused(t *testing.T) {
	authority := New(stubAdmin{err: errors.New("session not found")})
	if err := authority.RequirePoster(context.Background(), "sess_gone", domain.PurposeSaleSettlement); !errors.Is(err, ErrNotPoster) {
		t.Fatalf("error = %v, want ErrNotPoster", err)
	}
}
