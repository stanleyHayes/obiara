package application

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/domain"
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

type PrivateMatch struct {
	ID             string `json:"id"`
	Round          int    `json:"round"`
	Slot           int    `json:"slot"`
	FirstLabel     string `json:"firstLabel"`
	SecondLabel    string `json:"secondLabel"`
	WinnerLabel    string `json:"winnerLabel,omitempty"`
	ResultRecorded bool   `json:"resultRecorded"`
	YouArePlaying  bool   `json:"youArePlaying"`
}
type PrivateLadderEntry struct {
	Label  string `json:"label"`
	Played int    `json:"played"`
	Wins   int    `json:"wins"`
	You    bool   `json:"you"`
}
type PrivateReview struct {
	ID         string              `json:"id"`
	MatchID    string              `json:"matchId"`
	Status     domain.ReviewStatus `json:"status"`
	Decision   domain.Decision     `json:"decision"`
	Yours      bool                `json:"yours"`
	OpenedAt   time.Time           `json:"openedAt"`
	ResolvedAt time.Time           `json:"resolvedAt,omitempty"`
}
type PrivateProjection struct {
	ID       string
	Status   domain.Status
	Revision uint64
	Matches  []PrivateMatch
	Ladder   []PrivateLadderEntry
	Reviews  []PrivateReview
}
type MatchAccess struct {
	CompetitionID string
	MatchID       string
	ActorKey      string
	FirstKey      string
	SecondKey     string
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

// CreateKeyed is the composition boundary for a locked opt-in cohort. Entrant
// keys are already privacy-keyed by the cohort service and never cross HTTP.
func (s Service) CreateKeyed(ctx context.Context, c Command, entrantKeys []string) (PrivateProjection, error) {
	if !s.ready() || len(entrantKeys) < domain.MinCohort || len(entrantKeys) > domain.MaxCohort {
		return PrivateProjection{}, ErrNotAvailable
	}
	cohort, err := s.k.Key("game-competition:cohort", strings.TrimSpace(c.CohortID))
	if err != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	competition, err := domain.Create(s.competitionIDs.NewID(), cohort, entrantKeys, s.command(c))
	if err != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	if err = s.r.Create(ctx, competition); err != nil {
		if errors.Is(err, ErrApplied) {
			if prior, findErr := s.r.FindByCommand(ctx, strings.TrimSpace(c.ID)); findErr == nil {
				return privateProject(prior, ""), nil
			}
		}
		return PrivateProjection{}, ErrNotAvailable
	}
	return privateProject(competition, ""), nil
}
func (s Service) RecordResult(ctx context.Context, c Command, matchID, resultRef, winnerID string) (domain.Projection, error) {
	x, e := s.member(ctx, c)
	if e != nil {
		return domain.Projection{}, e
	}
	if s.results.Revalidate(ctx, resultRef, c.CompetitionID, matchID, winnerID) != nil {
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

// RecordVerifiedResult is reserved for an in-process authoritative game
// adapter. The adapter supplies the winner privacy key only after re-reading
// the completed server-owned session bound to this exact competition match.
func (s Service) RecordVerifiedResult(ctx context.Context, c Command, matchID, resultRef, winnerKey string) (PrivateProjection, error) {
	x, err := s.member(ctx, c)
	if err != nil {
		return PrivateProjection{}, err
	}
	result, err := s.k.Key("game-competition:result", strings.TrimSpace(resultRef))
	if err != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	next, err := x.RecordResult(strings.TrimSpace(matchID), strings.TrimSpace(winnerKey), result, s.command(c))
	if err != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	if _, err = s.append(ctx, x, next, c.ID); err != nil {
		return PrivateProjection{}, err
	}
	actor, err := s.k.Key("game-competition:entrant", strings.TrimSpace(c.ActorID))
	if err != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	return privateProject(next, actor), nil
}

// AccessMatch is an internal game-composition boundary. It returns only
// privacy keys after current cohort membership and match participation have
// both been revalidated; it must never be projected over HTTP.
func (s Service) AccessMatch(ctx context.Context, c Command, matchID string) (MatchAccess, error) {
	x, err := s.member(ctx, c)
	if err != nil {
		return MatchAccess{}, err
	}
	actor, err := s.k.Key("game-competition:entrant", strings.TrimSpace(c.ActorID))
	if err != nil {
		return MatchAccess{}, ErrNotAvailable
	}
	for _, match := range x.Matches() {
		if match.ID == strings.TrimSpace(matchID) && match.WinnerKey == "" &&
			(match.FirstKey == actor || match.SecondKey == actor) {
			return MatchAccess{
				CompetitionID: x.ID(), MatchID: match.ID, ActorKey: actor,
				FirstKey: match.FirstKey, SecondKey: match.SecondKey,
			}, nil
		}
	}
	return MatchAccess{}, ErrNotAvailable
}
func (s Service) OpenReview(ctx context.Context, c Command, matchID, evidenceRef string) (domain.Projection, error) {
	x, e := s.member(ctx, c)
	if e != nil {
		return domain.Projection{}, e
	}
	if s.fair.Revalidate(ctx, c.CompetitionID, matchID, evidenceRef) != nil {
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

func (s Service) OpenReviewPrivate(ctx context.Context, c Command, matchID, evidenceRef string) (PrivateProjection, error) {
	if _, err := s.OpenReview(ctx, c, matchID, evidenceRef); err != nil {
		return PrivateProjection{}, err
	}
	return s.ViewPrivate(ctx, c)
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
func (s Service) ResolveReviewPrivate(ctx context.Context, c Command, reviewID string, decision domain.Decision) (PrivateProjection, error) {
	if _, err := s.ResolveReview(ctx, c, reviewID, decision); err != nil {
		return PrivateProjection{}, err
	}
	x, err := s.r.Find(ctx, strings.TrimSpace(c.CompetitionID))
	if err != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	return privateProject(x, ""), nil
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
func (s Service) AppealPrivate(ctx context.Context, c Command, reviewID string) (PrivateProjection, error) {
	if _, err := s.Appeal(ctx, c, reviewID); err != nil {
		return PrivateProjection{}, err
	}
	return s.ViewPrivate(ctx, c)
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
func (s Service) ResolveAppealPrivate(ctx context.Context, c Command, reviewID string, decision domain.Decision) (PrivateProjection, error) {
	if _, err := s.ResolveAppeal(ctx, c, reviewID, decision); err != nil {
		return PrivateProjection{}, err
	}
	x, err := s.r.Find(ctx, strings.TrimSpace(c.CompetitionID))
	if err != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	return privateProject(x, ""), nil
}
func (s Service) View(ctx context.Context, c Command) (domain.Projection, error) {
	x, e := s.member(ctx, c)
	if e != nil {
		return domain.Projection{}, e
	}
	return x.Project(), nil
}

func (s Service) ViewPrivate(ctx context.Context, c Command) (PrivateProjection, error) {
	competition, err := s.member(ctx, c)
	if err != nil {
		return PrivateProjection{}, err
	}
	actor, err := s.k.Key("game-competition:entrant", strings.TrimSpace(c.ActorID))
	if err != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	return privateProject(competition, actor), nil
}

func (s Service) ViewForReviewer(ctx context.Context, c Command) (PrivateProjection, error) {
	if !s.ready() || s.a.RequireReviewer(ctx, c.ActorID) != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	x, err := s.r.Find(ctx, strings.TrimSpace(c.CompetitionID))
	if err != nil {
		return PrivateProjection{}, ErrNotAvailable
	}
	cohort, err := s.k.Key("game-competition:cohort", strings.TrimSpace(c.CohortID))
	if err != nil || cohort != x.CohortKey() {
		return PrivateProjection{}, ErrNotAvailable
	}
	return privateProject(x, ""), nil
}

func privateProject(competition domain.Competition, actor string) PrivateProjection {
	entrants := competition.Entrants()
	label := func(key string) string {
		index := slices.Index(entrants, key)
		if index < 0 {
			return ""
		}
		if key == actor {
			return "You"
		}
		return "Entrant " + strconv.Itoa(index+1)
	}
	result := PrivateProjection{
		ID: competition.ID(), Status: competition.Status(), Revision: competition.Revision(),
		Matches: make([]PrivateMatch, 0, len(competition.Matches())),
		Ladder:  make([]PrivateLadderEntry, 0, len(competition.Ladder())),
		Reviews: make([]PrivateReview, 0, len(competition.Reviews())),
	}
	for _, match := range competition.Matches() {
		result.Matches = append(result.Matches, PrivateMatch{
			ID: match.ID, Round: match.Round, Slot: match.Slot,
			FirstLabel: label(match.FirstKey), SecondLabel: label(match.SecondKey),
			WinnerLabel: label(match.WinnerKey), ResultRecorded: match.ResultKey != "",
			YouArePlaying: match.FirstKey == actor || match.SecondKey == actor,
		})
	}
	for _, entry := range competition.Ladder() {
		result.Ladder = append(result.Ladder, PrivateLadderEntry{
			Label: label(entry.MemberKey), Played: entry.Played,
			Wins: entry.Wins, You: entry.MemberKey == actor,
		})
	}
	for _, review := range competition.Reviews() {
		result.Reviews = append(result.Reviews, PrivateReview{
			ID: review.ID, MatchID: review.MatchID, Status: review.Status,
			Decision: review.Decision, Yours: review.OpenedByKey == actor,
			OpenedAt: review.OpenedAt, ResolvedAt: review.ResolvedAt,
		})
	}
	return result
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
