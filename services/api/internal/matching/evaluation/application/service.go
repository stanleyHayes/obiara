package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/matching/evaluation/domain"
)

var (
	ErrInvalid  = errors.New("invalid offline evaluation request")
	ErrNotFound = errors.New("offline evaluation not found")
	ErrConflict = errors.New("offline evaluation conflict")
	ErrApplied  = errors.New("offline evaluation command already applied")
)

type Service struct {
	repo      Repository
	snapshots SnapshotVerifier
	slices    SlicePolicy
	authority Authority
	keyer     Keyer
	ids       IDSource
	now       func() time.Time
}

func NewService(r Repository, v SnapshotVerifier, p SlicePolicy, a Authority, k Keyer, ids IDSource, now func() time.Time) Service {
	return Service{repo: r, snapshots: v, slices: p, authority: a, keyer: k, ids: ids, now: now}
}

type CreateCommand struct {
	Actor, Candidate, CommandID string
	CandidateVersion            uint64
}

func (s Service) Create(ctx context.Context, q CreateCommand) (domain.Evaluation, error) {
	if !s.ready() {
		return domain.Evaluation{}, ErrInvalid
	}
	if err := s.authority.RequireEvaluator(ctx, q.Actor); err != nil {
		return domain.Evaluation{}, err
	}
	e, err := domain.Create(s.ids.NewID(), q.Candidate, q.CandidateVersion, domain.Command{ID: q.CommandID, At: s.now().UTC()})
	if err != nil {
		return domain.Evaluation{}, err
	}
	if err = s.repo.Create(ctx, e); errors.Is(err, ErrApplied) {
		return s.repo.FindByCommand(ctx, q.CommandID)
	}
	return e, err
}

type RecordCommand struct {
	Actor, ID, CommandID string
	ExpectedRevision     uint64
	Snapshot             domain.Snapshot
	Metrics              domain.Metrics
}

func (s Service) Record(ctx context.Context, q RecordCommand) (domain.Evaluation, error) {
	if !s.ready() {
		return domain.Evaluation{}, ErrInvalid
	}
	if err := s.authority.RequireEvaluator(ctx, q.Actor); err != nil {
		return domain.Evaluation{}, err
	}
	if err := s.snapshots.Revalidate(ctx, q.Snapshot); err != nil {
		return domain.Evaluation{}, err
	}
	keys := make([]string, 0, len(q.Metrics.Slices))
	for _, m := range q.Metrics.Slices {
		keys = append(keys, m.PolicyKey)
	}
	if err := s.slices.RequireApproved(ctx, keys); err != nil {
		return domain.Evaluation{}, err
	}
	return s.mutate(ctx, q.ID, q.CommandID, q.ExpectedRevision, func(e domain.Evaluation, c domain.Command) (domain.Evaluation, error) {
		return e.Record(q.Snapshot, q.Metrics, c)
	})
}

type CardCommand struct {
	Actor, ID, CommandID string
	ExpectedRevision     uint64
	Card                 domain.ModelCard
}

func (s Service) AttachCard(ctx context.Context, q CardCommand) (domain.Evaluation, error) {
	if !s.ready() {
		return domain.Evaluation{}, ErrInvalid
	}
	if err := s.authority.RequireEvaluator(ctx, q.Actor); err != nil {
		return domain.Evaluation{}, err
	}
	return s.mutate(ctx, q.ID, q.CommandID, q.ExpectedRevision, func(e domain.Evaluation, c domain.Command) (domain.Evaluation, error) { return e.AttachCard(q.Card, c) })
}

type ApproveCommand struct {
	Reviewer, ID, CommandID string
	ExpectedRevision        uint64
	ExpiresAt               time.Time
}

func (s Service) Approve(ctx context.Context, q ApproveCommand) (domain.Evaluation, error) {
	if !s.ready() {
		return domain.Evaluation{}, ErrInvalid
	}
	if err := s.authority.RequireHumanApprover(ctx, q.Reviewer); err != nil {
		return domain.Evaluation{}, err
	}
	reviewer, err := s.keyer.Key("matching-evaluation-reviewer", q.Reviewer)
	if err != nil {
		return domain.Evaluation{}, err
	}
	return s.mutate(ctx, q.ID, q.CommandID, q.ExpectedRevision, func(e domain.Evaluation, c domain.Command) (domain.Evaluation, error) {
		return e.Approve(reviewer, q.ExpiresAt, c)
	})
}
func (s Service) Ready(ctx context.Context, id string) (bool, error) {
	if !s.ready() {
		return false, ErrInvalid
	}
	e, err := s.repo.Find(ctx, id)
	if err != nil {
		return false, err
	}
	if !e.Ready(s.now().UTC()) {
		return false, nil
	}
	if err = s.snapshots.Revalidate(ctx, e.Snapshot()); err != nil {
		return false, nil
	}
	keys := make([]string, 0, len(e.Metrics().Slices))
	for _, m := range e.Metrics().Slices {
		keys = append(keys, m.PolicyKey)
	}
	if err = s.slices.RequireApproved(ctx, keys); err != nil {
		return false, nil
	}
	return true, nil
}
func (s Service) mutate(ctx context.Context, id, command string, revision uint64, fn func(domain.Evaluation, domain.Command) (domain.Evaluation, error)) (domain.Evaluation, error) {
	e, err := s.repo.Find(ctx, id)
	if err != nil {
		return domain.Evaluation{}, err
	}
	next, err := fn(e, domain.Command{ID: command, ExpectedRevision: revision, At: s.now().UTC()})
	if err != nil {
		return domain.Evaluation{}, err
	}
	if err = s.repo.Append(ctx, next, revision, command); errors.Is(err, ErrApplied) {
		return s.repo.FindByCommand(ctx, command)
	}
	return next, err
}
func (s Service) ready() bool {
	return s.repo != nil && s.snapshots != nil && s.slices != nil && s.authority != nil && s.keyer != nil && s.ids != nil && s.now != nil
}
