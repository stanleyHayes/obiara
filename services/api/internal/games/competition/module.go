package competition

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/application"
	cohortmongodb "github.com/stanleyHayes/obiara/services/api/internal/games/competition/cohort/adapters/outbound/mongodb"
	cohortapp "github.com/stanleyHayes/obiara/services/api/internal/games/competition/cohort/application"
	owaremongodb "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/adapters/outbound/mongodb"
)

type Module struct {
	Cohorts      cohortapp.Service
	Competitions application.Service
	Manager      Manager
	Oware        TournamentOware
}

type Manager struct {
	cohorts      cohortapp.Service
	competitions application.Service
}

type StartCommand struct {
	ID, CohortID, ActorID string
	ExpectedRevision      uint64
}

type StartResult struct {
	Cohort      cohortapp.Projection
	Competition application.PrivateProjection
}

func (manager Manager) Start(ctx context.Context, command StartCommand) (StartResult, error) {
	keys, err := manager.cohorts.LockedEntrantKeys(ctx, cohortapp.Command{
		CohortID: command.CohortID, ActorID: command.ActorID,
	})
	if err != nil {
		return StartResult{}, err
	}
	competition, err := manager.competitions.CreateKeyed(ctx, application.Command{
		ID: command.ID, CohortID: command.CohortID, ActorID: command.ActorID,
	}, keys)
	if err != nil {
		return StartResult{}, err
	}
	cohort, err := manager.cohorts.MarkStarted(ctx, cohortapp.Command{
		ID: command.ID + ":cohort", CohortID: command.CohortID,
		ActorID: command.ActorID, ExpectedRevision: command.ExpectedRevision,
	}, competition.ID)
	if err != nil {
		return StartResult{}, err
	}
	return StartResult{Cohort: cohort, Competition: competition}, nil
}

func NewModule(
	ctx context.Context,
	database *mongo.Database,
	secret string,
	admins cohortapp.AdminAuthority,
) (Module, error) {
	cohortRepository := cohortmongodb.NewRepository(database)
	if err := cohortRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	competitionRepository := mongodb.NewRepository(database)
	if err := competitionRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	cohortKeyer, err := privacy.NewKeyer([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	competitionKeyer, err := privacy.NewKeyer([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	cohorts := cohortapp.NewService(
		cohortRepository, admins, cohortKeyer, idSource{"cohort_"}, time.Now,
	)
	owareRepository := owaremongodb.NewRepository(database)
	if err := owareRepository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	authority := composedAuthority{cohorts: cohorts, admins: admins}
	resultVerifier := owareResultVerifier{sessions: owareRepository, keyer: competitionKeyer}
	fairPlayVerifier := owareFairPlayVerifier{sessions: owareRepository, keyer: competitionKeyer}
	competitions := application.NewService(
		competitionRepository, authority, cohorts, resultVerifier,
		fairPlayVerifier, competitionKeyer,
		idSource{"competition_"}, idSource{"review_"}, time.Now,
	)
	return Module{
		Cohorts: cohorts, Competitions: competitions,
		Manager: Manager{cohorts: cohorts, competitions: competitions},
		Oware: TournamentOware{
			competitions: competitions, sessions: owareRepository,
			keyer: competitionKeyer, ids: idSource{"oware_comp_"}, now: time.Now,
		},
	}, nil
}

type composedAuthority struct {
	cohorts cohortapp.Service
	admins  cohortapp.AdminAuthority
}

func (authority composedAuthority) RequireCohortMember(ctx context.Context, cohortID, actorID string) error {
	return authority.cohorts.RequireCohortMember(ctx, cohortID, actorID)
}
func (authority composedAuthority) RequireReviewer(ctx context.Context, actorID string) error {
	return authority.admins.RequireTournamentManager(ctx, actorID)
}
func (composedAuthority) RevalidateCohort(context.Context, string, []string) error {
	return errors.New("raw entrant cohort creation is disabled")
}

type denyResultVerifier struct{}

func (denyResultVerifier) Revalidate(context.Context, string, string, string, string) error {
	return errors.New("verified game result adapter is not composed")
}

type denyFairPlayVerifier struct{}

func (denyFairPlayVerifier) Revalidate(context.Context, string, string, string) error {
	return errors.New("fair-play evidence adapter is not composed")
}

type idSource struct{ prefix string }

func (source idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return source.prefix + base64.RawURLEncoding.EncodeToString(value)
}
