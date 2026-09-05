// Package safety is the composition root of the Trust, Safety and Care
// bounded context slice for report/block intake (E12-S01). Triage queues,
// evidence viewers and action ladders land with E12-S02+.
package safety

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/internal/safety/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/safety/application"
)

type Module struct {
	Safety   application.SafetyService
	Cases    application.CaseService
	Actions  application.ActionService
	Evidence application.EvidenceService
	Care     application.CareService
}

// NewModule builds the safety context. Enforcement ports (identity
// suspension/block, session revocation) are injected at composition
// (agent_plan.md §7.2: cross-context calls happen at the root).
func NewModule(ctx context.Context, database *mongo.Database, outboxStore application.OutboxAppender, identity application.IdentityEnforcement, sessions application.SessionRevoker) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	caseRepository := mongodb.NewCaseRepository(database)
	if err := caseRepository.EnsureCaseIndexes(ctx); err != nil {
		return Module{}, err
	}
	actionLog := mongodb.NewActionLogStore(database)
	// The unique command index is what stops two operators clicking at once
	// from both writing an action, which would escalate the subject's next
	// one for a decision that was taken only once.
	if err := actionLog.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Safety:   application.NewSafetyService(repository, repository, outboxStore, time.Now, newID),
		Cases:    application.NewCaseService(caseRepository, time.Now, newID),
		Actions:  application.NewActionService(caseRepository, actionLog, identity, sessions, actionLog, time.Now, newID),
		Evidence: application.NewEvidenceService(repository, caseRepository, mongodb.NewAccessAuditStore(database), time.Now, newID),
		Care:     application.NewCareService(mongodb.NewCareRepository(database), mongodb.NewCareRepository(database), time.Now, newID),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "rep_" + base64.RawURLEncoding.EncodeToString(id)
}
