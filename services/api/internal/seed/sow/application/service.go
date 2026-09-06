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
	delivery   Delivery
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

// WithDelivery attaches the step that places a delivered sow at the
// recipient's house front. Without it a sow is accepted, charged and marked
// delivered while nobody receives anything, so a service composed without it
// refuses instead.
func (s Service) WithDelivery(delivery Delivery) Service {
	s.delivery = delivery
	return s
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
		strings.TrimSpace(command.TargetID) == "" ||
		len(command.MediaRefs) == 0 || len(command.MediaRefs) > 4 {
		return Result{}, domain.ErrInvalid
	}
	if err := s.mayReach(ctx, command.ActorID, command.TargetID); err != nil {
		return Result{}, err
	}
	// Checked before screening: a member must not be able to have somebody
	// else's recording screened, and a sow carrying a voice that is not
	// theirs should never reach a reviewer looking like theirs.
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
		status, decision.Reference,
		domain.Delivery{
			SowerID: command.ActorID, TargetID: command.TargetID, MediaRefs: command.MediaRefs,
		},
		s.now())
	if err != nil {
		return Result{}, err
	}
	accepted, replayed, err := s.acceptance.Accept(ctx, candidate)
	if err != nil {
		return Result{}, err
	}
	// A sow screening cleared outright is delivered now. One held for a
	// person is delivered when they release it, in Review.
	if !replayed && accepted.Status == domain.StatusDelivered {
		if err := s.deliver(ctx, accepted); err != nil {
			return Result{}, err
		}
	}
	return Result{Sow: accepted, Replayed: replayed}, nil
}

// deliver places the sow at the recipient's house front.
//
// Without this a released sow was marked delivered and went nowhere: nothing
// read StatusDelivered, nothing created a pod, and the recipient never learned
// anything had been sent while the sender's seed stayed spent
// (agent_plan.md §64).
//
// A missing delivery port refuses rather than silently marking a sow
// delivered that will never arrive.
func (s Service) deliver(ctx context.Context, sow domain.Sow) error {
	if s.delivery == nil {
		return ErrUnavailable
	}
	if err := s.delivery.Place(ctx, Deliverable{
		SowID:     sow.ID,
		SowerID:   sow.Delivery.SowerID,
		TargetID:  sow.Delivery.TargetID,
		MediaRefs: append([]string(nil), sow.Delivery.MediaRefs...),
	}); err != nil {
		return ErrNotDelivered
	}
	return nil
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
	// Settled first, delivered second, in that order for the same reason the
	// review desk settles before it records: a failure between them leaves a
	// sow that a reviewer released and nobody received, which a retry can
	// finish. The other order would place a pod for a sow whose release was
	// never written down.
	if decided.Status == domain.StatusDelivered {
		if err := s.deliver(ctx, decided); err != nil {
			return Result{Sow: decided}, err
		}
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
