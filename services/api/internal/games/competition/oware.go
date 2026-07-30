package competition

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/application"
	owaresessionapp "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/application"
	owaresession "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/domain"
)

type owareSessions interface {
	Create(context.Context, owaresession.Session) error
	Find(context.Context, string) (owaresession.Session, error)
	FindByCommand(context.Context, string) (owaresession.Session, error)
	FindCurrentByRoom(context.Context, string) (owaresession.Session, error)
	Append(context.Context, owaresession.Session, uint64, string) error
}

type TournamentOwareCommand struct {
	ID, CohortID, CompetitionID, MatchID, SessionID, ActorID string
	ExpectedRevision, ExpectedCompetitionRevision            uint64
}

func (service TournamentOware) Finalize(ctx context.Context, command TournamentOwareCommand) (application.PrivateProjection, error) {
	game, access, err := service.current(ctx, command)
	if err != nil || game.Status() != owaresession.StatusCompleted {
		return application.PrivateProjection{}, application.ErrNotAvailable
	}
	winnerIndex := game.Board().Winner()
	players := game.Players()
	if winnerIndex < 0 || winnerIndex > 1 {
		return application.PrivateProjection{}, application.ErrNotAvailable
	}
	return service.competitions.RecordVerifiedResult(ctx, application.Command{
		ID: strings.TrimSpace(command.ID), CompetitionID: access.CompetitionID,
		CohortID: strings.TrimSpace(command.CohortID), ActorID: strings.TrimSpace(command.ActorID),
		ExpectedRevision: command.ExpectedCompetitionRevision,
	}, access.MatchID, game.ID(), players[winnerIndex])
}

type TournamentOwareProjection struct {
	ID                       string
	Houses                   [12]int
	Captured                 [2]int
	YourPlayer, Turn, Winner int
	Status                   owaresession.Status
	Revision                 uint64
	MoveDeadline, ServerTime time.Time
}

type TournamentOware struct {
	competitions application.Service
	sessions     owareSessions
	keyer        application.Keyer
	ids          application.IDSource
	now          func() time.Time
}

func (service TournamentOware) Launch(ctx context.Context, command TournamentOwareCommand, window time.Duration) (TournamentOwareProjection, error) {
	access, err := service.access(ctx, command)
	if err != nil {
		return TournamentOwareProjection{}, err
	}
	room, err := service.matchKey(access.CompetitionID, access.MatchID)
	if err != nil {
		return TournamentOwareProjection{}, application.ErrNotAvailable
	}
	now := service.now().UTC()
	if current, findErr := service.sessions.FindCurrentByRoom(ctx, room); findErr == nil {
		winner := current.Board().Winner()
		if current.Status() == owaresession.StatusActive || winner == 0 || winner == 1 {
			return projectTournamentOware(current, access.ActorKey, now), nil
		}
	}
	game, err := owaresession.Create(
		service.ids.NewID(), room, []string{access.FirstKey, access.SecondKey},
		window, now, owaresession.Command{ID: strings.TrimSpace(command.ID), At: now},
	)
	if err != nil {
		return TournamentOwareProjection{}, application.ErrNotAvailable
	}
	if err = service.sessions.Create(ctx, game); err != nil {
		if errors.Is(err, owaresessionapp.ErrApplied) {
			if prior, findErr := service.sessions.FindByCommand(ctx, strings.TrimSpace(command.ID)); findErr == nil {
				return projectTournamentOware(prior, access.ActorKey, now), nil
			}
		}
		if current, findErr := service.sessions.FindCurrentByRoom(ctx, room); findErr == nil &&
			current.Status() == owaresession.StatusActive {
			return projectTournamentOware(current, access.ActorKey, now), nil
		}
		return TournamentOwareProjection{}, application.ErrNotAvailable
	}
	return projectTournamentOware(game, access.ActorKey, now), nil
}

func (service TournamentOware) View(ctx context.Context, command TournamentOwareCommand) (TournamentOwareProjection, error) {
	game, access, err := service.current(ctx, command)
	if err != nil {
		return TournamentOwareProjection{}, err
	}
	return projectTournamentOware(game, access.ActorKey, service.now().UTC()), nil
}

func (service TournamentOware) Move(ctx context.Context, command TournamentOwareCommand, pit int) (TournamentOwareProjection, error) {
	game, access, err := service.current(ctx, command)
	if err != nil {
		return TournamentOwareProjection{}, err
	}
	now := service.now().UTC()
	next, err := game.Move(
		access.ActorKey, pit, now,
		owaresession.Command{ID: strings.TrimSpace(command.ID), ExpectedRevision: command.ExpectedRevision, At: now},
	)
	if err != nil {
		return TournamentOwareProjection{}, application.ErrNotAvailable
	}
	if err = service.sessions.Append(ctx, next, game.Revision(), strings.TrimSpace(command.ID)); err != nil {
		if errors.Is(err, owaresessionapp.ErrApplied) {
			if prior, findErr := service.sessions.FindByCommand(ctx, strings.TrimSpace(command.ID)); findErr == nil {
				return projectTournamentOware(prior, access.ActorKey, now), nil
			}
		}
		return TournamentOwareProjection{}, application.ErrConflict
	}
	return projectTournamentOware(next, access.ActorKey, now), nil
}

func (service TournamentOware) current(ctx context.Context, command TournamentOwareCommand) (owaresession.Session, application.MatchAccess, error) {
	access, err := service.access(ctx, command)
	if err != nil {
		return owaresession.Session{}, application.MatchAccess{}, err
	}
	game, err := service.sessions.Find(ctx, strings.TrimSpace(command.SessionID))
	if err != nil {
		return owaresession.Session{}, application.MatchAccess{}, application.ErrNotAvailable
	}
	room, err := service.matchKey(access.CompetitionID, access.MatchID)
	players := []string{access.FirstKey, access.SecondKey}
	slices.Sort(players)
	if err != nil || game.RoomRef() != room || !slices.Equal(game.Players(), players) {
		return owaresession.Session{}, application.MatchAccess{}, application.ErrNotAvailable
	}
	if game.Status() == owaresession.StatusActive && !service.now().UTC().Before(game.Deadline()) {
		now := service.now().UTC()
		expired, expireErr := game.Expire(now, owaresession.Command{
			ID: "competition-expire:" + game.ID(), ExpectedRevision: game.Revision(), At: now,
		})
		if expireErr == nil {
			if appendErr := service.sessions.Append(ctx, expired, game.Revision(), "competition-expire:"+game.ID()); appendErr == nil {
				game = expired
			} else if current, findErr := service.sessions.Find(ctx, game.ID()); findErr == nil {
				game = current
			}
		}
	}
	return game, access, nil
}

func (service TournamentOware) access(ctx context.Context, command TournamentOwareCommand) (application.MatchAccess, error) {
	if service.sessions == nil || service.keyer == nil || service.ids == nil || service.now == nil {
		return application.MatchAccess{}, application.ErrNotAvailable
	}
	return service.competitions.AccessMatch(ctx, application.Command{
		CompetitionID: strings.TrimSpace(command.CompetitionID),
		CohortID:      strings.TrimSpace(command.CohortID),
		ActorID:       strings.TrimSpace(command.ActorID),
	}, strings.TrimSpace(command.MatchID))
}

func (service TournamentOware) matchKey(competitionID, matchID string) (string, error) {
	return service.keyer.Key("game-competition:oware-match", strings.TrimSpace(competitionID)+":"+strings.TrimSpace(matchID))
}

func projectTournamentOware(game owaresession.Session, actor string, now time.Time) TournamentOwareProjection {
	board := game.Board()
	players := game.Players()
	yourPlayer := slices.Index(players, actor)
	return TournamentOwareProjection{
		ID: game.ID(), Houses: board.Houses(), Captured: board.Captured(),
		YourPlayer: yourPlayer, Turn: int(game.Turn()), Winner: board.Winner(),
		Status: game.Status(), Revision: game.Revision(),
		MoveDeadline: game.Deadline(), ServerTime: now.UTC(),
	}
}

type owareResultVerifier struct {
	sessions owareSessions
	keyer    application.Keyer
}

type owareFairPlayVerifier struct {
	sessions owareSessions
	keyer    application.Keyer
}

func (verifier owareFairPlayVerifier) Revalidate(ctx context.Context, competitionID, matchID, evidenceRef string) error {
	game, err := verifier.sessions.Find(ctx, strings.TrimSpace(evidenceRef))
	if err != nil || game.Status() != owaresession.StatusExpired {
		return application.ErrNotAvailable
	}
	room, err := verifier.keyer.Key(
		"game-competition:oware-match",
		strings.TrimSpace(competitionID)+":"+strings.TrimSpace(matchID),
	)
	if err != nil || game.RoomRef() != room {
		return application.ErrNotAvailable
	}
	return nil
}

func (verifier owareResultVerifier) Revalidate(ctx context.Context, resultRef, competitionID, matchID, winnerID string) error {
	game, err := verifier.sessions.Find(ctx, strings.TrimSpace(resultRef))
	if err != nil || game.Status() != owaresession.StatusCompleted {
		return application.ErrNotAvailable
	}
	room, err := verifier.keyer.Key(
		"game-competition:oware-match",
		strings.TrimSpace(competitionID)+":"+strings.TrimSpace(matchID),
	)
	if err != nil || game.RoomRef() != room {
		return application.ErrNotAvailable
	}
	winner, err := verifier.keyer.Key("game-competition:entrant", strings.TrimSpace(winnerID))
	winnerIndex := game.Board().Winner()
	players := game.Players()
	if err != nil || winnerIndex < 0 || winnerIndex > 1 || players[winnerIndex] != winner {
		return application.ErrNotAvailable
	}
	return nil
}
