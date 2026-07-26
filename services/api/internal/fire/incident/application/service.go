package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/incident/domain"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("incident not found")
	ErrConflict     = errors.New("incident conflict")
	ErrApplied      = errors.New("incident command applied")
	ErrNotAvailable = errors.New("incident not available")
)

type Action string

const (
	ActionLeave     Action = "leave"
	ActionMuteLocal Action = "mute_local"
)

type Command struct {
	ID, FireID, ActorID, EvidenceRef string
	Category                         domain.Category
	Action                           Action
}
type Service struct {
	r            Repository
	participants ParticipantAuthority
	safety       SafetyAction
	router       TrustSafetyRouter
	k            Keyer
	ids          IDSource
	now          func() time.Time
}

func NewService(r Repository, p ParticipantAuthority, s SafetyAction, router TrustSafetyRouter, k Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, p, s, router, k, ids, now}
}
func (s Service) Trigger(ctx context.Context, c Command) (domain.Projection, error) {
	if !s.ready() || (c.Action != ActionLeave && c.Action != ActionMuteLocal) || s.participants.RequireParticipant(ctx, c.FireID, c.ActorID) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	if s.safety.Apply(ctx, c.ID, c.FireID, c.ActorID, string(c.Action)) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	fire, e := s.k.Key("fire-incident:fire", strings.TrimSpace(c.FireID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	actor, e := s.k.Key("fire-incident:actor", strings.TrimSpace(c.ActorID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	evidence := ""
	if strings.TrimSpace(c.EvidenceRef) != "" {
		evidence, e = s.k.Key("fire-incident:evidence", strings.TrimSpace(c.EvidenceRef))
		if e != nil {
			return domain.Projection{}, ErrNotAvailable
		}
	}
	now := s.now().UTC()
	incident, e := domain.Create(s.ids.NewID(), fire, actor, c.Category, evidence, now, domain.Command{ID: strings.TrimSpace(c.ID), At: now})
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	if e = s.r.Create(ctx, incident); e != nil {
		if !errors.Is(e, ErrApplied) {
			return domain.Projection{}, ErrNotAvailable
		}
		incident, e = s.r.FindByCommand(ctx, c.ID)
		if e != nil {
			return domain.Projection{}, ErrNotAvailable
		}
		if incident.Status() == domain.StatusRouted {
			return incident.Project(), nil
		}
	}
	if s.router.Revalidate(ctx) != nil || s.router.Route(ctx, incident.Case()) != nil {
		return incident.Project(), ErrNotAvailable
	}
	routeID := strings.TrimSpace(c.ID) + ":route"
	next, e := incident.Route(s.now().UTC(), domain.Command{ID: routeID, ExpectedRevision: incident.Revision(), At: s.now().UTC()})
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	if e = s.r.Append(ctx, next, incident.Revision(), routeID); e != nil {
		if errors.Is(e, ErrApplied) {
			old, x := s.r.FindByCommand(ctx, routeID)
			if x == nil {
				return old.Project(), nil
			}
		}
		return domain.Projection{}, ErrNotAvailable
	}
	return next.Project(), nil
}
func (s Service) ready() bool {
	return s.r != nil && s.participants != nil && s.safety != nil && s.router != nil && s.k != nil && s.ids != nil
}
