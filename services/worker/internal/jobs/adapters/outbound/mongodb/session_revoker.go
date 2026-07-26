// Package mongodb provides the worker-side session revocation adapter for
// deletion processing. It mirrors the identity context's
// RevokeAllForMember scope against the shared sessions collection; if a
// third consumer appears, the session repository should move to
// internal/platform.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SessionRevoker struct {
	database *mongo.Database
	clock    func() time.Time
}

func NewSessionRevoker(database *mongo.Database, clock func() time.Time) *SessionRevoker {
	return &SessionRevoker{database: database, clock: clock}
}

func (revoker *SessionRevoker) RevokeMemberSessions(ctx context.Context, memberID string) error {
	_, err := revoker.database.Collection("sessions").UpdateMany(ctx,
		bson.M{"memberId": memberID, "status": "active"},
		bson.A{
			bson.M{"$set": bson.M{"status": "revoked", "updatedAt": revoker.clock().UTC()}},
			bson.M{"$set": bson.M{"version": bson.M{"$add": bson.A{"$version", 1}}}},
		})
	return err
}
