package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
)

type Command struct {
	ID      string
	ActorID string
	// TargetID is who the sow is toward.
	TargetID  string
	Body      string
	MediaRefs []string
	Confirmed bool
}
type Result struct {
	Sow      domain.Sow
	Replayed bool
}

type Service struct {
	screening  Screening
	acceptance Acceptance
	keyer      Keyer
	media      MediaOwnership
	listen     ListenGate
	blocks     BlockList
	declines   DeclineLock
	ids        IDSource
	now        func() time.Time
	units      int64
}

func New(screening Screening, acceptance Acceptance, keyer Keyer, ids IDSource, now func() time.Time, units int64) Service {
	return Service{screening: screening, acceptance: acceptance, keyer: keyer, ids: ids, now: now, units: units}
}

// WithMediaOwnership attaches the check that a sow carries only the sower's
// own recordings.
func (s Service) WithMediaOwnership(media MediaOwnership) Service {
	s.media = media
	return s
}

// WithReachRules attaches the three questions every reach toward a person has
// to answer: is either of you blocked, did they already decline you, and have
// you actually heard them.
func (s Service) WithReachRules(listen ListenGate, blocks BlockList, declines DeclineLock) Service {
	s.listen, s.blocks, s.declines = listen, blocks, declines
	return s
}

func (s Service) Send(ctx context.Context, command Command) (Result, error) {
	if s.screening == nil || s.acceptance == nil || s.keyer == nil || s.ids == nil || s.now == nil || s.units <= 0 {
		return Result{}, ErrUnavailable
	}
	if !command.Confirmed {
		return Result{}, domain.ErrNotConfirmed
	}
	body := strings.TrimSpace(command.Body)
	if body == "" || strings.TrimSpace(command.ID) == "" || strings.TrimSpace(command.ActorID) == "" ||
		strings.TrimSpace(command.TargetID) == "" || len(command.MediaRefs) > 4 {
		return Result{}, domain.ErrInvalid
	}
	if err := s.mayReach(ctx, command.ActorID, command.TargetID); err != nil {
		return Result{}, err
	}
	// Checked before screening: a member must not be able to have somebody
	// else's recording screened, and a sow carrying a voice that is not
	// theirs should never reach a reviewer looking like theirs.
	if len(command.MediaRefs) > 0 {
		if s.media == nil {
			return Result{}, ErrUnavailable
		}
		owned, ownErr := s.media.OwnedBy(ctx, command.ActorID, command.MediaRefs)
		if ownErr != nil {
			return Result{}, ErrUnavailable
		}
		if !owned {
			return Result{}, ErrMediaNotOwned
		}
	}

	// Three outcomes, not two. Screening can clear a sow, refuse it, or send
	// it to a person — and the third used to arrive here as an error and be
	// reported to the member as "service unavailable", which is neither true
	// nor something they could act on.
	status := domain.StatusDelivered
	decision, err := s.screening.Screen(ctx, body, append([]string(nil), command.MediaRefs...))
	switch {
	case errors.Is(err, ErrHumanReviewRequired):
		if strings.TrimSpace(decision.Reference) == "" {
			return Result{}, ErrUnavailable
		}
		status = domain.StatusPendingReview
	case err != nil:
		return Result{}, ErrUnavailable
	case !decision.Approved:
		return Result{}, domain.ErrScreeningRejected
	}
	actorKey, err := s.keyer.Key("allowance-subject", command.ActorID)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	targetKey, err := s.keyer.Key("participant", command.TargetID)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	screeningKey, err := s.keyer.Key("screening", decision.Reference)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	media := make([]domain.Media, 0, len(command.MediaRefs))
	for _, ref := range command.MediaRefs {
		key, keyErr := s.keyer.Key("media", ref)
		if keyErr != nil {
			return Result{}, ErrUnavailable
		}
		media = append(media, domain.Media{Key: key, ScreeningKey: screeningKey})
	}
	fp := fingerprint(command.ID, actorKey, targetKey, body, command.MediaRefs, s.units)
	candidate, err := domain.Accept(s.ids.NewID(), actorKey, targetKey, body, media, command.ID, fp, s.units,
		status, decision.Reference, s.now())
	if err != nil {
		return Result{}, err
	}
	accepted, replayed, err := s.acceptance.Accept(ctx, candidate)
	if err != nil {
		return Result{}, err
	}
	return Result{Sow: accepted, Replayed: replayed}, nil
}

// fingerprint binds a command id to what it asked for. The target is part of
// it: without that, retrying one command id toward a different person would
// look like a replay and be answered with the first sow.
func fingerprint(commandID, actorKey, targetKey, body string, media []string, units int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%q|%q|%q|%q|%q|%d", commandID, actorKey, targetKey, body, media, units)))
	return hex.EncodeToString(sum[:])
}

// Review settles a held sow after a person decided it.
//
// The decision reference replaces the review reference on the sow, so what
// the aggregate carries afterwards is the judgement that settled it rather
// than the queue entry that asked for one.
func (s Service) Review(ctx context.Context, screeningRef string, approve bool, decisionRef string) (Result, error) {
	if s.acceptance == nil || s.now == nil {
		return Result{}, ErrUnavailable
	}
	screeningRef = strings.TrimSpace(screeningRef)
	if screeningRef == "" || strings.TrimSpace(decisionRef) == "" {
		return Result{}, domain.ErrInvalid
	}
	held, err := s.acceptance.FindByScreening(ctx, screeningRef)
	if err != nil {
		return Result{}, err
	}
	decided := held
	if approve {
		decided, err = held.Release(decisionRef, s.now())
	} else {
		decided, err = held.Refuse(decisionRef, s.now())
	}
	if err != nil {
		// Includes ErrNotPending, which is what a second decision on the
		// same sow gets. Deciding twice would refund a seed twice.
		return Result{}, err
	}
	if err := s.acceptance.Settle(ctx, decided, !approve); err != nil {
		return Result{}, err
	}
	return Result{Sow: decided}, nil
}

// mayReach asks the three questions in the same order the sprout path asks
// them, and for the same reasons: a block is the strongest thing either
// member can have said about the other, being told no should cost nothing,
// and none of it should happen after the seed is spent.
//
// A missing check refuses. A sow that skipped one is exactly the outcome each
// of them exists to prevent.
func (s Service) mayReach(ctx context.Context, actorID, targetID string) error {
	if s.blocks == nil || s.listen == nil || s.declines == nil {
		return ErrUnavailable
	}
	blocked, err := s.blocks.Blocked(ctx, actorID, targetID)
	if err != nil {
		return ErrUnavailable
	}
	if blocked {
		return ErrReachNotAvailable
	}
	heard, err := s.listen.Heard(ctx, actorID, targetID)
	if err != nil {
		return ErrUnavailable
	}
	if !heard {
		return ErrNotHeard
	}
	locked, err := s.declines.Locked(ctx, actorID, targetID)
	if err != nil {
		return ErrUnavailable
	}
	if locked {
		return ErrReachNotAvailable
	}
	return nil
}
