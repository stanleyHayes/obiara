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
		// Target the revision that is actually stored. This used to be a
		// hardcoded zero, which meant "I expect no prior record" — true only
		// on a member's very first attempt. Anyone who accepted the Promise
		// and then dropped off before finishing the walk came back to a
		// permanent 409: the receipt they had already given was read as a
		// stale write, and there was no way past the step that produced it.
		revision, err := service.consents.Revision(ctx, command.SubjectID, spec.id)
		if err != nil {
			return OnboardingResult{}, fmt.Errorf("%s: %w", spec.id, err)
		}

		// Already given, at this exact version: the end state this command
		// asks for is the one that holds. Re-granting would either conflict
		// or write a second receipt for a choice made once.
		effective, err := service.consents.Effective(ctx, command.SubjectID, spec.id, CurrentVersion)
		if err != nil {
			return OnboardingResult{}, fmt.Errorf("%s: %w", spec.id, err)
		}
		if effective {
			revisions = append(revisions, revision)
			continue
		}

		result, err := service.consents.Grant(ctx, Command{
			CommandID: command.CommandID + ":" + spec.id,
			SubjectID: command.SubjectID, PurposeID: spec.id,
			PurposeVersion: CurrentVersion, ExpectedRevision: revision,
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
