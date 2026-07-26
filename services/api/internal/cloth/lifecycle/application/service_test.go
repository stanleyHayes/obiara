package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/lifecycle/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestDeleteFailClosedOnLegalHold(t *testing.T) {
	c := gomock.NewController(t)
	r := NewMockRepository(c)
	k := NewMockKeyer(c)
	h := NewMockLegalHold(c)
	v, _ := domain.New(key(9), []string{key(1), key(2)}, domain.Provenance{BandVersion: 1, RecipeRef: "ref_recipeabcdefghijklmnop"})
	k.EXPECT().Key("cloth_pair", "pair").Return(key(9), nil).Times(2)
	k.EXPECT().Key("cloth_member", "member").Return(key(1), nil).Times(2)
	r.EXPECT().Find(gomock.Any(), key(9)).Return(v, nil).Times(2)
	h.EXPECT().DeletionAllowed(gomock.Any(), key(9)).Return(false, nil)
	svc := NewService(r, k, h, func() time.Time { return time.Unix(100, 0) })
	if _, e := svc.Delete(context.Background(), "pair", "member", "delete-command-01", 0); e != domain.ErrDenied {
		t.Fatalf("%v", e)
	}
	h.EXPECT().DeletionAllowed(gomock.Any(), key(9)).Return(true, nil)
	k.EXPECT().Key("cloth_receipt", gomock.Any()).Return(key(8), nil)
	r.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "delete-command-01").Return(nil)
	got, e := svc.Delete(context.Background(), "pair", "member", "delete-command-01", 0)
	if e != nil || got.Status() != domain.StatusDeleted {
		t.Fatalf("%v %v", got.Status(), e)
	}
}
