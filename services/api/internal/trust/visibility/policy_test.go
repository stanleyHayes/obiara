package visibility

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	trustapplication "github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

func TestDisclosurePolicyFailsClosedOnWithdrawalAndHiddenEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	edges := NewMockEdgeReader(ctrl)
	consent := NewMockConsentEvaluator(ctrl)
	endpoints := NewMockEndpointAuthorizer(ctrl)
	edge := visibleEdge(t, "edge-ab", "a", "b", nil)
	step := stepFrom(edge)
	policy := NewDisclosurePolicy(edges, consent, endpoints, func() time.Time {
		return visibilityTime().Add(time.Hour)
	})

	edges.EXPECT().Find(gomock.Any(), "edge-ab").Return(edge, nil)
	consent.EXPECT().Allows(gomock.Any(), "requester", edge.ConsentRef()).Return(false, nil)
	if policy.Allows(context.Background(), "requester", step) {
		t.Fatal("withdrawn edge was disclosed")
	}

	edges.EXPECT().Find(gomock.Any(), "edge-ab").Return(edge, nil)
	consent.EXPECT().Allows(gomock.Any(), "requester", edge.ConsentRef()).Return(true, nil)
	endpoints.EXPECT().CanReveal(gomock.Any(), "requester", "a").Return(true, nil)
	endpoints.EXPECT().CanReveal(gomock.Any(), "requester", "b").Return(false, nil)
	if policy.Allows(context.Background(), "requester", step) {
		t.Fatal("path with hidden endpoint was disclosed")
	}
}

func TestDisclosurePolicyRejectsExpiredAndMismatchedProjectionWithoutDetail(t *testing.T) {
	ctrl := gomock.NewController(t)
	edges := NewMockEdgeReader(ctrl)
	consent := NewMockConsentEvaluator(ctrl)
	endpoints := NewMockEndpointAuthorizer(ctrl)
	expiry := visibilityTime().Add(30 * time.Minute)
	expired := visibleEdge(t, "edge-expired", "a", "b", &expiry)
	policy := NewDisclosurePolicy(edges, consent, endpoints, func() time.Time {
		return visibilityTime().Add(time.Hour)
	})

	edges.EXPECT().Find(gomock.Any(), "edge-expired").Return(expired, nil)
	if policy.Allows(context.Background(), "requester", stepFrom(expired)) {
		t.Fatal("expired edge was disclosed")
	}

	active := visibleEdge(t, "edge-active", "a", "b", nil)
	mismatch := stepFrom(active)
	mismatch.TargetID = "hidden-target"
	edges.EXPECT().Find(gomock.Any(), "edge-active").Return(active, nil)
	if policy.Allows(context.Background(), "requester", mismatch) {
		t.Fatal("stale or forged projected step was disclosed")
	}
}

func stepFrom(edge domain.Edge) trustapplication.Step {
	return trustapplication.Step{
		EdgeID: edge.ID(), SourceID: edge.SourceID(), TargetID: edge.TargetID(),
		Type: edge.Type(), ProvenanceRef: edge.ProvenanceRef(),
	}
}

func visibleEdge(t *testing.T, id, source, target string, expiresAt *time.Time) domain.Edge {
	t.Helper()
	edge, err := domain.Create(domain.Params{
		ID: id, SourceID: source, TargetID: target, Type: domain.EdgeKnown,
		ProvenanceRef: "provenance-" + id, ConsentRef: "consent-" + id,
		Visibility: domain.VisibilityConsentedPath, CreatedAt: visibilityTime(),
		ExpiresAt: expiresAt,
	}, domain.Command{
		ID: "create-" + id, ActorRef: "actor-1", Kind: "edge.create",
		Payload: id, RecordedAt: visibilityTime(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return edge
}

func visibilityTime() time.Time {
	return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
}
