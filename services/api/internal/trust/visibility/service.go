package visibility

import (
	"context"
	"errors"
	"sort"
	"strings"

	trustapplication "github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

const (
	DefaultDepth = 3
	DefaultNodes = 50
)

type ReasonCode string

const (
	ReasonSharedCircle      ReasonCode = "shared_circle"
	ReasonVouchedConnection ReasonCode = "vouched_connection"
	ReasonKnownConnection   ReasonCode = "known_connection"
	ReasonHostConnection    ReasonCode = "host_connection"
)

type ExplanationStep struct {
	SourceID string     `json:"sourceId"`
	TargetID string     `json:"targetId"`
	Reason   ReasonCode `json:"reason"`
}

type Explanation struct {
	TargetID string            `json:"targetId"`
	Hops     int               `json:"hops"`
	Steps    []ExplanationStep `json:"steps"`
}

type Request struct {
	RequesterID string
	RootID      string
	MaxDepth    int
	MaxNodes    int
}

type Service struct {
	projector Projector
	policy    DisclosurePolicy
}

func NewService(projector Projector, policy DisclosurePolicy) Service {
	return Service{projector: projector, policy: policy}
}

func (service Service) Explain(ctx context.Context, request Request) ([]Explanation, error) {
	request.RequesterID, request.RootID = strings.TrimSpace(request.RequesterID), strings.TrimSpace(request.RootID)
	if request.MaxDepth == 0 {
		request.MaxDepth = DefaultDepth
	}
	if request.MaxNodes == 0 {
		request.MaxNodes = DefaultNodes
	}
	if request.MaxDepth < 1 || request.MaxDepth > domain.MaxProjectionDepth ||
		request.MaxNodes < 2 || request.MaxNodes > domain.MaxProjectionNodes {
		return nil, ErrInvalidBounds
	}
	if request.RequesterID == "" || request.RootID == "" || service.projector == nil {
		return nil, ErrNotVisible
	}
	projection, err := service.projector.Project(ctx, trustapplication.ProjectionRequest{
		OwnerID: request.RequesterID, RootID: request.RootID,
		MaxDepth: request.MaxDepth, MaxNodes: request.MaxNodes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrProjectionBounds) {
			return nil, ErrInvalidBounds
		}
		return nil, ErrNotVisible
	}
	explanations := make([]Explanation, 0, len(projection.Paths))
	for _, path := range projection.Paths {
		if len(path.Steps) == 0 || len(path.Steps) > request.MaxDepth {
			continue
		}
		explanation := Explanation{Hops: len(path.Steps)}
		visible := true
		for _, step := range path.Steps {
			reason := reasonFor(step.Type)
			if reason == "" || !service.policy.Allows(ctx, request.RequesterID, step) {
				visible = false
				break
			}
			explanation.Steps = append(explanation.Steps, ExplanationStep{
				SourceID: step.SourceID, TargetID: step.TargetID, Reason: reason,
			})
		}
		if !visible {
			continue
		}
		explanation.TargetID = path.Steps[len(path.Steps)-1].TargetID
		explanations = append(explanations, explanation)
	}
	sort.Slice(explanations, func(i, j int) bool {
		if explanations[i].TargetID != explanations[j].TargetID {
			return explanations[i].TargetID < explanations[j].TargetID
		}
		if explanations[i].Hops != explanations[j].Hops {
			return explanations[i].Hops < explanations[j].Hops
		}
		return explanationKey(explanations[i]) < explanationKey(explanations[j])
	})
	return explanations, nil
}

func explanationKey(explanation Explanation) string {
	var builder strings.Builder
	for _, step := range explanation.Steps {
		builder.WriteString(step.SourceID)
		builder.WriteByte(0)
		builder.WriteString(step.TargetID)
		builder.WriteByte(0)
		builder.WriteString(string(step.Reason))
		builder.WriteByte(0)
	}
	return builder.String()
}
