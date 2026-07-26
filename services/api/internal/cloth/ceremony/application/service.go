package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/ceremony/domain"
	"strings"
	"time"
)

type OpenCommand struct {
	ID, ActorID string
	MemberIDs   [2]string
}
type Command struct {
	ID, CeremonyID, ActorID, DestinationRef string
	ExpectedRevision                        uint64
}
type Result struct {
	Ceremony domain.Ceremony
	Replayed bool
}
type Service struct {
	repository  Repository
	keyer       Keyer
	ids         IDSource
	revalidator PublishRevalidator
	publisher   CirclePublisher
	now         func() time.Time
}

func New(r Repository, k Keyer, ids IDSource, v PublishRevalidator, p CirclePublisher, now func() time.Time) Service {
	return Service{r, k, ids, v, p, now}
}
func (s Service) Open(ctx context.Context, c OpenCommand) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	members, err := s.members(c.MemberIDs)
	if err != nil {
		return Result{}, err
	}
	actor, err := s.key("member", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	id := s.ids.NewID()
	fp := domain.Fingerprint(strings.TrimSpace(c.ID)+"|"+members[0]+"|"+members[1], domain.ActionOpened, actor, "", 0)
	ceremony, err := domain.Open(id, members, domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: actor, Fingerprint: fp, At: s.now()})
	if err != nil {
		return Result{}, err
	}
	if err = s.repository.Create(ctx, ceremony); err == nil {
		return Result{ceremony, false}, nil
	}
	return Result{}, err
}
func (s Service) Confirm(ctx context.Context, c Command) (Result, error) {
	return s.mutate(ctx, c, domain.ActionConfirmed)
}
func (s Service) ProposeAnnouncement(ctx context.Context, c Command) (Result, error) {
	return s.mutate(ctx, c, domain.ActionAnnouncementProposed)
}
func (s Service) ConsentAnnouncement(ctx context.Context, c Command) (Result, error) {
	return s.mutate(ctx, c, domain.ActionAnnouncementConsented)
}
func (s Service) mutate(ctx context.Context, c Command, action domain.Action) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	current, err := s.repository.Find(ctx, strings.TrimSpace(c.CeremonyID))
	if err != nil {
		return Result{}, err
	}
	actor, err := s.key("member", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	destination := ""
	if action == domain.ActionAnnouncementProposed {
		destination, err = s.key("destination", c.DestinationRef)
		if err != nil {
			return Result{}, err
		}
	} else if state := current.State(); state.Announcement != nil {
		destination = state.Announcement.DestinationKey
	}
	fp := domain.Fingerprint(current.ID(), action, actor, destination, c.ExpectedRevision)
	command := domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: actor, Fingerprint: fp, ExpectedRevision: c.ExpectedRevision, At: s.now()}
	var next domain.Ceremony
	switch action {
	case domain.ActionConfirmed:
		next, err = current.Confirm(command)
	case domain.ActionAnnouncementProposed:
		next, err = current.ProposeAnnouncement(destination, command)
	default:
		next, err = current.ConsentAnnouncement(command)
	}
	if err != nil {
		return Result{}, err
	}
	return s.persist(ctx, current, next, c.ID)
}
func (s Service) Publish(ctx context.Context, c Command) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	current, err := s.repository.Find(ctx, strings.TrimSpace(c.CeremonyID))
	if err != nil {
		return Result{}, err
	}
	actor, err := s.key("member", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	state := current.State()
	if !current.AnnouncementReady() || state.Announcement == nil {
		return Result{}, domain.ErrAnnouncementNotReady
	}
	fp := domain.Fingerprint(current.ID(), domain.ActionAnnouncementPublished, actor, state.Announcement.DestinationKey, c.ExpectedRevision)
	next, err := current.PublishAnnouncement(domain.Command{ID: c.ID, ActorKey: actor, Fingerprint: fp, ExpectedRevision: c.ExpectedRevision, At: s.now()})
	if err != nil {
		return Result{}, err
	}
	if err = s.revalidator.Authorize(ctx, state.Members, state.Announcement.DestinationKey); err != nil {
		return Result{}, ErrPublishDenied
	}
	if err = s.publisher.Publish(ctx, PublishRequest{c.ID, state.Announcement.DestinationKey, domain.AnnouncementKind}); err != nil {
		return Result{}, ErrPublishDenied
	}
	return s.persist(ctx, current, next, c.ID)
}
func (s Service) persist(ctx context.Context, current, next domain.Ceremony, commandID string) (Result, error) {
	if next.Revision() == current.Revision() {
		return Result{next, true}, nil
	}
	if err := s.repository.Append(ctx, next, current.Revision(), commandID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			existing, findErr := s.repository.FindByCommand(ctx, commandID)
			if findErr == nil {
				return Result{existing, true}, nil
			}
		}
		return Result{}, err
	}
	return Result{next, false}, nil
}
func (s Service) members(ids [2]string) ([2]string, error) {
	var out [2]string
	for i, id := range ids {
		key, err := s.key("member", id)
		if err != nil {
			return out, err
		}
		out[i] = key
	}
	return out, nil
}
func (s Service) key(namespace, value string) (string, error) {
	key, err := s.keyer.Key(namespace, strings.TrimSpace(value))
	if err != nil {
		return "", ErrUnavailable
	}
	return key, nil
}
func (s Service) ready() bool {
	return s.repository != nil && s.keyer != nil && s.ids != nil && s.revalidator != nil && s.publisher != nil && s.now != nil
}
