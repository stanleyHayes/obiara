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
	ID        string
	ActorID   string
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
	ids        IDSource
	now        func() time.Time
	units      int64
}

func New(screening Screening, acceptance Acceptance, keyer Keyer, ids IDSource, now func() time.Time, units int64) Service {
	return Service{screening, acceptance, keyer, ids, now, units}
}

func (s Service) Send(ctx context.Context, command Command) (Result, error) {
	if s.screening == nil || s.acceptance == nil || s.keyer == nil || s.ids == nil || s.now == nil || s.units <= 0 {
		return Result{}, ErrUnavailable
	}
	if !command.Confirmed {
		return Result{}, domain.ErrNotConfirmed
	}
	body := strings.TrimSpace(command.Body)
	if body == "" || strings.TrimSpace(command.ID) == "" || strings.TrimSpace(command.ActorID) == "" || len(command.MediaRefs) > 4 {
		return Result{}, domain.ErrInvalid
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
	fp := fingerprint(command.ID, actorKey, body, command.MediaRefs, s.units)
	candidate, err := domain.Accept(s.ids.NewID(), actorKey, body, media, command.ID, fp, s.units,
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

func fingerprint(commandID, actorKey, body string, media []string, units int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%q|%q|%q|%q|%d", commandID, actorKey, body, media, units)))
	return hex.EncodeToString(sum[:])
}
