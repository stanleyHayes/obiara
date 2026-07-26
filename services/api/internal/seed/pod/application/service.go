package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
	"strings"
	"time"
)

var (
	ErrNotFound           = errors.New("seed pod not found")
	ErrOptimisticConflict = errors.New("seed pod optimistic conflict")
	ErrCommandApplied     = errors.New("seed pod command applied")
	ErrUnavailable        = errors.New("seed pod unavailable")
	ErrNotAvailable       = errors.New("seed pod not available")
)

type Command struct {
	ID, PodID, ActorID, ReasonCode string
	ExpectedRevision               uint64
}
type Proposal struct {
	OwnerID, MediaRef string
	RecipientIDs      []string
	TTL               time.Duration
}
type Result struct {
	Pod           domain.Pod
	PlaybackToken string
	Replayed      bool
}
type Service struct {
	r   Repository
	a   Authorizer
	e   PlaybackEligibility
	i   MediaIssuer
	k   Keyer
	ids IDSource
	now func() time.Time
}

func NewService(r Repository, a Authorizer, e PlaybackEligibility, i MediaIssuer, k Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{r, a, e, i, k, ids, now}
}
func (s Service) Create(ctx context.Context, c Command, p Proposal) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	if c.ActorID != p.OwnerID || len(p.RecipientIDs) > domain.MaxRecipients {
		return Result{}, ErrNotAvailable
	}
	if err := s.a.Require(ctx, c.ActorID, "seed.pod.create", ""); err != nil {
		return Result{}, ErrNotAvailable
	}
	owner, err := s.key("seed-pod:member", p.OwnerID)
	if err != nil {
		return Result{}, err
	}
	media, err := s.key("seed-pod:media", p.MediaRef)
	if err != nil {
		return Result{}, err
	}
	recipients := make([]string, 0, len(p.RecipientIDs))
	for _, id := range p.RecipientIDs {
		x, e := s.key("seed-pod:member", id)
		if e != nil {
			return Result{}, e
		}
		recipients = append(recipients, x)
	}
	pod, err := domain.Create(s.ids.NewID(), owner, media, recipients, s.now().Add(p.TTL), s.command(c, owner))
	if err != nil {
		return Result{}, err
	}
	if err = s.r.Create(ctx, pod); err != nil {
		if !errors.Is(err, ErrCommandApplied) {
			return Result{}, s.translate(err)
		}
		old, e := s.r.FindByCommand(ctx, c.ID)
		if e != nil {
			return Result{}, domain.ErrCommandMismatch
		}
		return Result{Pod: old, Replayed: true}, nil
	}
	return Result{Pod: pod}, nil
}
func (s Service) Playback(ctx context.Context, c Command) (Result, error) {
	if !s.ready() {
		return Result{}, ErrUnavailable
	}
	p, err := s.r.Find(ctx, strings.TrimSpace(c.PodID))
	if err != nil {
		return Result{}, s.translate(err)
	}
	if err = s.a.Require(ctx, c.ActorID, "seed.pod.playback", p.ID()); err != nil {
		return Result{}, ErrNotAvailable
	}
	// Revalidation deliberately occurs before replay handling: an old command can
	// never bypass a later consent withdrawal, revocation, or eligibility expiry.
	if err = s.e.Revalidate(ctx, c.ActorID, p.ID()); err != nil {
		return Result{}, ErrNotAvailable
	}
	actor, err := s.key("seed-pod:member", c.ActorID)
	if err != nil {
		return Result{}, err
	}
	replay := p.HasCommand(c.ID)
	next, err := p.Play(s.command(c, actor))
	if err != nil {
		return Result{}, ErrNotAvailable
	}
	if !replay {
		if err = s.r.Append(ctx, next, p.Revision(), c.ID); err != nil {
			if !errors.Is(err, ErrCommandApplied) {
				return Result{}, s.translate(err)
			}
			next, err = s.r.FindByCommand(ctx, c.ID)
			if err != nil {
				return Result{}, domain.ErrCommandMismatch
			}
			replay = true
		}
	}
	token, err := s.i.Issue(ctx, next.MediaKey(), c.ID, 5*time.Minute)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	return Result{Pod: next, PlaybackToken: token, Replayed: replay}, nil
}
func (s Service) ready() bool {
	return s.r != nil && s.a != nil && s.e != nil && s.i != nil && s.k != nil && s.ids != nil
}
func (s Service) key(n, v string) (string, error) {
	x, e := s.k.Key(n, strings.TrimSpace(v))
	if e != nil {
		return "", ErrUnavailable
	}
	return x, nil
}
func (s Service) command(c Command, a string) domain.Command {
	return domain.Command{ID: strings.TrimSpace(c.ID), ActorKey: a, ReasonCode: strings.TrimSpace(c.ReasonCode), ExpectedRevision: c.ExpectedRevision, At: s.now().UTC()}
}
func (s Service) translate(e error) error {
	if errors.Is(e, ErrNotFound) || errors.Is(e, ErrCommandApplied) {
		return e
	}
	if errors.Is(e, ErrOptimisticConflict) {
		return domain.ErrStaleRevision
	}
	return ErrUnavailable
}
