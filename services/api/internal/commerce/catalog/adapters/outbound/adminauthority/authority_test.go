package adminauthority

import (
	"context"
	"errors"
	"testing"
	"time"

	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
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

// TestCommercialRolesMayCurate pins who decides what members can be charged
// for: the two roles that already carry commercial responsibility, not a new
// one invented at the adapter.
func TestCommercialRolesMayCurate(t *testing.T) {
	for _, role := range []admindomain.Role{admindomain.RoleFinance, admindomain.RoleAdmin} {
		authority := New(stubAdmin{principal: principalWith(t, role)})
		if err := authority.RequireCatalogEditor(context.Background(), "sess_1"); err != nil {
			t.Errorf("%s was refused: %v", role, err)
		}
	}
}

func TestOtherRolesMayNotCurate(t *testing.T) {
	for _, role := range []admindomain.Role{admindomain.RoleVerifier, admindomain.RoleTSAgent, admindomain.RoleHost} {
		authority := New(stubAdmin{principal: principalWith(t, role)})
		if err := authority.RequireCatalogEditor(context.Background(), "sess_1"); !errors.Is(err, ErrNotCatalogEditor) {
			t.Errorf("%s was allowed to edit the catalog: %v", role, err)
		}
	}
}

// TestUnreadableSessionIsRefusedIdentically stops the check being used to
// probe which admin sessions are live.
func TestUnreadableSessionIsRefusedIdentically(t *testing.T) {
	authority := New(stubAdmin{err: errors.New("session not found")})
	if err := authority.RequireCatalogEditor(context.Background(), "sess_gone"); !errors.Is(err, ErrNotCatalogEditor) {
		t.Fatalf("error = %v, want ErrNotCatalogEditor", err)
	}
}
