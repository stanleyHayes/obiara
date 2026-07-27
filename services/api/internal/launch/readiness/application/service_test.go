package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/launch/readiness/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }

type ids struct{}

func (ids) NewID() string { return key(1) }

type clock struct{ at time.Time }

func (c clock) Now() time.Time { return c.at }
func TestProjectUsesCurrentAggregateEvidence(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	now := time.Now().UTC()
	a := NewMockAuthority(ctrl)
	f := NewMockFamilyProjection(ctrl)
	h := NewMockHostProjection(ctrl)
	l := NewMockLicenseProjection(ctrl)
	r := NewMockRepository(ctrl)
	families := domain.FamilyDensity{Market: "GH", ConsentedFamilies: 100, TargetFamilies: 100, DenseCircles: 5, RequiredDenseCircles: 5, EvidenceVersion: 1, Complete: true, ObservedAt: now}
	hosts := domain.HostCoverage{Market: "GH", Trained: 10, Certified: 10, Required: 10, TrainingVersion: 2, CertificationVersion: 3, CertifiedUntil: now.Add(time.Hour), Complete: true, ObservedAt: now}
	licenses := domain.LicenseCoverage{Market: "GH", Jurisdiction: "gh-accra", Licensed: 4, Required: 4, LicenseVersion: 3, LicensedUntil: now.Add(time.Hour), Complete: true, ObservedAt: now}
	a.EXPECT().RequireLaunchReviewer(ctx, key(2)).Return(nil)
	f.EXPECT().CurrentFamilyDensity(ctx, "GH").Return(families, nil)
	h.EXPECT().CurrentHostCoverage(ctx, "GH").Return(hosts, nil)
	l.EXPECT().CurrentLicenseCoverage(ctx, "GH", "gh-accra").Return(licenses, nil)
	r.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s domain.Snapshot) error {
		if !s.Ready() {
			t.Fatal("expected ready")
		}
		return nil
	})
	s := New(a, f, h, l, r, ids{}, clock{now})
	if _, e := s.Project(ctx, key(2), "GH", "gh-accra", "review-1"); e != nil {
		t.Fatal(e)
	}
}
func TestAuthorityFailureDoesNotReadSources(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	a := NewMockAuthority(ctrl)
	a.EXPECT().RequireLaunchReviewer(ctx, key(2)).Return(fmt.Errorf("denied"))
	s := New(a, NewMockFamilyProjection(ctrl), NewMockHostProjection(ctrl), NewMockLicenseProjection(ctrl), NewMockRepository(ctrl), ids{}, clock{time.Now()})
	if _, e := s.Project(ctx, key(2), "GH", "gh-accra", "review-1"); e != ErrUnavailable {
		t.Fatalf("got %v", e)
	}
}
