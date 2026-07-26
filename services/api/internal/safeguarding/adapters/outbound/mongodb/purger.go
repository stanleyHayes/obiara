package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
)

// ArtifactPurger deletes known subject-owned Mongo records transactionally.
// External object/media providers implement the same application port and
// should be composed with this adapter before production.
type ArtifactPurger struct {
	database *mongo.Database
}

func NewArtifactPurger(database *mongo.Database) *ArtifactPurger {
	return &ArtifactPurger{database: database}
}

func (purger *ArtifactPurger) Purge(ctx context.Context, subjectID, sourceRef string) error {
	return apimongo.WithTransaction(ctx, purger.database.Client(), func(transaction context.Context) error {
		deletions := []struct {
			collection string
			filter     bson.M
		}{
			{"verification_cases", bson.M{"$or": bson.A{bson.M{"accountId": subjectID}, bson.M{"_id": sourceRef}}}},
			{"accounts", bson.M{"_id": subjectID}},
			{"sessions", bson.M{"memberId": subjectID}},
			{"consent_records", bson.M{"subjectId": subjectID}},
			{"media_assets", bson.M{"ownerId": subjectID}},
		}
		for _, deletion := range deletions {
			if _, err := purger.database.Collection(deletion.collection).DeleteMany(transaction, deletion.filter); err != nil {
				return err
			}
		}
		return nil
	})
}
