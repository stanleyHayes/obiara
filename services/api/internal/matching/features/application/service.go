package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/matching/features/domain"
)

var (
	ErrInvalid      = errors.New("invalid feature request")
	ErrNotFound     = errors.New("feature grant not found")
	ErrConflict     = errors.New("feature grant conflict")
	ErrApplied      = errors.New("feature command already applied")
	ErrNotEffective = errors.New("feature consent is not effective")
)

type Service struct {
	catalog   Catalog
	grants    GrantRepository
	decisions DecisionRepository
	authority Authority
	keyer     Keyer
	ids       IDSource
	now       func() time.Time
}

func NewService(c Catalog, g GrantRepository, d DecisionRepository, a Authority, k Keyer, ids IDSource, now func() time.Time) Service {
	return Service{catalog: c, grants: g, decisions: d, authority: a, keyer: k, ids: ids, now: now}
}

type GrantCommand struct {
	Actor, Member, Feature, Purpose, CommandID     string
	FeatureVersion, GrantVersion, ExpectedRevision uint64
}

func (s Service) Grant(ctx context.Context, q GrantCommand) (domain.Grant, error) {
	if err := s.ready(); err != nil || strings.TrimSpace(q.Purpose) == "" {
		return domain.Grant{}, ErrInvalid
	}
	if err := s.authority.RequireMember(ctx, q.Actor, q.Member); err != nil {
		return domain.Grant{}, err
	}
	d, err := s.catalog.FindDefinition(ctx, q.Feature, q.FeatureVersion)
	if err != nil {
		return domain.Grant{}, err
	}
	if d.Purpose != q.Purpose {
		return domain.Grant{}, ErrInvalid
	}
	member, err := s.keyer.Key("matching-feature-member", q.Member)
	if err != nil {
		return domain.Grant{}, err
	}
	change := domain.Command{ID: q.CommandID, ExpectedRevision: q.ExpectedRevision, At: s.now().UTC()}
	var g domain.Grant
	if q.ExpectedRevision == 0 {
		g, err = domain.GrantFeature(member, d, q.GrantVersion, change)
	} else {
		current, findErr := s.grants.Find(ctx, member, q.Feature)
		if findErr != nil {
			return domain.Grant{}, findErr
		}
		g, err = current.Regrant(d, q.GrantVersion, change)
	}
	if err != nil {
		return domain.Grant{}, err
	}
	if q.ExpectedRevision == 0 {
		err = s.grants.Create(ctx, g)
	} else {
		err = s.grants.Append(ctx, g, q.ExpectedRevision, q.CommandID)
	}
	if errors.Is(err, ErrApplied) {
		return s.grants.FindByCommand(ctx, q.CommandID)
	}
	return g, err
}

type WithdrawCommand struct {
	Actor, Member, Feature, CommandID string
	ExpectedRevision                  uint64
}

func (s Service) Withdraw(ctx context.Context, q WithdrawCommand) (domain.Grant, error) {
	if err := s.ready(); err != nil {
		return domain.Grant{}, ErrInvalid
	}
	if err := s.authority.RequireMember(ctx, q.Actor, q.Member); err != nil {
		return domain.Grant{}, err
	}
	member, err := s.keyer.Key("matching-feature-member", q.Member)
	if err != nil {
		return domain.Grant{}, err
	}
	g, err := s.grants.Find(ctx, member, q.Feature)
	if err != nil {
		return domain.Grant{}, err
	}
	next, err := g.Withdraw(domain.Command{ID: q.CommandID, ExpectedRevision: q.ExpectedRevision, At: s.now().UTC()})
	if err != nil {
		return domain.Grant{}, err
	}
	if err = s.grants.Append(ctx, next, q.ExpectedRevision, q.CommandID); errors.Is(err, ErrApplied) {
		return s.grants.FindByCommand(ctx, q.CommandID)
	}
	return next, err
}

type PairRequest struct{ Actor, First, Second string }

func (s Service) Decide(ctx context.Context, q PairRequest) (domain.Decision, error) {
	if err := s.ready(); err != nil || q.First == q.Second {
		return domain.Decision{}, ErrInvalid
	}
	if err := s.authority.RequirePair(ctx, q.Actor, q.First, q.Second); err != nil {
		return domain.Decision{}, err
	}
	first, err := s.keyer.Key("matching-feature-member", q.First)
	if err != nil {
		return domain.Decision{}, err
	}
	second, err := s.keyer.Key("matching-feature-member", q.Second)
	if err != nil {
		return domain.Decision{}, err
	}
	at := s.now().UTC()
	a, err := s.effective(ctx, first, at)
	if err != nil {
		return domain.Decision{}, err
	}
	b, err := s.effective(ctx, second, at)
	if err != nil {
		return domain.Decision{}, err
	}
	var enabled []domain.EnabledFeature
	for key, ga := range a {
		gb, ok := b[key]
		if !ok || ga.Purpose() != gb.Purpose() || ga.FeatureVersion() != gb.FeatureVersion() {
			continue
		}
		enabled = append(enabled, domain.EnabledFeature{Key: key, FeatureVersion: ga.FeatureVersion(), Purpose: ga.Purpose(), Consents: []domain.ConsentRef{{MemberKey: first, GrantVersion: ga.GrantVersion()}, {MemberKey: second, GrantVersion: gb.GrantVersion()}}})
	}
	d, err := domain.NewDecision(s.ids.NewID(), first, second, enabled, at)
	if err != nil {
		return domain.Decision{}, err
	}
	return d, s.decisions.CreateDecision(ctx, d)
}
func (s Service) Revalidate(ctx context.Context, id string) (bool, error) {
	if err := s.ready(); err != nil {
		return false, ErrInvalid
	}
	d, err := s.decisions.FindDecision(ctx, id)
	if err != nil {
		return false, err
	}
	at := s.now().UTC()
	for _, f := range d.Features {
		def, e := s.catalog.Current(ctx, f.Key)
		if e != nil || def.Version != f.FeatureVersion || def.Purpose != f.Purpose || !def.Active(at) {
			return false, nil
		}
		for _, ref := range f.Consents {
			g, e := s.grants.Find(ctx, ref.MemberKey, f.Key)
			if e != nil || g.GrantVersion() != ref.GrantVersion || !g.Effective(def, at) {
				return false, nil
			}
		}
	}
	return true, nil
}
func (s Service) effective(ctx context.Context, member string, at time.Time) (map[string]domain.Grant, error) {
	gs, err := s.grants.ListEffective(ctx, member)
	if err != nil {
		return nil, err
	}
	out := make(map[string]domain.Grant)
	for _, g := range gs {
		d, e := s.catalog.Current(ctx, g.FeatureKey())
		if e == nil && g.Effective(d, at) {
			out[g.FeatureKey()] = g
		}
	}
	return out, nil
}
func (s Service) ready() error {
	if s.catalog == nil || s.grants == nil || s.decisions == nil || s.authority == nil || s.keyer == nil || s.ids == nil || s.now == nil {
		return ErrInvalid
	}
	return nil
}
