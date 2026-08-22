package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
)

type Clock func() time.Time

type Service struct {
	repository      Repository
	keyer           SubjectKeyer
	week            domain.WeekPolicy
	clock           Clock
	weeklyAllowance int64
}

func NewService(repository Repository, keyer SubjectKeyer, week domain.WeekPolicy, clock Clock, weeklyAllowance int64) (*Service, error) {
	if repository == nil || keyer == nil || clock == nil || weeklyAllowance <= 0 {
		return nil, domain.ErrInvalidLedger
	}
	return &Service{repository: repository, keyer: keyer, week: week, clock: clock, weeklyAllowance: weeklyAllowance}, nil
}

func (s *Service) Issue(ctx context.Context, subjectID, commandID string) (domain.Ledger, error) {
	key, err := s.keyer.Key(subjectID)
	if err != nil {
		return domain.Ledger{}, err
	}
	now := s.clock()
	ledger, err := domain.Issue(key, s.weeklyAllowance, s.week.Start(now), now, commandID, fingerprint("issuance", key, commandID, strconv.FormatInt(s.weeklyAllowance, 10)))
	if err != nil {
		return domain.Ledger{}, err
	}
	if err = s.repository.Create(ctx, ledger); err != nil {
		return domain.Ledger{}, err
	}
	return ledger, nil
}

func (s *Service) Spend(ctx context.Context, subjectID, commandID string, units int64) (domain.Ledger, error) {
	key, err := s.keyer.Key(subjectID)
	if err != nil {
		return domain.Ledger{}, err
	}
	current, err := s.repository.Find(ctx, key)
	if err != nil {
		return domain.Ledger{}, err
	}
	now := s.clock()
	fp := fingerprint("spend", key, commandID, strconv.FormatInt(units, 10))
	next, changed, err := current.Spend(units, s.week.Start(now), now, commandID, fp)
	if err != nil || !changed {
		return next, err
	}
	if err = s.repository.Save(ctx, next, current.Version()); err != nil {
		return domain.Ledger{}, err
	}
	return next, nil
}

func (s *Service) Renew(ctx context.Context, subjectID, commandID string) (domain.Ledger, error) {
	key, err := s.keyer.Key(subjectID)
	if err != nil {
		return domain.Ledger{}, err
	}
	current, err := s.repository.Find(ctx, key)
	if err != nil {
		return domain.Ledger{}, err
	}
	now := s.clock()
	start := s.week.Start(now)
	next, changed, err := current.Renew(start, now, commandID, fingerprint("renewal", key, commandID, start.Format(time.RFC3339)))
	if err != nil || !changed {
		return next, err
	}
	if err = s.repository.Save(ctx, next, current.Version()); err != nil {
		return domain.Ledger{}, err
	}
	return next, nil
}

func fingerprint(parts ...string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%q", parts)))
	return hex.EncodeToString(sum[:])
}

// CurrentOrIssue returns a member's ledger, issuing this week's allowance
// the first time they ask for it.
//
// The ledger is created lazily rather than at sign-up because most accounts
// never reach the sowing stage, and a row per account per week for people
// who never sow is cost without a reader. Issuing grants exactly the
// configured weekly allowance and never more, so a first read can only ever
// produce the same balance a scheduled issuance would have.
func (s *Service) CurrentOrIssue(ctx context.Context, subjectID, commandID string) (domain.Ledger, error) {
	key, err := s.keyer.Key(subjectID)
	if err != nil {
		return domain.Ledger{}, err
	}
	ledger, err := s.repository.Find(ctx, key)
	if err == nil {
		return ledger, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return domain.Ledger{}, err
	}
	return s.Issue(ctx, subjectID, commandID)
}
