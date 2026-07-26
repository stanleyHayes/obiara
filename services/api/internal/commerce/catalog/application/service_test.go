package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type keyer struct{}

func (keyer) Key(string, string) (string, error) {
	return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
}

type ids struct{}

func (ids) NewID() string { return "sku:1" }
func TestCreateHasNoCheckoutProviderOrProfileDependency(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	authority := NewMockAuthority(ctrl)
	now := time.Date(2026, 7, 26, 21, 0, 0, 0, time.UTC)
	p, _ := domain.NewPrice(domain.CurrencyGHS, 5000)
	authority.EXPECT().RequireCatalogEditor(gomock.Any(), "editor").Return(nil)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, s domain.SKU) error {
		if s.Price() != p || s.Kind() != domain.KindPhysicalGood {
			t.Fatal("sku mismatch")
		}
		return nil
	})
	service := NewService(repo, authority, keyer{}, ids{}, func() time.Time { return now })
	_, e := service.Create(context.Background(), CreateCommand{Actor: "editor", SKUKey: "cloth.item", Title: "woven cloth", CommandID: "create", Kind: domain.KindPhysicalGood, Price: p})
	if e != nil {
		t.Fatal(e)
	}
}
