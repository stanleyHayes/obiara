package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/domain"
	"slices"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("anansesem story not found")
	ErrConflict     = errors.New("anansesem conflict")
	ErrApplied      = errors.New("anansesem command applied")
	ErrNotAvailable = errors.New("anansesem not available")
)

type Command struct {
	ID, StoryID, RoomID, ActorID, FirstAuthorID, SecondAuthorID string
	ExpectedRevision                                            uint64
}

type PassageView struct {
	ID        string
	Ordinal   int
	Content   string
	Yours     bool
	CreatedAt time.Time
	EditedAt  time.Time
}

type Projection struct {
	ID          string
	TitleCode   string
	Passages    []PassageView
	YourTurn    bool
	YourGrant   bool
	OtherGrant  bool
	BothGranted bool
	Editions    []domain.Edition
	Revision    uint64
}
type Service struct {
	r                    Repository
	a                    Authority
	k                    Keyer
	storyIDs, passageIDs IDSource
	now                  func() time.Time
}

func NewService(r Repository, a Authority, k Keyer, storyIDs, passageIDs IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, a, k, storyIDs, passageIDs, now}
}
func (s Service) Create(ctx context.Context, c Command, title string) (domain.Story, error) {
	if !s.ready() || c.ActorID != c.FirstAuthorID || s.a.RevalidateAuthors(ctx, c.RoomID, c.FirstAuthorID, c.SecondAuthorID) != nil {
		return domain.Story{}, ErrNotAvailable
	}
	room, e := s.k.Key("anansesem:room", strings.TrimSpace(c.RoomID))
	if e != nil {
		return domain.Story{}, ErrNotAvailable
	}
	authors, e := s.authors(c.FirstAuthorID, c.SecondAuthorID)
	if e != nil {
		return domain.Story{}, ErrNotAvailable
	}
	story, e := domain.Create(s.storyIDs.NewID(), room, strings.TrimSpace(title), authors, s.command(c))
	if e != nil || s.r.Create(ctx, story) != nil {
		return domain.Story{}, ErrNotAvailable
	}
	return story, nil
}
func (s Service) Add(ctx context.Context, c Command, content string) (domain.Story, error) {
	story, actor, e := s.current(ctx, c)
	if e != nil {
		return domain.Story{}, e
	}
	next, e := story.Add(s.passageIDs.NewID(), actor, content, s.now().UTC(), s.command(c))
	if e != nil {
		return domain.Story{}, ErrNotAvailable
	}
	return s.append(ctx, story, next, c.ID)
}
func (s Service) Edit(ctx context.Context, c Command, passageID, content string) (domain.Story, error) {
	story, actor, e := s.current(ctx, c)
	if e != nil {
		return domain.Story{}, e
	}
	next, e := story.Edit(strings.TrimSpace(passageID), actor, content, s.now().UTC(), s.command(c))
	if e != nil {
		return domain.Story{}, ErrNotAvailable
	}
	return s.append(ctx, story, next, c.ID)
}
func (s Service) Grant(ctx context.Context, c Command) (domain.Story, error) {
	story, actor, e := s.current(ctx, c)
	if e != nil {
		return domain.Story{}, e
	}
	next, e := story.Grant(actor, s.command(c))
	if e != nil {
		return domain.Story{}, ErrNotAvailable
	}
	return s.append(ctx, story, next, c.ID)
}
func (s Service) Publish(ctx context.Context, c Command) (domain.Edition, error) {
	story, _, e := s.current(ctx, c)
	if e != nil {
		return domain.Edition{}, e
	}
	next, _, e := story.Publish(s.now().UTC(), s.command(c))
	if e != nil {
		return domain.Edition{}, ErrNotAvailable
	}
	saved, e := s.append(ctx, story, next, c.ID)
	if e != nil {
		return domain.Edition{}, e
	}
	editions := saved.Editions()
	return editions[len(editions)-1], nil
}

func (s Service) View(ctx context.Context, c Command) (Projection, error) {
	story, actor, e := s.current(ctx, c)
	if e != nil {
		return Projection{}, e
	}
	passages := story.Passages()
	result := Projection{
		ID: story.ID(), TitleCode: story.TitleCode(),
		Passages: make([]PassageView, 0, len(passages)),
		Editions: story.Editions(), Revision: story.Revision(),
	}
	authors := story.Authors()
	result.YourTurn = authors[len(passages)%2] == actor
	for _, grant := range story.Grants() {
		if grant.AuthorKey == actor {
			result.YourGrant = true
		} else {
			result.OtherGrant = true
		}
	}
	result.BothGranted = len(story.Grants()) == 2
	for _, passage := range passages {
		revisions := passage.Revisions
		latest := revisions[len(revisions)-1]
		result.Passages = append(result.Passages, PassageView{
			ID: passage.ID, Ordinal: passage.Ordinal, Content: latest.Content,
			Yours: passage.AuthorKey == actor, CreatedAt: passage.CreatedAt,
			EditedAt: latest.EditedAt,
		})
	}
	return result, nil
}
func (s Service) current(ctx context.Context, c Command) (domain.Story, string, error) {
	if !s.ready() || s.a.RevalidateAuthors(ctx, c.RoomID, c.FirstAuthorID, c.SecondAuthorID) != nil {
		return domain.Story{}, "", ErrNotAvailable
	}
	story, e := s.r.Find(ctx, strings.TrimSpace(c.StoryID))
	if e != nil {
		return domain.Story{}, "", ErrNotAvailable
	}
	room, e := s.k.Key("anansesem:room", strings.TrimSpace(c.RoomID))
	if e != nil || room != story.RoomKey() {
		return domain.Story{}, "", ErrNotAvailable
	}
	authors, e := s.authors(c.FirstAuthorID, c.SecondAuthorID)
	if e != nil || !slices.Equal(authors, story.Authors()) {
		return domain.Story{}, "", ErrNotAvailable
	}
	actor, e := s.k.Key("anansesem:author", strings.TrimSpace(c.ActorID))
	if e != nil || !slices.Contains(authors, actor) {
		return domain.Story{}, "", ErrNotAvailable
	}
	return story, actor, nil
}
func (s Service) authors(a, b string) ([]string, error) {
	x, e := s.k.Key("anansesem:author", strings.TrimSpace(a))
	if e != nil {
		return nil, e
	}
	y, e := s.k.Key("anansesem:author", strings.TrimSpace(b))
	if e != nil {
		return nil, e
	}
	v := []string{x, y}
	slices.Sort(v)
	return v, nil
}
func (s Service) append(ctx context.Context, current, next domain.Story, id string) (domain.Story, error) {
	e := s.r.Append(ctx, next, current.Revision(), id)
	if e == nil {
		return next, nil
	}
	if errors.Is(e, ErrApplied) {
		old, x := s.r.FindByCommand(ctx, id)
		if x == nil {
			return old, nil
		}
	}
	return domain.Story{}, ErrNotAvailable
}
func (s Service) command(c Command) domain.Command {
	return domain.Command{ID: strings.TrimSpace(c.ID), ExpectedRevision: c.ExpectedRevision, At: s.now().UTC()}
}
func (s Service) ready() bool {
	return s.r != nil && s.a != nil && s.k != nil && s.storyIDs != nil && s.passageIDs != nil
}
