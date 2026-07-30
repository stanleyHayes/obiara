package application

import (
	"context"
	"fmt"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/domain"
)

const (
	CommunityPromiseID = "promise.community"
	ServiceTermsID     = "terms.service"
	AdultAgeID         = "age.adult"
	CurrentVersion     = uint64(1)
)

type OnboardingService struct {
	consents Service
}

func NewOnboardingService(consents Service) OnboardingService {
	return OnboardingService{consents: consents}
}

type OnboardingCommand struct {
	CommandID string
	SubjectID string
	Source    domain.Source
}

type OnboardingResult struct {
	PromiseRevision uint64
	TermsRevision   uint64
	AgeRevision     uint64
}

func (service OnboardingService) Accept(ctx context.Context, command OnboardingCommand) (OnboardingResult, error) {
	specs := []struct {
		id       string
		evidence domain.EvidenceKind
		ref      string
	}{
		{CommunityPromiseID, domain.EvidenceAcknowledgement, "evidence.promise.v1"},
		{ServiceTermsID, domain.EvidenceAcknowledgement, "evidence.terms.v1"},
		{AdultAgeID, domain.EvidenceAgeAffirmation, "evidence.age.v1"},
	}
	revisions := make([]uint64, 0, len(specs))
	for _, spec := range specs {
		evidence, err := domain.NewEvidence(spec.evidence, CurrentVersion, spec.ref)
		if err != nil {
			return OnboardingResult{}, err
		}
		result, err := service.consents.Grant(ctx, Command{
			CommandID: command.CommandID + ":" + spec.id,
			SubjectID: command.SubjectID, PurposeID: spec.id,
			PurposeVersion: CurrentVersion, ExpectedRevision: 0,
			ActorID: command.SubjectID, ActorKind: domain.ActorSubject,
			Source: command.Source, Evidence: evidence,
		})
		if err != nil {
			return OnboardingResult{}, fmt.Errorf("%s: %w", spec.id, err)
		}
		revisions = append(revisions, result.Record.Revision())
	}
	return OnboardingResult{
		PromiseRevision: revisions[0],
		TermsRevision:   revisions[1],
		AgeRevision:     revisions[2],
	}, nil
}
