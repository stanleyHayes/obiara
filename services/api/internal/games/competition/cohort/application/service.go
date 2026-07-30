package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/cohort/domain"
)

var (
	ErrNotFound     = errors.New("competition cohort not found")
	ErrConflict     = errors.New("competition cohort conflict")
	ErrApplied      = errors.New("competition cohort command applied")
	ErrNotAvailable = errors.New("competition cohort not available")
)

type Command struct {
	ID, CohortID, ActorID string
	ExpectedRevision      uint64
}

type Projection struct {
	ID            string
	Capacity      int
	Enrolled      int
	Joined        bool
	Status        domain.Status
	CompetitionID string
	Revision      uint64
}

type Service struct {
	repository Repository
	admins     AdminAuthority
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func NewService(r Repository, a AdminAuthority, k Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: r, admins: a, keyer: k, ids: ids, now: now}
}

func (service Service) Create(ctx context.Context, command Command, capacity int) (Projection, error) {
	if !service.ready() || service.admins.RequireTournamentManager(ctx, command.ActorID) != nil {
		return Projection{}, ErrNotAvailable
	}
	cohort, err := domain.Create(service.ids.NewID(), capacity, service.command(command))
	if err != nil || service.repository.Create(ctx, cohort) != nil {
		return Projection{}, ErrNotAvailable
	}
	return project(cohort, ""), nil
}

func (service Service) Join(ctx context.Context, command Command) (Projection, error) {
	cohort, member, err := service.current(ctx, command)
	if err != nil {
		return Projection{}, err
	}
	next, err := cohort.Join(member, service.command(command))
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	return service.append(ctx, cohort, next, command.ID, member)
}

func (service Service) Leave(ctx context.Context, command Command) (Projection, error) {
	cohort, member, err := service.current(ctx, command)
	if err != nil {
		return Projection{}, err
	}
	next, err := cohort.Leave(member, service.command(command))
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	return service.append(ctx, cohort, next, command.ID, member)
}

func (service Service) View(ctx context.Context, command Command) (Projection, error) {
	cohort, member, err := service.current(ctx, command)
	if err != nil {
		return Projection{}, err
	}
	return project(cohort, member), nil
}

func (service Service) ViewForManager(ctx context.Context, command Command) (Projection, error) {
	if !service.ready() || service.admins.RequireTournamentManager(ctx, command.ActorID) != nil {
		return Projection{}, ErrNotAvailable
	}
	cohort, err := service.repository.Find(ctx, strings.TrimSpace(command.CohortID))
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	return project(cohort, ""), nil
}

// LockedEntrantKeys is an internal composition port. It never belongs in an
// HTTP projection and returns data only after manager authority revalidation.
func (service Service) LockedEntrantKeys(ctx context.Context, command Command) ([]string, error) {
	if !service.ready() || service.admins.RequireTournamentManager(ctx, command.ActorID) != nil {
		return nil, ErrNotAvailable
	}
	cohort, err := service.repository.Find(ctx, strings.TrimSpace(command.CohortID))
	if err != nil || cohort.Status() != domain.StatusLocked {
		return nil, ErrNotAvailable
	}
	return cohort.MemberKeys(), nil
}

func (service Service) MarkStarted(ctx context.Context, command Command, competitionID string) (Projection, error) {
	if !service.ready() || service.admins.RequireTournamentManager(ctx, command.ActorID) != nil {
		return Projection{}, ErrNotAvailable
	}
	cohort, err := service.repository.Find(ctx, strings.TrimSpace(command.CohortID))
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	next, err := cohort.Start(strings.TrimSpace(competitionID), service.command(command))
	if err != nil {
		return Projection{}, ErrNotAvailable
	}
	return service.append(ctx, cohort, next, command.ID, "")
}

func (service Service) RequireCohortMember(ctx context.Context, cohortID, actorID string) error {
	cohort, member, err := service.current(ctx, Command{CohortID: cohortID, ActorID: actorID})
	if err != nil || !slices.Contains(cohort.MemberKeys(), member) {
		return ErrNotAvailable
	}
	return nil
}

func (service Service) Revalidate(ctx context.Context, cohortID, actorID string) error {
	return service.RequireCohortMember(ctx, cohortID, actorID)
}

func (service Service) current(ctx context.Context, command Command) (domain.Cohort, string, error) {
	if !service.ready() {
		return domain.Cohort{}, "", ErrNotAvailable
	}
	cohort, err := service.repository.Find(ctx, strings.TrimSpace(command.CohortID))
	if err != nil {
		return domain.Cohort{}, "", ErrNotAvailable
	}
	member, err := service.keyer.Key("game-competition:entrant", strings.TrimSpace(command.ActorID))
	if err != nil {
		return domain.Cohort{}, "", ErrNotAvailable
	}
	return cohort, member, nil
}

func (service Service) append(ctx context.Context, current, next domain.Cohort, commandID, member string) (Projection, error) {
	err := service.repository.Append(ctx, next, current.Revision(), strings.TrimSpace(commandID))
	if err == nil {
		return project(next, member), nil
	}
	if errors.Is(err, ErrApplied) {
		if prior, findErr := service.repository.FindByCommand(ctx, strings.TrimSpace(commandID)); findErr == nil {
			return project(prior, member), nil
		}
	}
	return Projection{}, ErrConflict
}

func project(cohort domain.Cohort, member string) Projection {
	return Projection{
		ID: cohort.ID(), Capacity: cohort.Capacity(), Enrolled: len(cohort.MemberKeys()),
		Joined: member != "" && slices.Contains(cohort.MemberKeys(), member),
		Status: cohort.Status(), CompetitionID: cohort.CompetitionID(),
		Revision: cohort.Revision(),
	}
}

func (service Service) command(command Command) domain.Command {
	return domain.Command{
		ID: strings.TrimSpace(command.ID), ExpectedRevision: command.ExpectedRevision,
		At: service.now().UTC(),
	}
}

func (service Service) ready() bool {
	return service.repository != nil && service.admins != nil && service.keyer != nil && service.ids != nil
}
