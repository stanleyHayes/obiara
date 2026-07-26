package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/safety/sikashield/domain"
	"go.uber.org/mock/gomock"
)

func TestConsentedVoiceMetadataRoutesHumanCaseOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalog := NewMockCatalog(ctrl)
	metrics := NewMockMetricsGate(ctrl)
	evidence := NewMockEvidenceVerifier(ctrl)
	cases := NewMockCaseRouter(ctrl)
	authority := NewMockAuthority(ctrl)
	now := time.Date(2026, 7, 26, 17, 0, 0, 0, time.UTC)
	p, _ := domain.NewPattern("coercive.payment", 2, domain.SourceVoiceMetadata, "pattern:2", "reviewer:key", now.Add(-time.Hour))
	m, _ := domain.NewMetrics(2000, 200, 196, 4, 100)
	signal := domain.Signal{PatternKey: p.Key, PatternVersion: p.Version, Source: p.Source, EvidenceRef: "voice:derived:metadata", Confidence: .99}
	authority.EXPECT().RequireOfflineEvaluator(gomock.Any(), "evaluator").Return(nil)
	catalog.EXPECT().Current(gomock.Any(), p.Key).Return(p, nil)
	metrics.EXPECT().Current(gomock.Any(), p.Key, p.Version).Return(m, nil)
	evidence.EXPECT().Revalidate(gomock.Any(), signal.EvidenceRef, domain.SourceVoiceMetadata).Return(nil)
	cases.EXPECT().OpenHumanCase(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, d domain.Decision) error {
		if d.Outcome != domain.OutcomeHumanReview {
			t.Fatalf("unexpected route: %+v", d)
		}
		return nil
	})
	service := NewService(catalog, metrics, evidence, cases, authority, func() time.Time { return now })
	d, err := service.Evaluate(context.Background(), Request{Actor: "evaluator", Signal: signal})
	if err != nil || d.Outcome != domain.OutcomeHumanReview {
		t.Fatalf("decision=%+v err=%v", d, err)
	}
}

func TestEvidenceOrRouterFailureFailsClosedToNoAction(t *testing.T) {
	for _, routerFails := range []bool{false, true} {
		t.Run(map[bool]string{false: "consent stale", true: "router unavailable"}[routerFails], func(t *testing.T) {
			ctrl := gomock.NewController(t)
			catalog := NewMockCatalog(ctrl)
			metrics := NewMockMetricsGate(ctrl)
			evidence := NewMockEvidenceVerifier(ctrl)
			cases := NewMockCaseRouter(ctrl)
			authority := NewMockAuthority(ctrl)
			now := time.Date(2026, 7, 26, 17, 0, 0, 0, time.UTC)
			p, _ := domain.NewPattern("coercive.payment", 2, domain.SourceVoiceMetadata, "pattern:2", "reviewer:key", now.Add(-time.Hour))
			m, _ := domain.NewMetrics(2000, 200, 196, 4, 100)
			signal := domain.Signal{PatternKey: p.Key, PatternVersion: p.Version, Source: p.Source, EvidenceRef: "voice:derived:metadata", Confidence: .99}
			authority.EXPECT().RequireOfflineEvaluator(gomock.Any(), "evaluator").Return(nil)
			catalog.EXPECT().Current(gomock.Any(), p.Key).Return(p, nil)
			metrics.EXPECT().Current(gomock.Any(), p.Key, p.Version).Return(m, nil)
			if routerFails {
				evidence.EXPECT().Revalidate(gomock.Any(), signal.EvidenceRef, signal.Source).Return(nil)
				cases.EXPECT().OpenHumanCase(gomock.Any(), gomock.Any()).Return(errors.New("unavailable"))
			} else {
				evidence.EXPECT().Revalidate(gomock.Any(), signal.EvidenceRef, signal.Source).Return(errors.New("withdrawn"))
			}
			service := NewService(catalog, metrics, evidence, cases, authority, func() time.Time { return now })
			d, err := service.Evaluate(context.Background(), Request{Actor: "evaluator", Signal: signal})
			if err != nil || d.Outcome != domain.OutcomeNoAction {
				t.Fatalf("decision=%+v err=%v", d, err)
			}
		})
	}
}
