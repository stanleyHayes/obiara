package visibility

import (
	"context"
	"time"

	trustapplication "github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

type DisclosurePolicy struct {
	edges     EdgeReader
	consent   ConsentEvaluator
	endpoints EndpointAuthorizer
	now       func() time.Time
}

func NewDisclosurePolicy(edges EdgeReader, consent ConsentEvaluator, endpoints EndpointAuthorizer, now func() time.Time) DisclosurePolicy {
	if now == nil {
		now = time.Now
	}
	return DisclosurePolicy{edges: edges, consent: consent, endpoints: endpoints, now: now}
}

// Allows revalidates a projected step immediately before disclosure. Every
// failure is intentionally collapsed to false so callers cannot distinguish a
// withdrawn edge, hidden member, consent decision, expiry or dependency error.
func (policy DisclosurePolicy) Allows(ctx context.Context, requesterID string, step trustapplication.Step) bool {
	if policy.edges == nil || policy.consent == nil || policy.endpoints == nil {
		return false
	}
	edge, err := policy.edges.Find(ctx, step.EdgeID)
	if err != nil || edge.ID() != step.EdgeID || edge.SourceID() != step.SourceID ||
		edge.TargetID() != step.TargetID || edge.Type() != step.Type ||
		edge.ProvenanceRef() != step.ProvenanceRef || !edge.Active(policy.now().UTC()) {
		return false
	}
	consented, err := policy.consent.Allows(ctx, requesterID, edge.ConsentRef())
	if err != nil || !consented || !edge.VisibleTo(requesterID, consented, policy.now().UTC()) {
		return false
	}
	sourceVisible, err := policy.endpoints.CanReveal(ctx, requesterID, edge.SourceID())
	if err != nil || !sourceVisible {
		return false
	}
	targetVisible, err := policy.endpoints.CanReveal(ctx, requesterID, edge.TargetID())
	return err == nil && targetVisible
}

func reasonFor(edgeType domain.EdgeType) ReasonCode {
	switch edgeType {
	case domain.EdgeCircleMember:
		return ReasonSharedCircle
	case domain.EdgeVouch:
		return ReasonVouchedConnection
	case domain.EdgeKnown:
		return ReasonKnownConnection
	case domain.EdgeHost:
		return ReasonHostConnection
	default:
		return ""
	}
}
