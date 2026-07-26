package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/domain"
	"time"
)

type Service struct {
	repository Repository
	keys       Keyer
	now        func() time.Time
}

func NewService(r Repository, k Keyer, n func() time.Time) Service { return Service{r, k, n} }
func (s Service) Block(ctx context.Context, roomID, memberID, commandID string, expected uint64) (domain.Safety, error) {
	return s.apply(ctx, roomID, memberID, commandID, domain.ActionBlock, "", "", expected)
}
func (s Service) Report(ctx context.Context, roomID, memberID, commandID string, category domain.Category, evidenceRef string, expected uint64) (domain.Safety, error) {
	return s.apply(ctx, roomID, memberID, commandID, domain.ActionReport, category, evidenceRef, expected)
}
func (s Service) apply(ctx context.Context, roomID, memberID, commandID string, action domain.Action, category domain.Category, evidence string, expected uint64) (domain.Safety, error) {
	roomKey, err := s.keys.Key("safety_room", roomID)
	if err != nil {
		return domain.Safety{}, err
	}
	actorKey, err := s.keys.Key("safety_actor", memberID)
	if err != nil {
		return domain.Safety{}, err
	}
	current, err := s.repository.Find(ctx, roomKey)
	if err != nil {
		return domain.Safety{}, err
	}
	cmd := domain.Command{ID: commandID, ActorKey: actorKey, Category: category, EvidenceRef: evidence, ExpectedRevision: expected, At: s.now()}
	var changed domain.Safety
	if action == domain.ActionBlock {
		changed, err = current.Block(cmd)
	} else if action == domain.ActionReport {
		changed, err = current.Report(cmd)
	} else {
		return domain.Safety{}, domain.ErrInvalid
	}
	if err != nil {
		return domain.Safety{}, err
	}
	if changed.Revision() != current.Revision() {
		if err = s.repository.Save(ctx, changed, current.Revision(), commandID); err != nil {
			return domain.Safety{}, err
		}
	}
	return changed, nil
}
