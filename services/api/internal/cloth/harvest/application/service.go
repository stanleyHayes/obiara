package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/cloth/harvest/domain"
)

var (
	ErrNotFound     = errors.New("harvest not found")
	ErrConflict     = errors.New("harvest conflict")
	ErrApplied      = errors.New("harvest command applied")
	ErrNotAvailable = errors.New("harvest not available")
)

type Command struct {
	ID, HarvestID, ActorID, FirstMemberID, SecondMemberID string
	ExpectedRevision                                      uint64
}
type Draft struct {
	RecipeRef, RecipeVersion, RenderSeed, Format, DeliveryRef, PolicyVersion string
	ProductionTokens                                                         []string
}
type Service struct {
	r                      Repository
	a                      Authorizer
	pair                   PairPolicy
	owner                  Ownership
	recipes                RecipeValidator
	provider               ProviderAuth
	keyer                  Keyer
	harvestIDs, handoffIDs IDSource
	now                    func() time.Time
}

func NewService(r Repository, a Authorizer, pair PairPolicy, owner Ownership, recipes RecipeValidator, provider ProviderAuth, keyer Keyer, harvestIDs, handoffIDs IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, a, pair, owner, recipes, provider, keyer, harvestIDs, handoffIDs, now}
}

func (s Service) Create(ctx context.Context, c Command, draft Draft) (domain.Harvest, error) {
	if !s.ready() || c.ActorID != c.FirstMemberID || s.a.Require(ctx, c.ActorID, "cloth.harvest.create", "") != nil ||
		s.pair.Revalidate(ctx, c.FirstMemberID, c.SecondMemberID) != nil ||
		s.owner.Revalidate(ctx, c.FirstMemberID, c.SecondMemberID, draft.RecipeRef) != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	payload, err := s.payload(draft)
	if err != nil || s.recipes.Revalidate(ctx, payload) != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	members, err := s.members(c.FirstMemberID, c.SecondMemberID)
	if err != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	h, err := domain.Create(s.harvestIDs.NewID(), members, payload, s.command(c, members[0]))
	if err != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	if err = s.r.Create(ctx, h); err != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	return h, nil
}

func (s Service) Approve(ctx context.Context, c Command) (domain.Harvest, error) {
	h, actor, err := s.current(ctx, c, "cloth.harvest.approve")
	if err != nil {
		return domain.Harvest{}, err
	}
	if s.owner.Revalidate(ctx, c.FirstMemberID, c.SecondMemberID, h.Payload().RecipeKey) != nil ||
		s.recipes.Revalidate(ctx, h.Payload()) != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	next, err := h.Approve(s.command(c, actor))
	if err != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	if err = s.r.Append(ctx, next, h.Revision(), c.ID); err == nil {
		return next, nil
	}
	if errors.Is(err, ErrApplied) {
		old, findErr := s.r.FindByCommand(ctx, c.ID)
		if findErr == nil {
			return old, nil
		}
	}
	if !errors.Is(err, ErrConflict) {
		return domain.Harvest{}, ErrNotAvailable
	}
	fresh, findErr := s.r.Find(ctx, h.ID())
	if findErr != nil || fresh.Status() != domain.StatusAwaiting {
		return domain.Harvest{}, ErrNotAvailable
	}
	c.ExpectedRevision = fresh.Revision()
	converged, applyErr := fresh.Approve(s.command(c, actor))
	if applyErr != nil || s.r.Append(ctx, converged, fresh.Revision(), c.ID) != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	return converged, nil
}

func (s Service) Revise(ctx context.Context, c Command, draft Draft) (domain.Harvest, error) {
	h, actor, err := s.current(ctx, c, "cloth.harvest.revise")
	if err != nil {
		return domain.Harvest{}, err
	}
	payload, err := s.payload(draft)
	if err != nil || s.owner.Revalidate(ctx, c.FirstMemberID, c.SecondMemberID, draft.RecipeRef) != nil ||
		s.recipes.Revalidate(ctx, payload) != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	next, err := h.Revise(payload, s.command(c, actor))
	if err != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	return s.append(ctx, h, next, c.ID)
}

func (s Service) Handoff(ctx context.Context, c Command) (domain.Envelope, error) {
	h, actor, err := s.current(ctx, c, "cloth.harvest.handoff")
	if err != nil {
		return domain.Envelope{}, err
	}
	if s.owner.Revalidate(ctx, c.FirstMemberID, c.SecondMemberID, h.Payload().RecipeKey) != nil ||
		s.recipes.Revalidate(ctx, h.Payload()) != nil {
		return domain.Envelope{}, ErrNotAvailable
	}
	next, err := h.Handoff(s.handoffIDs.NewID(), s.command(c, actor))
	if err != nil {
		return domain.Envelope{}, ErrNotAvailable
	}
	saved, err := s.append(ctx, h, next, c.ID)
	if err != nil {
		return domain.Envelope{}, err
	}
	envelope, err := saved.ProviderEnvelope(s.now().UTC())
	if err != nil {
		return domain.Envelope{}, ErrNotAvailable
	}
	return envelope, nil
}

func (s Service) Cancel(ctx context.Context, c Command) error {
	h, actor, err := s.current(ctx, c, "cloth.harvest.cancel")
	if err != nil {
		return err
	}
	next, err := h.Cancel(s.command(c, actor))
	if err != nil {
		return ErrNotAvailable
	}
	_, err = s.append(ctx, h, next, c.ID)
	return err
}

func (s Service) ProviderFetch(ctx context.Context, providerID, handoffID string) (domain.Envelope, error) {
	if !s.ready() || s.provider.Require(ctx, providerID, handoffID) != nil {
		return domain.Envelope{}, ErrNotAvailable
	}
	h, err := s.r.FindByHandoff(ctx, strings.TrimSpace(handoffID))
	if err != nil {
		return domain.Envelope{}, ErrNotAvailable
	}
	envelope, err := h.ProviderEnvelope(s.now().UTC())
	if err != nil {
		return domain.Envelope{}, ErrNotAvailable
	}
	return envelope, nil
}

func (s Service) Callback(ctx context.Context, providerID, handoffID, reasonCode string, next domain.Status, c Command) (domain.Harvest, error) {
	if !s.ready() || s.provider.Require(ctx, providerID, handoffID) != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	h, err := s.r.FindByHandoff(ctx, strings.TrimSpace(handoffID))
	if err != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	updated, err := h.Callback(next, strings.TrimSpace(reasonCode), s.command(c, ""))
	if err != nil {
		return domain.Harvest{}, ErrNotAvailable
	}
	return s.append(ctx, h, updated, c.ID)
}

func (s Service) current(ctx context.Context, c Command, permission string) (domain.Harvest, string, error) {
	if !s.ready() || s.a.Require(ctx, c.ActorID, permission, c.HarvestID) != nil ||
		s.pair.Revalidate(ctx, c.FirstMemberID, c.SecondMemberID) != nil {
		return domain.Harvest{}, "", ErrNotAvailable
	}
	h, err := s.r.Find(ctx, strings.TrimSpace(c.HarvestID))
	if err != nil {
		return domain.Harvest{}, "", ErrNotAvailable
	}
	members, err := s.members(c.FirstMemberID, c.SecondMemberID)
	if err != nil || !slices.Equal(members, h.Members()) {
		return domain.Harvest{}, "", ErrNotAvailable
	}
	actor, err := s.keyer.Key("cloth-harvest:member", strings.TrimSpace(c.ActorID))
	if err != nil || !h.HasMember(actor) {
		return domain.Harvest{}, "", ErrNotAvailable
	}
	return h, actor, nil
}
func (s Service) payload(d Draft) (domain.Payload, error) {
	recipe, err := s.keyer.Key("cloth-harvest:recipe", strings.TrimSpace(d.RecipeRef))
	if err != nil {
		return domain.Payload{}, err
	}
	delivery, err := s.keyer.Key("cloth-harvest:delivery", strings.TrimSpace(d.DeliveryRef))
	if err != nil {
		return domain.Payload{}, err
	}
	return domain.Payload{RecipeKey: recipe, RecipeVersion: strings.TrimSpace(d.RecipeVersion), RenderSeed: strings.TrimSpace(d.RenderSeed),
		ProductionTokens: append([]string(nil), d.ProductionTokens...), Format: strings.TrimSpace(d.Format), DeliveryRef: delivery, PolicyVersion: strings.TrimSpace(d.PolicyVersion)}, nil
}
func (s Service) members(a, b string) ([]string, error) {
	x, e := s.keyer.Key("cloth-harvest:member", strings.TrimSpace(a))
	if e != nil {
		return nil, e
	}
	y, e := s.keyer.Key("cloth-harvest:member", strings.TrimSpace(b))
	if e != nil {
		return nil, e
	}
	v := []string{x, y}
	slices.Sort(v)
	return v, nil
}
func (s Service) command(c Command, actor string) domain.Command {
	return domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: actor, ExpectedRevision: c.ExpectedRevision, At: s.now().UTC()}
}
func (s Service) append(ctx context.Context, current, next domain.Harvest, id string) (domain.Harvest, error) {
	err := s.r.Append(ctx, next, current.Revision(), id)
	if err == nil {
		return next, nil
	}
	if errors.Is(err, ErrApplied) {
		old, e := s.r.FindByCommand(ctx, id)
		if e == nil {
			return old, nil
		}
	}
	return domain.Harvest{}, ErrNotAvailable
}
func (s Service) ready() bool {
	return s.r != nil && s.a != nil && s.pair != nil && s.owner != nil && s.recipes != nil && s.provider != nil && s.keyer != nil && s.harvestIDs != nil && s.handoffIDs != nil
}
