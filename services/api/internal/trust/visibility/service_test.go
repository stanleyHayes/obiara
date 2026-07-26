package visibility

import (
	"context"
	"math/rand"
	"testing"
	"testing/quick"
	"time"

	trustapplication "github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

func TestExplainReturnsOnlyFullyRevalidatedPaths(t *testing.T) {
	first := visibleEdge(t, "edge-ab", "a", "b", nil)
	second := visibleEdge(t, "edge-bc", "b", "c", nil)
	hidden := visibleEdge(t, "edge-cd", "c", "hidden-d", nil)
	projector := staticProjector{projection: trustapplication.Projection{
		RootID: "a",
		Paths: []trustapplication.Path{
			{Steps: []trustapplication.Step{stepFrom(first), stepFrom(second)}},
			{Steps: []trustapplication.Step{stepFrom(first), stepFrom(second), stepFrom(hidden)}},
		},
	}}
	edges := mapEdgeReader{first.ID(): first, second.ID(): second, hidden.ID(): hidden}
	policy := NewDisclosurePolicy(edges, allowAllConsent{}, endpointSet{
		"a": true, "b": true, "c": true, "hidden-d": false,
	}, func() time.Time { return visibilityTime().Add(time.Hour) })
	service := NewService(projector, policy)

	explanations, err := service.Explain(context.Background(), Request{
		RequesterID: "requester", RootID: "a", MaxDepth: 4, MaxNodes: 10,
	})
	if err != nil || len(explanations) != 1 {
		t.Fatalf("explanations = %#v, error = %v", explanations, err)
	}
	if explanations[0].TargetID != "c" || explanations[0].Hops != 2 {
		t.Fatalf("unexpected explanation: %#v", explanations[0])
	}
	for _, step := range explanations[0].Steps {
		if step.SourceID == "hidden-d" || step.TargetID == "hidden-d" {
			t.Fatal("hidden endpoint leaked")
		}
	}
}

func TestExplanationOrderingIsDeterministicForRandomProjectionOrder(t *testing.T) {
	property := func(seed uint64) bool {
		edgeAB, _ := makeVisibleEdge("edge-ab", "a", "b")
		edgeAC, _ := makeVisibleEdge("edge-ac", "a", "c")
		edgeAD, _ := makeVisibleEdge("edge-ad", "a", "d")
		edges := []domain.Edge{edgeAB, edgeAC, edgeAD}
		random := rand.New(rand.NewSource(int64(seed)))
		random.Shuffle(len(edges), func(i, j int) { edges[i], edges[j] = edges[j], edges[i] })
		paths := make([]trustapplication.Path, 0, len(edges))
		reader := mapEdgeReader{}
		for _, edge := range edges {
			paths = append(paths, trustapplication.Path{Steps: []trustapplication.Step{stepFrom(edge)}})
			reader[edge.ID()] = edge
		}
		service := NewService(
			staticProjector{projection: trustapplication.Projection{RootID: "a", Paths: paths}},
			NewDisclosurePolicy(reader, allowAllConsent{}, endpointSet{
				"a": true, "b": true, "c": true, "d": true,
			}, func() time.Time { return visibilityTime().Add(time.Hour) }),
		)
		result, err := service.Explain(context.Background(), Request{
			RequesterID: "requester", RootID: "a", MaxDepth: 2, MaxNodes: 10,
		})
		return err == nil && len(result) == 3 &&
			result[0].TargetID == "b" && result[1].TargetID == "c" && result[2].TargetID == "d"
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatal(err)
	}
}

type staticProjector struct {
	projection trustapplication.Projection
	err        error
}

func (projector staticProjector) Project(context.Context, trustapplication.ProjectionRequest) (trustapplication.Projection, error) {
	return projector.projection, projector.err
}

type mapEdgeReader map[string]domain.Edge

func (reader mapEdgeReader) Find(_ context.Context, id string) (domain.Edge, error) {
	edge, ok := reader[id]
	if !ok {
		return domain.Edge{}, trustapplication.ErrNotFound
	}
	return edge, nil
}

type allowAllConsent struct{}

func (allowAllConsent) Allows(context.Context, string, string) (bool, error) {
	return true, nil
}

type endpointSet map[string]bool

func (set endpointSet) CanReveal(_ context.Context, _ string, endpoint string) (bool, error) {
	return set[endpoint], nil
}

func makeVisibleEdge(id, source, target string) (domain.Edge, error) {
	return domain.Create(domain.Params{
		ID: id, SourceID: source, TargetID: target, Type: domain.EdgeKnown,
		ProvenanceRef: "provenance-" + id, ConsentRef: "consent-" + id,
		Visibility: domain.VisibilityConsentedPath, CreatedAt: visibilityTime(),
	}, domain.Command{
		ID: "create-" + id, ActorRef: "actor-1", Kind: "edge.create",
		Payload: id, RecordedAt: visibilityTime(),
	})
}
