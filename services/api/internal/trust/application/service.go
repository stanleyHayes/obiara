package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

type Service struct {
	repository Repository
	authorizer ProjectionAuthorizer
	consent    ConsentEvaluator
	now        func() time.Time
}

func NewService(repository Repository, authorizer ProjectionAuthorizer, consent ConsentEvaluator, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: repository, authorizer: authorizer, consent: consent, now: now}
}

type WriteCommand struct {
	ID               string
	EdgeID           string
	ActorRef         string
	ExpectedRevision uint64
}

type WriteResult struct {
	Edge     domain.Edge
	Replayed bool
}

func (service Service) Create(ctx context.Context, command WriteCommand, params domain.Params) (WriteResult, error) {
	if service.repository == nil {
		return WriteResult{}, ErrUnavailable
	}
	params.ID = command.EdgeID
	change := service.command(command, "edge.create", fmt.Sprintf(
		"%s:%s:%s:%s:%s:%s", params.SourceID, params.TargetID, params.Type,
		params.ProvenanceRef, params.ConsentRef, params.Visibility,
	))
	params.CreatedAt = change.RecordedAt
	edge, err := domain.Create(params, change)
	if err != nil {
		return WriteResult{}, err
	}
	return service.save(ctx, edge, command.ExpectedRevision, command.ID, func(reloaded domain.Edge) error {
		return reloaded.VerifyReplay(change)
	})
}

func (service Service) Revoke(ctx context.Context, command WriteCommand) (WriteResult, error) {
	if service.repository == nil {
		return WriteResult{}, ErrUnavailable
	}
	current, err := service.repository.Find(ctx, strings.TrimSpace(command.EdgeID))
	if err != nil {
		return WriteResult{}, err
	}
	change := service.command(command, "edge.revoke", command.EdgeID)
	wasApplied := current.HasCommand(command.ID)
	next, err := current.Revoke(change)
	if err != nil {
		return WriteResult{}, err
	}
	if wasApplied {
		return WriteResult{Edge: next, Replayed: true}, nil
	}
	return service.save(ctx, next, command.ExpectedRevision, command.ID, func(reloaded domain.Edge) error {
		_, replayErr := reloaded.Revoke(change)
		return replayErr
	})
}

func (service Service) save(ctx context.Context, edge domain.Edge, expected uint64, commandID string, verify func(domain.Edge) error) (WriteResult, error) {
	err := service.repository.Save(ctx, edge, expected, strings.TrimSpace(commandID))
	if err == nil {
		return WriteResult{Edge: edge}, nil
	}
	if errors.Is(err, ErrCommandAlreadyApplied) {
		reloaded, findErr := service.repository.Find(ctx, edge.ID())
		if findErr == nil && reloaded.HasCommand(commandID) {
			if replayErr := verify(reloaded); replayErr == nil {
				return WriteResult{Edge: reloaded, Replayed: true}, nil
			} else {
				return WriteResult{}, replayErr
			}
		}
		return WriteResult{}, domain.ErrCommandMismatch
	}
	if errors.Is(err, ErrOptimisticConflict) {
		return WriteResult{}, domain.ErrStaleRevision
	}
	return WriteResult{}, ErrUnavailable
}

func (service Service) command(command WriteCommand, kind, payload string) domain.Command {
	return domain.Command{
		ID: strings.TrimSpace(command.ID), ExpectedRevision: command.ExpectedRevision,
		ActorRef: strings.TrimSpace(command.ActorRef), Kind: kind,
		Payload: strings.TrimSpace(payload), RecordedAt: service.now().UTC(),
	}
}

type ProjectionRequest struct {
	OwnerID  string
	RootID   string
	MaxDepth int
	MaxNodes int
}

type Step struct {
	EdgeID        string
	SourceID      string
	TargetID      string
	Type          domain.EdgeType
	ProvenanceRef string
}

type Path struct {
	Steps []Step
}

type Projection struct {
	RootID string
	Paths  []Path
}

type frontierItem struct {
	nodeID string
	path   Path
	depth  int
}

func (service Service) Project(ctx context.Context, request ProjectionRequest) (Projection, error) {
	request.OwnerID, request.RootID = strings.TrimSpace(request.OwnerID), strings.TrimSpace(request.RootID)
	if service.repository == nil || service.authorizer == nil || service.consent == nil {
		return Projection{}, ErrUnavailable
	}
	if request.MaxDepth < 1 || request.MaxDepth > domain.MaxProjectionDepth ||
		request.MaxNodes < 1 || request.MaxNodes > domain.MaxProjectionNodes {
		return Projection{}, domain.ErrProjectionBounds
	}
	allowed, err := service.authorizer.CanProject(ctx, request.OwnerID, request.RootID)
	if err != nil || !allowed {
		return Projection{}, domain.ErrAccessDenied
	}
	projection := Projection{RootID: request.RootID}
	visited := map[string]struct{}{request.RootID: {}}
	frontier := []frontierItem{{nodeID: request.RootID}}
	now := service.now().UTC()
	for len(frontier) > 0 {
		sourceIDs := make([]string, 0, len(frontier))
		bySource := make(map[string]frontierItem, len(frontier))
		for _, item := range frontier {
			if item.depth < request.MaxDepth {
				sourceIDs = append(sourceIDs, item.nodeID)
				bySource[item.nodeID] = item
			}
		}
		if len(sourceIDs) == 0 {
			break
		}
		edges, err := service.repository.Outgoing(ctx, sourceIDs)
		if err != nil {
			return Projection{}, ErrUnavailable
		}
		next := make([]frontierItem, 0)
		for _, edge := range edges {
			parent, expected := bySource[edge.SourceID()]
			if !expected || !edge.Active(now) {
				continue
			}
			consented, err := service.consent.Allows(ctx, request.OwnerID, edge.ConsentRef())
			if err != nil || !consented || !edge.VisibleTo(request.OwnerID, consented, now) {
				continue
			}
			if _, cycle := visited[edge.TargetID()]; cycle {
				continue
			}
			if len(visited) >= request.MaxNodes {
				return projection, nil
			}
			visited[edge.TargetID()] = struct{}{}
			path := Path{Steps: append(append([]Step(nil), parent.path.Steps...), Step{
				EdgeID: edge.ID(), SourceID: edge.SourceID(), TargetID: edge.TargetID(),
				Type: edge.Type(), ProvenanceRef: edge.ProvenanceRef(),
			})}
			projection.Paths = append(projection.Paths, path)
			next = append(next, frontierItem{nodeID: edge.TargetID(), path: path, depth: parent.depth + 1})
		}
		frontier = next
	}
	return projection, nil
}
