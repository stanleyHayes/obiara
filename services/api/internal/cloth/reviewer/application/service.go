package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/cloth/reviewer/domain"
)

var (
	ErrNotFound           = errors.New("reviewer access not found")
	ErrOptimisticConflict = errors.New("reviewer access optimistic conflict")
	ErrCommandApplied     = errors.New("reviewer command applied")
	ErrNotAvailable       = errors.New("reviewer access not available")
)

type Command struct {
	ID, ReviewID, ActorID, FirstMemberID, SecondMemberID string
	ExpectedRevision                                     uint64
}

type CreateRequest struct {
	ReviewerID                  string
	OTP                         string
	QuestionRefs, MaterialRefs  []string
	OTPValidity, InviteValidity time.Duration
}

type RedeemRequest struct {
	InviteToken, OTP string
}

type CreateResult struct {
	Review      domain.Review
	InviteToken string
}

type AccessResult struct {
	Projection domain.Projection
	Session    string
	Replayed   bool
}

type Service struct {
	repository             Repository
	authorizer             Authorizer
	policy                 PairPolicy
	keyer                  Keyer
	inviteTokens, sessions TokenSource
	watermarker            Watermarker
	ids                    IDSource
	now                    func() time.Time
}

func NewService(repository Repository, authorizer Authorizer, policy PairPolicy, keyer Keyer, inviteTokens, sessions TokenSource, watermarker Watermarker, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository, authorizer, policy, keyer, inviteTokens, sessions, watermarker, ids, now}
}

func (service Service) Create(ctx context.Context, command Command, request CreateRequest) (CreateResult, error) {
	if !service.ready() || command.ActorID != command.FirstMemberID {
		return CreateResult{}, ErrNotAvailable
	}
	if err := service.authorizer.Require(ctx, command.ActorID, "cloth.reviewer.create", ""); err != nil {
		return CreateResult{}, ErrNotAvailable
	}
	if err := service.policy.Revalidate(ctx, command.FirstMemberID, command.SecondMemberID); err != nil {
		return CreateResult{}, ErrNotAvailable
	}
	members, err := service.memberKeys(command.FirstMemberID, command.SecondMemberID)
	if err != nil {
		return CreateResult{}, ErrNotAvailable
	}
	reviewer, err := service.keyer.Key("cloth-reviewer:reviewer", strings.TrimSpace(request.ReviewerID))
	if err != nil {
		return CreateResult{}, ErrNotAvailable
	}
	invite, err := service.inviteTokens.Token()
	if err != nil {
		return CreateResult{}, ErrNotAvailable
	}
	inviteDigest, err := service.keyer.Key("cloth-reviewer:invite", invite)
	if err != nil {
		return CreateResult{}, ErrNotAvailable
	}
	otpDigest, err := service.keyer.Key("cloth-reviewer:otp", strings.TrimSpace(request.OTP))
	if err != nil {
		return CreateResult{}, ErrNotAvailable
	}
	id := service.ids.NewID()
	watermark, err := service.watermarker.Ref(reviewer, id)
	if err != nil {
		return CreateResult{}, ErrNotAvailable
	}
	now := service.now().UTC()
	review, err := domain.Create(domain.State{
		ID: id, Members: members, ReviewerKey: reviewer, InviteDigest: inviteDigest, OTPDigest: otpDigest,
		WatermarkRef: watermark, QuestionRefs: request.QuestionRefs, MaterialRefs: request.MaterialRefs,
		OTPExpiresAt: now.Add(request.OTPValidity), InviteExpiresAt: now.Add(request.InviteValidity),
	}, now, service.domainCommand(command))
	if err != nil {
		return CreateResult{}, ErrNotAvailable
	}
	if err = service.repository.Create(ctx, review); err != nil {
		if !errors.Is(err, ErrCommandApplied) {
			return CreateResult{}, ErrNotAvailable
		}
		old, findErr := service.repository.FindByCommand(ctx, command.ID)
		if findErr != nil || old.ReviewerKey() != reviewer {
			return CreateResult{}, ErrNotAvailable
		}
		return CreateResult{Review: old}, nil
	}
	return CreateResult{Review: review, InviteToken: invite}, nil
}

func (service Service) Redeem(ctx context.Context, command Command, request RedeemRequest) (AccessResult, error) {
	review, reviewer, err := service.access(ctx, command, "cloth.reviewer.redeem")
	if err != nil {
		return AccessResult{}, err
	}
	inviteDigest, err := service.keyer.Key("cloth-reviewer:invite", strings.TrimSpace(request.InviteToken))
	if err != nil {
		return AccessResult{}, ErrNotAvailable
	}
	otpDigest, err := service.keyer.Key("cloth-reviewer:otp", strings.TrimSpace(request.OTP))
	if err != nil {
		return AccessResult{}, ErrNotAvailable
	}
	if review.HasCommand(command.ID) {
		if inviteDigest != review.InviteDigest() || otpDigest != review.OTPDigest() {
			return AccessResult{}, ErrNotAvailable
		}
		projection, projectErr := review.Project(review.SessionDigest(), service.now().UTC())
		if projectErr != nil {
			return AccessResult{}, ErrNotAvailable
		}
		return AccessResult{Projection: projection, Replayed: true}, nil
	}
	session, err := service.sessions.Token()
	if err != nil {
		return AccessResult{}, ErrNotAvailable
	}
	sessionDigest, err := service.keyer.Key("cloth-reviewer:session", session)
	if err != nil {
		return AccessResult{}, ErrNotAvailable
	}
	next, err := review.Redeem(inviteDigest, otpDigest, sessionDigest, service.now().UTC(), service.domainCommand(command))
	if err != nil {
		return AccessResult{}, ErrNotAvailable
	}
	if err = service.repository.Append(ctx, next, review.Revision(), command.ID); err != nil {
		if errors.Is(err, ErrCommandApplied) {
			old, findErr := service.repository.FindByCommand(ctx, command.ID)
			if findErr == nil && old.IsReviewer(reviewer) {
				projection, projectErr := old.Project(old.SessionDigest(), service.now().UTC())
				if projectErr == nil {
					return AccessResult{Projection: projection, Replayed: true}, nil
				}
			}
		}
		return AccessResult{}, ErrNotAvailable
	}
	projection, err := next.Project(sessionDigest, service.now().UTC())
	if err != nil {
		return AccessResult{}, ErrNotAvailable
	}
	return AccessResult{Projection: projection, Session: session}, nil
}

func (service Service) View(ctx context.Context, command Command, session string) (AccessResult, error) {
	review, _, err := service.access(ctx, command, "cloth.reviewer.view")
	if err != nil {
		return AccessResult{}, err
	}
	sessionDigest, err := service.keyer.Key("cloth-reviewer:session", strings.TrimSpace(session))
	if err != nil {
		return AccessResult{}, ErrNotAvailable
	}
	projection, err := review.Project(sessionDigest, service.now().UTC())
	if err != nil {
		return AccessResult{}, ErrNotAvailable
	}
	return AccessResult{Projection: projection}, nil
}

func (service Service) Revoke(ctx context.Context, command Command) error {
	review, _, err := service.access(ctx, command, "cloth.reviewer.revoke")
	if err != nil {
		return err
	}
	actor, err := service.keyer.Key("cloth-reviewer:member", strings.TrimSpace(command.ActorID))
	if err != nil || !review.HasMember(actor) {
		return ErrNotAvailable
	}
	next, err := review.Revoke(service.now().UTC(), service.domainCommand(command))
	if err != nil {
		return ErrNotAvailable
	}
	if err = service.repository.Append(ctx, next, review.Revision(), command.ID); err != nil {
		return ErrNotAvailable
	}
	return nil
}

func (service Service) access(ctx context.Context, command Command, permission string) (domain.Review, string, error) {
	if !service.ready() {
		return domain.Review{}, "", ErrNotAvailable
	}
	if err := service.authorizer.Require(ctx, command.ActorID, permission, command.ReviewID); err != nil {
		return domain.Review{}, "", ErrNotAvailable
	}
	if err := service.policy.Revalidate(ctx, command.FirstMemberID, command.SecondMemberID); err != nil {
		return domain.Review{}, "", ErrNotAvailable
	}
	review, err := service.repository.Find(ctx, strings.TrimSpace(command.ReviewID))
	if err != nil {
		return domain.Review{}, "", ErrNotAvailable
	}
	members, err := service.memberKeys(command.FirstMemberID, command.SecondMemberID)
	if err != nil || !slices.Equal(members, review.Members()) {
		return domain.Review{}, "", ErrNotAvailable
	}
	reviewer, err := service.keyer.Key("cloth-reviewer:reviewer", strings.TrimSpace(command.ActorID))
	if err != nil {
		return domain.Review{}, "", ErrNotAvailable
	}
	if permission != "cloth.reviewer.revoke" && !review.IsReviewer(reviewer) {
		return domain.Review{}, "", ErrNotAvailable
	}
	return review, reviewer, nil
}

func (service Service) memberKeys(first, second string) ([]string, error) {
	a, err := service.keyer.Key("cloth-reviewer:member", strings.TrimSpace(first))
	if err != nil {
		return nil, err
	}
	b, err := service.keyer.Key("cloth-reviewer:member", strings.TrimSpace(second))
	if err != nil {
		return nil, err
	}
	members := []string{a, b}
	slices.Sort(members)
	return members, nil
}

func (service Service) domainCommand(command Command) domain.Command {
	return domain.Command{ID: strings.TrimSpace(command.ID), ExpectedRevision: command.ExpectedRevision, At: service.now().UTC()}
}

func (service Service) ready() bool {
	return service.repository != nil && service.authorizer != nil && service.policy != nil && service.keyer != nil &&
		service.inviteTokens != nil && service.sessions != nil && service.watermarker != nil && service.ids != nil
}
