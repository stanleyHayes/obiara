package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/communityops/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func k(n int) string { return fmt.Sprintf("%064x", n) }

type fixedID struct{}

func (fixedID) NewID() string { return k(9) }

type fixedClock struct{ at time.Time }

func (c fixedClock) Now() time.Time { return c.at }
func TestProposeAndAcknowledgeRevalidateWithoutMutationPorts(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	now := time.Now().UTC()
	a := NewMockAuthority(ctrl)
	d := NewMockDensitySource(ctrl)
	h := NewMockHostSource(ctrl)
	n := NewMockNoticeCatalog(ctrl)
	r := NewMockRepository(ctrl)
	density := domain.Density{CircleKey: k(1), FireKey: k(2), Participants: 20, Capacity: 50, Version: 3, ObservedAt: now}
	host := domain.HostEligibility{HostKey: k(3), VerificationVersion: 2, CertificationVersion: 4, VerifiedUntil: now.Add(time.Hour), CertifiedUntil: now.Add(time.Hour), Trained: true, SubanVetted: true}
	notice := domain.Notice{TemplateKey: "fire.host_change", Version: 2, Locale: "en-gh", AudienceCount: 20}
	notice.Digest = domain.NoticeDigest(notice.TemplateKey, notice.Version, notice.Locale, notice.AudienceCount)
	req := ProposeRequest{k(8), k(1), k(2), k(3), domain.ReplaceHost, domain.ReasonHostUnavailable, k(7), "fire.host_change", "en-gh", "propose-1"}
	a.EXPECT().RequireCommunityOperator(ctx, k(8)).Return(nil)
	d.EXPECT().CurrentDensity(ctx, k(1), k(2)).Return(density, nil)
	h.EXPECT().CurrentEligibility(ctx, k(3)).Return(host, nil)
	n.EXPECT().CurrentNotice(ctx, "fire.host_change", "en-gh", 20).Return(notice, nil)
	r.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	s := New(a, d, h, n, r, fixedID{}, fixedClock{now})
	p, e := s.Propose(ctx, req)
	if e != nil {
		t.Fatal(e)
	}
	a.EXPECT().RequireCommunityOperator(ctx, k(8)).Return(nil)
	r.EXPECT().Find(ctx, p.ID()).Return(p, nil)
	d.EXPECT().CurrentDensity(ctx, k(1), k(2)).Return(density, nil)
	h.EXPECT().CurrentEligibility(ctx, k(3)).Return(host, nil)
	n.EXPECT().CurrentNotice(ctx, "fire.host_change", "en-gh", 20).Return(notice, nil)
	r.EXPECT().Save(ctx, gomock.Any(), uint64(1), "ack-1").Return(nil)
	next, e := s.AcknowledgeNotice(ctx, p.ID(), k(8), notice.Digest, "ack-1")
	if e != nil || !next.ReadyForHumanReview() {
		t.Fatal(e)
	}
}
func TestAcknowledgementBlocksStaleDensity(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	now := time.Now().UTC()
	a := NewMockAuthority(ctrl)
	d := NewMockDensitySource(ctrl)
	h := NewMockHostSource(ctrl)
	n := NewMockNoticeCatalog(ctrl)
	r := NewMockRepository(ctrl)
	density := domain.Density{CircleKey: k(1), FireKey: k(2), Participants: 20, Capacity: 50, Version: 3, ObservedAt: now}
	notice := domain.Notice{TemplateKey: "fire.cancel", Version: 1, Locale: "en-gh", AudienceCount: 20}
	notice.Digest = domain.NoticeDigest(notice.TemplateKey, 1, notice.Locale, 20)
	p, _ := domain.Propose(k(9), k(8), domain.CancelFire, domain.ReasonScheduleConflict, k(7), density, nil, notice, "propose-1", now)
	a.EXPECT().RequireCommunityOperator(ctx, k(8)).Return(nil)
	r.EXPECT().Find(ctx, p.ID()).Return(p, nil)
	stale := density
	stale.Version++
	d.EXPECT().CurrentDensity(ctx, k(1), k(2)).Return(stale, nil)
	s := New(a, d, h, n, r, fixedID{}, fixedClock{now})
	if _, e := s.AcknowledgeNotice(ctx, p.ID(), k(8), notice.Digest, "ack-1"); e != ErrUnavailable {
		t.Fatalf("got %v", e)
	}
}
