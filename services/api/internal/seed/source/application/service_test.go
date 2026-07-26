package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "request-1" }
func digest(n int) string      { return fmt.Sprintf("%064x", n) }

func harness(t *testing.T) (*MockRepository, *MockAuthorizer, *MockSourcePolicy, *MockCandidateResolver, *MockConsentVisibility, *MockKeyer, Service) {
	t.Helper()
	ctrl := gomock.NewController(t)
	r, a, p := NewMockRepository(ctrl), NewMockAuthorizer(ctrl), NewMockSourcePolicy(ctrl)
	c, v, k := NewMockCandidateResolver(ctrl), NewMockConsentVisibility(ctrl), NewMockKeyer(ctrl)
	now := func() time.Time { return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC) }
	return r, a, p, c, v, k, NewService(r, a, p, c, v, k, fixedIDs{}, now)
}
func proposal() Proposal {
	return Proposal{RequesterID: "member-a", SourceType: domain.SourceCircle, SourceRef: "circle-a", TTL: time.Hour}
}
func openCommand() Command {
	return Command{ID: "open-1", ActorID: "member-a", ReasonCode: "user_requested"}
}

func TestOpenUsesPrivacyNeutralAuthorizationAndPolicyDenial(t *testing.T) {
	for _, tc := range []struct {
		name      string
		authorize bool
	}{
		{"authorization", false}, {"policy", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, a, p, _, _, _, service := harness(t)
			if tc.authorize {
				a.EXPECT().Require(gomock.Any(), "member-a", "seed.source.open", domain.SourceCircle, "circle-a")
				p.EXPECT().Allow(gomock.Any(), domain.SourceCircle, "circle-a").Return(errors.New("denied"))
			} else {
				a.EXPECT().Require(gomock.Any(), "member-a", "seed.source.open", domain.SourceCircle, "circle-a").Return(errors.New("denied"))
			}
			if _, err := service.Open(context.Background(), openCommand(), proposal()); !errors.Is(err, ErrNotAvailable) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestOpenFiltersConsentAndStoresOnlyOpaqueBoundedIDs(t *testing.T) {
	r, a, p, c, v, k, service := harness(t)
	a.EXPECT().Require(gomock.Any(), "member-a", "seed.source.open", domain.SourceCircle, "circle-a")
	p.EXPECT().Allow(gomock.Any(), domain.SourceCircle, "circle-a")
	c.EXPECT().CandidateIDs(gomock.Any(), domain.SourceCircle, "circle-a", domain.MaxCandidates).Return([]string{"candidate-b", "candidate-a"}, nil)
	v.EXPECT().Visible(gomock.Any(), "member-a", "candidate-b", domain.SourceCircle, "circle-a").Return(false, nil)
	v.EXPECT().Visible(gomock.Any(), "member-a", "candidate-a", domain.SourceCircle, "circle-a").Return(true, nil)
	k.EXPECT().Key("seed-source:candidate", "candidate-a").Return(digest(3), nil)
	k.EXPECT().Key("seed-source:requester", "member-a").Return(digest(1), nil)
	k.EXPECT().Key("seed-source:source", "circle-a").Return(digest(2), nil)
	r.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, request domain.Request) error {
		if !slices.Equal(request.CandidateIDs(), []string{digest(3)}) {
			t.Fatalf("candidate IDs=%v", request.CandidateIDs())
		}
		if request.RequesterKey() == "member-a" || request.Source().Key == "circle-a" {
			t.Fatal("raw identifier crossed persistence boundary")
		}
		return nil
	})
	result, err := service.Open(context.Background(), openCommand(), proposal())
	if err != nil || len(result.Request.CandidateIDs()) != 1 {
		t.Fatalf("result=%v err=%v", result.Request.CandidateIDs(), err)
	}
}

func TestOpenRejectsResolverThatViolatesBound(t *testing.T) {
	_, a, p, c, _, _, service := harness(t)
	a.EXPECT().Require(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	p.EXPECT().Allow(gomock.Any(), gomock.Any(), gomock.Any())
	candidates := make([]string, domain.MaxCandidates+1)
	c.EXPECT().CandidateIDs(gomock.Any(), gomock.Any(), gomock.Any(), domain.MaxCandidates).Return(candidates, nil)
	if _, err := service.Open(context.Background(), openCommand(), proposal()); !errors.Is(err, ErrNotAvailable) {
		t.Fatalf("err=%v", err)
	}
}
