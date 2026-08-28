// Package admin is the composition root of the Admin and Configuration
// bounded context slice for principals and MFA (E16-S01). Desk modules
// (verification, T&S, finance) build on these sessions and roles.
package admin

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/admin/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/application"
)

type Module struct {
	Admin application.AdminService
}

// NewModule builds the admin auth slice. sender bridges the email channel
// (Resend) for MFA code delivery.
func NewModule(ctx context.Context, database *mongo.Database, sender application.OperatorMailer) (Module, error) {
	principals := mongodb.NewPrincipalRepository(database)
	if err := principals.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	challenges := mongodb.NewChallengeRepository(database)
	if err := challenges.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	sessions := mongodb.NewSessionRepository(database)
	if err := sessions.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Admin: application.NewAdminService(principals, challenges, sessions, mongodb.NewAuditStore(database), sender, time.Now, newID),
	}, nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "adm_" + base64.RawURLEncoding.EncodeToString(id)
}
