package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe/domain"
)

var (
	ErrNotFound      = errors.New("ebe duel not found")
	ErrConflict      = errors.New("ebe duel conflict")
	ErrApplied       = errors.New("ebe command applied")
	ErrNotAvailable  = errors.New("ebe duel not available")
	ErrInvalidPrompt = errors.New("ebe prompt approval is invalid")
)

type PromptApproval struct {
	ID              string
	Version         uint64
	Language        string
	Cue             string
	AcceptedAnswers []string
	Source          domain.Source
}

type CatalogService struct {
	catalog   Catalog
	reviewers ReviewerAuthority
	keyer     Keyer
	reviewIDs IDSource
	now       func() time.Time
}

func NewCatalogService(c Catalog, a ReviewerAuthority, k Keyer, ids IDSource, now func() time.Time) CatalogService {
	if now == nil {
		now = time.Now
	}
	return CatalogService{c, a, k, ids, now}
}

func (s CatalogService) Approve(ctx context.Context, reviewerID string, input PromptApproval) (PromptProjection, error) {
	if s.catalog == nil || s.reviewers == nil || s.keyer == nil || s.reviewIDs == nil ||
		s.reviewers.RequireReviewer(ctx, reviewerID) != nil {
		return PromptProjection{}, ErrNotAvailable
	}
	reviewer, err := s.keyer.Key("ebe:reviewer", strings.TrimSpace(reviewerID))
	if err != nil {
		return PromptProjection{}, ErrNotAvailable
	}
	prompt, err := domain.NewApprovedPrompt(domain.PromptSpec{
		ID: strings.TrimSpace(input.ID), Version: input.Version,
		Language: strings.TrimSpace(input.Language), Cue: input.Cue,
		AcceptedAnswers: input.AcceptedAnswers, Source: input.Source,
		Review: domain.Review{ID: s.reviewIDs.NewID(), ReviewerKey: reviewer, Decision: domain.DecisionApproved, ReviewedAt: s.now().UTC()},
	})
	if err != nil {
		return PromptProjection{}, ErrInvalidPrompt
	}
	if err = s.catalog.SaveApproved(ctx, prompt); errors.Is(err, ErrConflict) {
		return PromptProjection{}, ErrConflict
	}
	if err != nil {
		return PromptProjection{}, ErrNotAvailable
	}
	return projectPrompt(prompt), nil
}

type Command struct {
	ID, DuelID, RoomID, ActorID, FirstPlayerID, SecondPlayerID string
	ExpectedRevision                                           uint64
}

type PromptProjection struct {
	ID             string
	Version        uint64
	Language       string
	Cue            string
	SourceKind     domain.SourceKind
	SourceCitation string
	SourceLocator  string
}

type TurnProjection struct {
	Number            uint64
	Prompt            PromptProjection
	Yours             bool
	YourAnswer        string
	YourAnswerCorrect *bool
}

type Projection struct {
	ID            string
	Revision      uint64
	Complete      bool
	YourTurn      bool
	CurrentPrompt *PromptProjection
	Turns         []TurnProjection
}

type DuelService struct {
	repository DuelRepository
	catalog    Catalog
	pairs      PairAuthority
	keyer      Keyer
	ids        IDSource
}

func NewDuelService(r DuelRepository, c Catalog, a PairAuthority, k Keyer, ids IDSource) DuelService {
	return DuelService{r, c, a, k, ids}
}

func (s DuelService) Create(ctx context.Context, command Command) (Projection, error) {
	if !s.ready() || command.ActorID != command.FirstPlayerID ||
		s.pairs.Revalidate(ctx, command.RoomID, command.FirstPlayerID, command.SecondPlayerID) != nil {
		return Projection{}, ErrNotAvailable
	}
	prompts, err := s.catalog.ListApproved(ctx, domain.MaxTurns)
	if err != nil || len(prompts) == 0 {
		return Projection{}, ErrNotAvailable
	}
	room, err := s.keyer.Key("ebe:room", strings.TrimSpace(command.RoomID))
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	first, err := s.keyer.Key("ebe:player", strings.TrimSpace(command.FirstPlayerID))
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	second, err := s.keyer.Key("ebe:player", strings.TrimSpace(command.SecondPlayerID))
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	duel, err := domain.NewDuel(domain.DuelSpec{ID: s.ids.NewID(), PlayerKeys: [2]string{first, second}, Prompts: prompts})
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	stored := StoredDuel{Duel: duel, RoomKey: room}
	if err = s.repository.Create(ctx, stored, strings.TrimSpace(command.ID)); err != nil {
		if errors.Is(err, ErrApplied) {
			if prior, findErr := s.repository.FindByCommand(ctx, strings.TrimSpace(command.ID)); findErr == nil {
				return project(prior.Duel, first)
			}
		}
		return Projection{}, ErrNotAvailable
	}
	return project(duel, first)
}

func (s DuelService) Answer(ctx context.Context, command Command, answer string) (Projection, error) {
	stored, actor, err := s.current(ctx, command)
	if err != nil {
		return Projection{}, err
	}
	next, err := stored.Duel.Answer(actor, answer, command.ExpectedRevision)
	if errors.Is(err, domain.ErrStaleRevision) {
		return Projection{}, ErrConflict
	}
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	updated := StoredDuel{Duel: next, RoomKey: stored.RoomKey}
	if err = s.repository.Append(ctx, updated, stored.Duel.Revision(), strings.TrimSpace(command.ID)); err != nil {
		if errors.Is(err, ErrApplied) {
			if prior, findErr := s.repository.FindByCommand(ctx, strings.TrimSpace(command.ID)); findErr == nil {
				return project(prior.Duel, actor)
			}
		}
		return Projection{}, ErrConflict
	}
	return project(next, actor)
}

func (s DuelService) View(ctx context.Context, command Command) (Projection, error) {
	stored, actor, err := s.current(ctx, command)
	if err != nil {
		return Projection{}, err
	}
	return project(stored.Duel, actor)
}

func (s DuelService) current(ctx context.Context, command Command) (StoredDuel, string, error) {
	if !s.ready() || s.pairs.Revalidate(ctx, command.RoomID, command.FirstPlayerID, command.SecondPlayerID) != nil {
		return StoredDuel{}, "", ErrNotAvailable
	}
	stored, err := s.repository.Find(ctx, strings.TrimSpace(command.DuelID))
	if err != nil {
		return StoredDuel{}, "", ErrNotAvailable
	}
	room, err := s.keyer.Key("ebe:room", strings.TrimSpace(command.RoomID))
	if err != nil || room != stored.RoomKey {
		return StoredDuel{}, "", ErrNotAvailable
	}
	actor, err := s.keyer.Key("ebe:player", strings.TrimSpace(command.ActorID))
	spec := stored.Duel.Specification()
	if err != nil || !slices.Contains(spec.PlayerKeys[:], actor) {
		return StoredDuel{}, "", ErrNotAvailable
	}
	return stored, actor, nil
}

func project(duel domain.Duel, actor string) (Projection, error) {
	spec := duel.Specification()
	index := slices.Index(spec.PlayerKeys[:], actor)
	if index < 0 {
		return Projection{}, ErrNotAvailable
	}
	turns := duel.Turns()
	result := Projection{ID: spec.ID, Revision: duel.Revision(), Complete: duel.Complete(), YourTurn: !duel.Complete() && len(turns)%2 == index, Turns: make([]TurnProjection, 0, len(turns))}
	for position, turn := range turns {
		yours := turn.PlayerKey == actor
		item := TurnProjection{Number: turn.Number, Prompt: projectPrompt(spec.Prompts[position]), Yours: yours}
		if yours {
			correct := turn.Correct
			item.YourAnswer, item.YourAnswerCorrect = turn.Answer, &correct
		}
		result.Turns = append(result.Turns, item)
	}
	if !duel.Complete() {
		current := projectPrompt(spec.Prompts[len(turns)])
		result.CurrentPrompt = &current
	}
	return result, nil
}

func projectPrompt(prompt domain.Prompt) PromptProjection {
	spec := prompt.Spec()
	return PromptProjection{ID: spec.ID, Version: spec.Version, Language: spec.Language, Cue: spec.Cue, SourceKind: spec.Source.Kind, SourceCitation: spec.Source.Citation, SourceLocator: spec.Source.Locator}
}

func (s DuelService) ready() bool {
	return s.repository != nil && s.catalog != nil && s.pairs != nil && s.keyer != nil && s.ids != nil
}
