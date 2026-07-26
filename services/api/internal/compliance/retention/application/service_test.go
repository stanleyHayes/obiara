package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/compliance/retention/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type keyer struct{}

func (keyer) Key(ns, v string) (string, error) {
	if ns == "retention-erasure-proof" {
		return "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", nil
	}
	return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
}

type ids struct{}

func (ids) NewID() string { return "retention:1" }
func TestCompleteRequiresVerifierBeforeMutation(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	policies := NewMockPolicyCatalog(ctrl)
	authority := NewMockAuthority(ctrl)
	verifier := NewMockErasureVerifier(ctrl)
	now := time.Date(2026, 7, 26, 19, 0, 0, 0, time.UTC)
	p, _ := domain.NewPolicy("messages.metadata", "safety.audit", 1, 24*time.Hour, now.Add(-time.Hour))
	r, _ := domain.Create("retention:1", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", p, domain.Command{ID: "create", At: now})
	r, _ = r.RequestErasure(domain.Command{ID: "request", ExpectedRevision: 1, At: now})
	authority.EXPECT().RequireSubject(gomock.Any(), "actor", "member").Return(nil)
	verifier.EXPECT().Verify(gomock.Any(), "member", "retention:1", "complete").Return("verified-external-proof", nil)
	repo.EXPECT().Find(gomock.Any(), "retention:1").Return(r, nil)
	repo.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(2), "complete").Return(nil)
	s := NewService(repo, policies, authority, verifier, keyer{}, ids{}, func() time.Time { return now })
	out, e := s.CompleteErasure(context.Background(), Change{Actor: "actor", Subject: "member", ID: "retention:1", CommandID: "complete", ExpectedRevision: 2})
	if e != nil || out.Status() != domain.StatusErased {
		t.Fatalf("status=%s err=%v", out.Status(), e)
	}
}
