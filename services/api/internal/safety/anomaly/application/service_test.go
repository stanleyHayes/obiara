package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/anomaly/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func setup(t *testing.T) (Service, *MockRuleCatalog, *MockConsentVerifier, *MockAuthority, *MockCaseRouter, domain.Rule, domain.Aggregate) {
	t.Helper()
	ctrl := gomock.NewController(t)
	rules := NewMockRuleCatalog(ctrl)
	consent := NewMockConsentVerifier(ctrl)
	authority := NewMockAuthority(ctrl)
	cases := NewMockCaseRouter(ctrl)
	now := time.Date(2026, 7, 26, 18, 0, 0, 0, time.UTC)
	r, _ := domain.NewRule("reviewed.syndicate", 1, domain.ShapeSyndicate, 4, 4, .2, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", now.Add(-time.Hour))
	a, _ := domain.NewAggregate("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", r.Key, r.Version, r.Shape, 400, 3000, 4, 4, .99, false, false)
	return NewService(rules, consent, authority, cases, func() time.Time { return now }), rules, consent, authority, cases, r, a
}
func TestRoutesOnlyAfterCurrentConsentAndRouteAuthority(t *testing.T) {
	s, rules, consent, authority, cases, r, a := setup(t)
	authority.EXPECT().RequireOfflineEvaluator(gomock.Any(), "actor").Return(nil)
	rules.EXPECT().Current(gomock.Any(), a.Shape).Return(r, nil)
	consent.EXPECT().Revalidate(gomock.Any(), a.EvidenceRef).Return(nil)
	authority.EXPECT().RequireHumanRoute(gomock.Any(), "actor", a.EvidenceRef).Return(nil)
	cases.EXPECT().OpenHumanCase(gomock.Any(), gomock.Any()).Return(nil)
	d, e := s.Evaluate(context.Background(), Request{Actor: "actor", Aggregate: a})
	if e != nil || d.Outcome != domain.OutcomeHumanReview {
		t.Fatalf("decision=%+v err=%v", d, e)
	}
}
func TestStaleConsentFailsClosedWithoutCase(t *testing.T) {
	s, rules, consent, authority, _, r, a := setup(t)
	authority.EXPECT().RequireOfflineEvaluator(gomock.Any(), "actor").Return(nil)
	rules.EXPECT().Current(gomock.Any(), a.Shape).Return(r, nil)
	consent.EXPECT().Revalidate(gomock.Any(), a.EvidenceRef).Return(errors.New("withdrawn"))
	d, e := s.Evaluate(context.Background(), Request{Actor: "actor", Aggregate: a})
	if e != nil || d.Outcome != domain.OutcomeNoAction {
		t.Fatalf("decision=%+v err=%v", d, e)
	}
}
