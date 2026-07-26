package application

import (
	"context"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/matching/features/domain"
	"go.uber.org/mock/gomock"
)

type testKeyer struct{}

func (testKeyer) Key(_ string, v string) (string, error) { return "key:" + v, nil }

type testIDs struct{}

func (testIDs) NewID() string { return "decision:one" }

func TestDecideUsesOnlyCurrentPairIntersection(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalog := NewMockCatalog(ctrl)
	grants := NewMockGrantRepository(ctrl)
	decisions := NewMockDecisionRepository(ctrl)
	authority := NewMockAuthority(ctrl)
	now := time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)
	def, _ := domain.NewDefinition("shared.rituals", 1, "matching.compatibility", now.Add(-time.Hour))
	a, _ := domain.GrantFeature("key:alice", def, 11, domain.Command{ID: "grant:a", At: now.Add(-time.Minute)})
	b, _ := domain.GrantFeature("key:bob", def, 12, domain.Command{ID: "grant:b", At: now.Add(-time.Minute)})
	extra, _ := domain.NewDefinition("shared.music", 1, "matching.compatibility", now.Add(-time.Hour))
	onlyA, _ := domain.GrantFeature("key:alice", extra, 13, domain.Command{ID: "grant:extra", At: now.Add(-time.Minute)})

	authority.EXPECT().RequirePair(gomock.Any(), "actor", "alice", "bob").Return(nil)
	grants.EXPECT().ListEffective(gomock.Any(), "key:alice").Return([]domain.Grant{a, onlyA}, nil)
	grants.EXPECT().ListEffective(gomock.Any(), "key:bob").Return([]domain.Grant{b}, nil)
	catalog.EXPECT().Current(gomock.Any(), "shared.rituals").Return(def, nil).Times(2)
	catalog.EXPECT().Current(gomock.Any(), "shared.music").Return(extra, nil)
	decisions.EXPECT().CreateDecision(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, d domain.Decision) error {
		if len(d.Features) != 1 || d.Features[0].Key != "shared.rituals" {
			t.Fatalf("unexpected intersection: %+v", d.Features)
		}
		return nil
	})
	s := NewService(catalog, grants, decisions, authority, testKeyer{}, testIDs{}, func() time.Time { return now })
	d, err := s.Decide(context.Background(), PairRequest{Actor: "actor", First: "alice", Second: "bob"})
	if err != nil || len(d.Features) != 1 {
		t.Fatalf("decision=%+v err=%v", d, err)
	}
}

func TestRevalidateDeniesImmediatelyAfterWithdrawal(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalog := NewMockCatalog(ctrl)
	grants := NewMockGrantRepository(ctrl)
	decisions := NewMockDecisionRepository(ctrl)
	authority := NewMockAuthority(ctrl)
	now := time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)
	def, _ := domain.NewDefinition("shared.rituals", 1, "matching.compatibility", now.Add(-time.Hour))
	d, _ := domain.NewDecision("decision:one", "key:alice", "key:bob", []domain.EnabledFeature{{Key: def.Key, FeatureVersion: 1, Purpose: def.Purpose, Consents: []domain.ConsentRef{{MemberKey: "key:alice", GrantVersion: 1}, {MemberKey: "key:bob", GrantVersion: 1}}}}, now)
	g, _ := domain.GrantFeature("key:alice", def, 1, domain.Command{ID: "grant:a", At: now.Add(-time.Minute)})
	g, _ = g.Withdraw(domain.Command{ID: "withdraw:a", ExpectedRevision: 1, At: now})
	decisions.EXPECT().FindDecision(gomock.Any(), "decision:one").Return(d, nil)
	catalog.EXPECT().Current(gomock.Any(), def.Key).Return(def, nil)
	grants.EXPECT().Find(gomock.Any(), "key:alice", def.Key).Return(g, nil)
	s := NewService(catalog, grants, decisions, authority, testKeyer{}, testIDs{}, func() time.Time { return now })
	ok, err := s.Revalidate(context.Background(), "decision:one")
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
