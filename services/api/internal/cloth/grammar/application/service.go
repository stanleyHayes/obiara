package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/grammar/domain"
	"strings"
)

type Command struct {
	ID, Version               string
	MemberRefs                [2]string
	ThemeRefs, ProvenanceRefs []string
}
type Result struct {
	Recipe   domain.Recipe
	Replayed bool
}
type Service struct {
	repository Repository
	keyer      Keyer
}

func New(repository Repository, keyer Keyer) Service { return Service{repository, keyer} }
func (s Service) Compile(ctx context.Context, c Command) (Result, error) {
	if s.repository == nil || s.keyer == nil {
		return Result{}, ErrUnavailable
	}
	pair := [2]string{}
	for i, ref := range c.MemberRefs {
		key, err := s.keyer.Key("pair", strings.TrimSpace(ref))
		if err != nil {
			return Result{}, ErrUnavailable
		}
		pair[i] = key
	}
	themes, err := s.keys("theme", c.ThemeRefs)
	if err != nil {
		return Result{}, err
	}
	provenance, err := s.keys("provenance", c.ProvenanceRefs)
	if err != nil {
		return Result{}, err
	}
	recipe, err := domain.Compile(domain.Input{Version: strings.TrimSpace(c.Version), PairKeys: pair, ThemeKeys: themes, ProvenanceKeys: provenance}, strings.TrimSpace(c.ID))
	if err != nil {
		return Result{}, err
	}
	stored, replayed, err := s.repository.Store(ctx, recipe, 0)
	if err != nil {
		return Result{}, err
	}
	return Result{stored, replayed}, nil
}
func (s Service) keys(namespace string, refs []string) ([]string, error) {
	keys := make([]string, 0, len(refs))
	for _, ref := range refs {
		key, err := s.keyer.Key(namespace, strings.TrimSpace(ref))
		if err != nil {
			return nil, ErrUnavailable
		}
		keys = append(keys, key)
	}
	return keys, nil
}
