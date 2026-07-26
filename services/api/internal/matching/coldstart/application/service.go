// Package application coordinates the E11-S02 cold-start rule boundary.
package application

import (
	"context"
	"errors"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/matching/coldstart/domain"
)

var (
	ErrUnavailable    = errors.New("cold-start introductions unavailable")
	ErrInvalidRequest = errors.New("invalid cold-start introduction request")
)

type Request struct {
	RequesterKey string
	Limit        int
}

type Explanation struct {
	CandidateKey string
	Reasons      []domain.ReasonCode
}

type Service struct {
	authority   Authority
	preferences Preferences
	trust       TrustPaths
	visibility  Visibility
}

func New(authority Authority, preferences Preferences, trust TrustPaths, visibility Visibility) Service {
	return Service{authority: authority, preferences: preferences, trust: trust, visibility: visibility}
}

func (service Service) Generate(ctx context.Context, request Request) ([]Explanation, error) {
	request.RequesterKey = strings.TrimSpace(request.RequesterKey)
	if service.authority == nil || service.preferences == nil || service.trust == nil || service.visibility == nil ||
		!domainOpaqueKey(request.RequesterKey) || request.Limit < 1 || request.Limit > domain.MaxCandidates {
		return nil, ErrInvalidRequest
	}
	if err := service.authority.AuthorizeColdStart(ctx, request.RequesterKey); err != nil {
		return nil, ErrUnavailable
	}
	preferences, err := service.preferences.Reciprocal(ctx, request.RequesterKey, domain.MaxInputCandidates)
	if err != nil {
		return nil, ErrUnavailable
	}
	summaries, err := service.trust.Summaries(ctx, request.RequesterKey, domain.MaxInputCandidates)
	if err != nil {
		return nil, ErrUnavailable
	}
	projected, err := domain.Project(request.RequesterKey, preferences, summaries, domain.MaxCandidates)
	if err != nil {
		return nil, ErrUnavailable
	}

	explanations := make([]Explanation, 0, len(projected))
	for _, candidate := range projected {
		visible, visibilityErr := service.visibility.CanIntroduce(ctx, request.RequesterKey, candidate.CandidateKey)
		if visibilityErr != nil {
			return nil, ErrUnavailable
		}
		if !visible {
			continue
		}
		explanations = append(explanations, Explanation{
			CandidateKey: candidate.CandidateKey,
			Reasons:      append([]domain.ReasonCode(nil), candidate.Reasons...),
		})
		if len(explanations) == request.Limit {
			break
		}
	}
	if err := service.authority.AuthorizeColdStart(ctx, request.RequesterKey); err != nil {
		return nil, ErrUnavailable
	}
	return explanations, nil
}

func domainOpaqueKey(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' && character < 'a' || character > 'f' {
			return false
		}
	}
	return true
}
