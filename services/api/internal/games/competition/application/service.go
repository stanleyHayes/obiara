package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/domain"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("competition not found")
	ErrConflict     = errors.New("competition conflict")
	ErrApplied      = errors.New("competition command applied")
	ErrNotAvailable = errors.New("competition not available")
)

type Command struct {
	ID, CompetitionID, CohortID, ActorID string
	ExpectedRevision                     uint64
}
type Service struct {
	r                         Repository
	a                         Authority
	opt                       OptIn
	results                   ResultVerifier
	fair                      FairPlayVerifier
	k                         Keyer
	competitionIDs, reviewIDs IDSource
	now                       func() time.Time
}

func NewService(r Repository, a Authority, opt OptIn, results ResultVerifier, fair FairPlayVerifier, k Keyer, competitionIDs, reviewIDs IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, a, opt, results, fair, k, competitionIDs, reviewIDs, now}
}
func (s Service) Create(ctx context.Context, c Command, entrantIDs []string) (domain.Projection, error) {
	if !s.ready() || s.a.RevalidateCohort(ctx, c.CohortID, entrantIDs) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	entrants := make([]string, 0, len(entrantIDs))
	for _, id := range entrantIDs {
		if s.opt.Revalidate(ctx, c.CohortID, id) != nil {
			return domain.Projection{}, ErrNotAvailable
		}
		v, e := s.k.Key("game-competition:entrant", strings.TrimSpace(id))
		if e != nil {
			return domain.Projection{}, ErrNotAvailable
		}
		entrants = append(entrants, v)
	}
	cohort, e := s.k.Key("game-competition:cohort", strings.TrimSpace(c.CohortID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	x, e := domain.Create(s.competitionIDs.NewID(), cohort, entrants, s.command(c))
	if e != nil || s.r.Create(ctx, x) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return x.Project(), nil
}
func (s Service) RecordResult(ctx context.Context, c Command, matchID, resultRef, winnerID string) (domain.Projection, error) {
	x, e := s.member(ctx, c)
	if e != nil {
		return domain.Projection{}, e
	}
	if s.results.Revalidate(ctx, resultRef, matchID, winnerID) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	winner, e := s.k.Key("game-competition:entrant", strings.TrimSpace(winnerID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	result, e := s.k.Key("game-competition:result", strings.TrimSpace(resultRef))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	next, e := x.RecordResult(strings.TrimSpace(matchID), winner, result, s.command(c))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return s.append(ctx, x, next, c.ID)
}
func (s Service) OpenReview(ctx context.Context, c Command, matchID, evidenceRef string) (domain.Projection, error) {
	x, e := s.member(ctx, c)
	if e != nil {
		return domain.Projection{}, e
	}
	if s.fair.Revalidate(ctx, matchID, evidenceRef) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	actor, e := s.k.Key("game-competition:entrant", strings.TrimSpace(c.ActorID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	evidence, e := s.k.Key("game-competition:evidence", strings.TrimSpace(evidenceRef))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	next, e := x.OpenReview(s.reviewIDs.NewID(), strings.TrimSpace(matchID), evidence, actor, s.now().UTC(), s.command(c))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return s.append(ctx, x, next, c.ID)
}
func (s Service) ResolveReview(ctx context.Context, c Command, reviewID string, decision domain.Decision) (domain.Projection, error) {
	if !s.ready() || s.a.RequireReviewer(ctx, c.ActorID) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	x, e := s.r.Find(ctx, c.CompetitionID)
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	next, e := x.ResolveReview(reviewID, decision, s.now().UTC(), s.command(c))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return s.append(ctx, x, next, c.ID)
}
func (s Service) Appeal(ctx context.Context, c Command, reviewID string) (domain.Projection, error) {
	x, e := s.member(ctx, c)
	if e != nil {
		return domain.Projection{}, e
	}
	actor, e := s.k.Key("game-competition:entrant", strings.TrimSpace(c.ActorID))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	next, e := x.Appeal(reviewID, actor, s.command(c))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return s.append(ctx, x, next, c.ID)
}
func (s Service) ResolveAppeal(ctx context.Context, c Command, reviewID string, decision domain.Decision) (domain.Projection, error) {
	if !s.ready() || s.a.RequireReviewer(ctx, c.ActorID) != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	x, e := s.r.Find(ctx, c.CompetitionID)
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	next, e := x.ResolveAppeal(reviewID, decision, s.now().UTC(), s.command(c))
	if e != nil {
		return domain.Projection{}, ErrNotAvailable
	}
	return s.append(ctx, x, next, c.ID)
}
func (s Service) View(ctx context.Context, c Command) (domain.Projection, error) {
	x, e := s.member(ctx, c)
	if e != nil {
		return domain.Projection{}, e
	}
	return x.Project(), nil
}
func (s Service) member(ctx context.Context, c Command) (domain.Competition, error) {
	if !s.ready() || s.a.RequireCohortMember(ctx, c.CohortID, c.ActorID) != nil {
		return domain.Competition{}, ErrNotAvailable
	}
	x, e := s.r.Find(ctx, strings.TrimSpace(c.CompetitionID))
	if e != nil {
		return domain.Competition{}, ErrNotAvailable
	}
	cohort, e := s.k.Key("game-competition:cohort", strings.TrimSpace(c.CohortID))
	if e != nil || cohort != x.CohortKey() {
		return domain.Competition{}, ErrNotAvailable
	}
	actor, e := s.k.Key("game-competition:entrant", strings.TrimSpace(c.ActorID))
	if e != nil {
		return domain.Competition{}, ErrNotAvailable
	}
	found := false
	for _, v := range x.Entrants() {
		if v == actor {
			found = true
		}
	}
	if !found {
		return domain.Competition{}, ErrNotAvailable
	}
	return x, nil
}
func (s Service) append(ctx context.Context, current, next domain.Competition, id string) (domain.Projection, error) {
	e := s.r.Append(ctx, next, current.Revision(), id)
	if e == nil {
		return next.Project(), nil
	}
	if errors.Is(e, ErrApplied) {
		old, x := s.r.FindByCommand(ctx, id)
		if x == nil {
			return old.Project(), nil
		}
	}
	return domain.Projection{}, ErrNotAvailable
}
func (s Service) command(c Command) domain.Command {
	return domain.Command{ID: strings.TrimSpace(c.ID), ExpectedRevision: c.ExpectedRevision, At: s.now().UTC()}
}
func (s Service) ready() bool {
	return s.r != nil && s.a != nil && s.opt != nil && s.results != nil && s.fair != nil && s.k != nil && s.competitionIDs != nil && s.reviewIDs != nil
}
