package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type keyer struct{}

func (keyer) Key(ns, v string) (string, error) {
	if v == "asset" {
		return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
	}
	if v == "revenue" {
		return "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	return "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", nil
}

type ids struct{}

func (ids) NewID() string { return "posting:1" }
func TestPostUsesCurrentAuthorityAndNoProviderCatalogOrAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	authority := NewMockAuthority(ctrl)
	now := time.Date(2026, 7, 26, 22, 0, 0, 0, time.UTC)
	authority.EXPECT().RequirePoster(gomock.Any(), "poster", domain.PurposeSaleSettlement).Return(nil)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, p domain.Posting) error {
		if len(p.Lines()) != 2 {
			t.Fatal("single entry")
		}
		return nil
	})
	s := NewService(repo, authority, keyer{}, ids{}, func() time.Time { return now })
	_, e := s.Post(context.Background(), PostCommand{Actor: "poster", CommandID: "command:1", ReferenceID: "sale:1", Purpose: domain.PurposeSaleSettlement, Currency: domain.CurrencyGHS, Lines: []Line{{"asset", domain.ClassAsset, domain.SideDebit, 100}, {"revenue", domain.ClassRevenue, domain.SideCredit, 100}}})
	if e != nil {
		t.Fatal(e)
	}
}
