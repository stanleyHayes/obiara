package application

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"testing/quick"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

func TestProjectionIsAuthorizedBoundedAndCycleSafe(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockProjectionAuthorizer(ctrl)
	consent := NewMockConsentEvaluator(ctrl)
	now := testTime().Add(time.Hour)
	service := NewService(repository, authorizer, consent, func() time.Time { return now })
	aToB := applicationEdge(t, "edge-ab", "a", "b", domain.VisibilityConsentedPath)
	bToC := applicationEdge(t, "edge-bc", "b", "c", domain.VisibilityConsentedPath)
	cToA := applicationEdge(t, "edge-ca", "c", "a", domain.VisibilityConsentedPath)
	cToD := applicationEdge(t, "edge-cd", "c", "d", domain.VisibilityConsentedPath)

	authorizer.EXPECT().CanProject(gomock.Any(), "owner", "a").Return(true, nil)
	gomock.InOrder(
		repository.EXPECT().Outgoing(gomock.Any(), []string{"a"}).Return([]domain.Edge{aToB}, nil),
		repository.EXPECT().Outgoing(gomock.Any(), []string{"b"}).Return([]domain.Edge{bToC}, nil),
		repository.EXPECT().Outgoing(gomock.Any(), []string{"c"}).Return([]domain.Edge{cToA, cToD}, nil),
		repository.EXPECT().Outgoing(gomock.Any(), []string{"d"}).Return(nil, nil),
	)
	consent.EXPECT().Allows(gomock.Any(), "owner", gomock.Any()).Return(true, nil).Times(4)
	projection, err := service.Project(context.Background(), ProjectionRequest{
		OwnerID: "owner", RootID: "a", MaxDepth: 4, MaxNodes: 10,
	})
	if err != nil || len(projection.Paths) != 3 {
		t.Fatalf("projection = %#v, error = %v", projection, err)
	}
	for _, path := range projection.Paths {
		if len(path.Steps) > 4 {
			t.Fatalf("path exceeded depth: %#v", path)
		}
		for _, step := range path.Steps {
			if step.TargetID == "a" {
				t.Fatalf("cycle returned to root: %#v", path)
			}
		}
	}
}

func TestProjectionFailsClosedBeforeGraphRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockProjectionAuthorizer(ctrl)
	consent := NewMockConsentEvaluator(ctrl)
	service := NewService(repository, authorizer, consent, time.Now)
	authorizer.EXPECT().CanProject(gomock.Any(), "owner", "root").Return(false, nil)
	if _, err := service.Project(context.Background(), ProjectionRequest{
		OwnerID: "owner", RootID: "root", MaxDepth: 2, MaxNodes: 10,
	}); !errors.Is(err, domain.ErrAccessDenied) {
		t.Fatalf("error = %v, want %v", err, domain.ErrAccessDenied)
	}
}

func TestProjectionRequiresConsentEvenForOwnerVisibleEdge(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockProjectionAuthorizer(ctrl)
	consent := NewMockConsentEvaluator(ctrl)
	service := NewService(repository, authorizer, consent, func() time.Time {
		return testTime().Add(time.Hour)
	})
	edge := applicationEdge(t, "edge-owner", "owner", "target", domain.VisibilityOwnerOnly)
	authorizer.EXPECT().CanProject(gomock.Any(), "owner", "owner").Return(true, nil)
	repository.EXPECT().Outgoing(gomock.Any(), []string{"owner"}).Return([]domain.Edge{edge}, nil)
	consent.EXPECT().Allows(gomock.Any(), "owner", edge.ConsentRef()).Return(false, nil)
	projection, err := service.Project(context.Background(), ProjectionRequest{
		OwnerID: "owner", RootID: "owner", MaxDepth: 2, MaxNodes: 10,
	})
	if err != nil || len(projection.Paths) != 0 {
		t.Fatalf("projection = %#v, error = %v", projection, err)
	}
}

func TestRandomCyclicGraphsAlwaysRespectProjectionBounds(t *testing.T) {
	property := func(seed uint64, requestedDepth uint8, requestedNodes uint8) bool {
		depth := int(requestedDepth%domain.MaxProjectionDepth) + 1
		nodes := int(requestedNodes%(domain.MaxProjectionNodes-1)) + 2
		random := rand.New(rand.NewSource(int64(seed)))
		edges := make(map[string][]domain.Edge, nodes)
		for index := 0; index < nodes*3; index++ {
			source := fmt.Sprintf("node-%d", random.Intn(nodes))
			target := fmt.Sprintf("node-%d", random.Intn(nodes))
			if source == target {
				continue
			}
			edge, err := makeEdge(
				fmt.Sprintf("edge-%d", index), source, target, domain.VisibilityConsentedPath,
			)
			if err == nil {
				edges[source] = append(edges[source], edge)
			}
		}
		repository := &propertyRepository{edges: edges}
		service := NewService(repository, allowProjection{}, allowConsent{}, func() time.Time {
			return testTime().Add(time.Hour)
		})
		projection, err := service.Project(context.Background(), ProjectionRequest{
			OwnerID: "owner", RootID: "node-0", MaxDepth: depth, MaxNodes: nodes,
		})
		if err != nil {
			return false
		}
		targets := map[string]struct{}{}
		for _, path := range projection.Paths {
			if len(path.Steps) > depth {
				return false
			}
			target := path.Steps[len(path.Steps)-1].TargetID
			if _, duplicate := targets[target]; duplicate || target == "node-0" {
				return false
			}
			targets[target] = struct{}{}
		}
		return len(targets) < nodes
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatal(err)
	}
}

type propertyRepository struct {
	edges map[string][]domain.Edge
}

func (repository *propertyRepository) Find(context.Context, string) (domain.Edge, error) {
	return domain.Edge{}, ErrNotFound
}
func (repository *propertyRepository) Save(context.Context, domain.Edge, uint64, string) error {
	return nil
}
func (repository *propertyRepository) Outgoing(_ context.Context, sources []string) ([]domain.Edge, error) {
	var result []domain.Edge
	for _, source := range sources {
		result = append(result, repository.edges[source]...)
	}
	return result, nil
}

type allowProjection struct{}

func (allowProjection) CanProject(context.Context, string, string) (bool, error) { return true, nil }

type allowConsent struct{}

func (allowConsent) Allows(context.Context, string, string) (bool, error) { return true, nil }

func applicationEdge(t *testing.T, id, source, target string, visibility domain.Visibility) domain.Edge {
	t.Helper()
	edge, err := makeEdge(id, source, target, visibility)
	if err != nil {
		t.Fatal(err)
	}
	return edge
}

func makeEdge(id, source, target string, visibility domain.Visibility) (domain.Edge, error) {
	return domain.Create(domain.Params{
		ID: id, SourceID: source, TargetID: target, Type: domain.EdgeKnown,
		ProvenanceRef: "provenance-" + id, ConsentRef: "consent-" + id,
		Visibility: visibility, CreatedAt: testTime(),
	}, domain.Command{
		ID: "create-" + id, ActorRef: "actor-1", Kind: "edge.create",
		Payload: id, RecordedAt: testTime(),
	})
}

func testTime() time.Time {
	return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
}
