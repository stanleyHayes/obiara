package mongodb

import (
	"context"
	"errors"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/retention/application"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/retention/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct{ database *mongo.Database }

func New(database *mongo.Database) *Store  { return &Store{database} }
func (s *Store) events() *mongo.Collection { return s.database.Collection("analytics_events") }
func (s *Store) receipts() *mongo.Collection {
	return s.database.Collection("analytics_retention_receipts")
}
func (s *Store) aggregates() *mongo.Collection {
	return s.database.Collection("analytics_monthly_aggregates")
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	if _, e := s.events().Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "occurredAt", Value: 1}, {Key: "pseudonymizedAt", Value: 1}}, Options: options.Index().SetName("analytics_retention_due")}, {Keys: bson.D{{Key: "retentionLeaseUntil", Value: 1}}, Options: options.Index().SetName("analytics_retention_lease")}}); e != nil {
		return e
	}
	if _, e := s.receipts().Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "eventId", Value: 1}, {Key: "action", Value: 1}, {Key: "policyVersion", Value: 1}}, Options: options.Index().SetUnique(true).SetName("analytics_retention_once")}); e != nil {
		return e
	}
	_, e := s.aggregates().Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "name", Value: 1}, {Key: "month", Value: 1}}, Options: options.Index().SetUnique(true).SetName("analytics_monthly_unique")})
	return e
}

type eventDocument struct {
	ID                  bson.ObjectID `bson:"_id"`
	Name                string        `bson:"name"`
	SubjectRef          string        `bson:"subjectRef"`
	OccurredAt          time.Time     `bson:"occurredAt"`
	PseudonymizedAt     time.Time     `bson:"pseudonymizedAt,omitempty"`
	PseudonymKeyVersion uint64        `bson:"pseudonymKeyVersion,omitempty"`
}

func (s *Store) ClaimDue(ctx context.Context, now time.Time, limit int, lease string, leaseUntil time.Time) ([]domain.Candidate, error) {
	if limit < 1 || limit > domain.MaxBatch {
		return nil, domain.ErrInvalid
	}
	filter := bson.M{"$and": []bson.M{{"$or": []bson.M{{"retentionLeaseUntil": bson.M{"$exists": false}}, {"retentionLeaseUntil": bson.M{"$lte": now}}}}, {"$or": []bson.M{{"occurredAt": bson.M{"$lte": now.AddDate(0, -13, 0)}}, {"occurredAt": bson.M{"$lte": now.Add(-domain.PseudonymizeAfter)}, "pseudonymizedAt": bson.M{"$exists": false}}}}}}
	result := make([]domain.Candidate, 0, limit)
	for range limit {
		var doc eventDocument
		e := s.events().FindOneAndUpdate(ctx, filter, bson.M{"$set": bson.M{"retentionLease": lease, "retentionLeaseUntil": leaseUntil}}, options.FindOneAndUpdate().SetSort(bson.D{{Key: "occurredAt", Value: 1}}).SetReturnDocument(options.After)).Decode(&doc)
		if errors.Is(e, mongo.ErrNoDocuments) {
			break
		}
		if e != nil {
			return nil, e
		}
		result = append(result, domain.Candidate{ID: doc.ID.Hex(), Name: doc.Name, SubjectRef: doc.SubjectRef, OccurredAt: doc.OccurredAt, PseudonymizedAt: doc.PseudonymizedAt, PseudonymKeyVersion: doc.PseudonymKeyVersion})
	}
	return result, nil
}

type receipt struct {
	ID            string        `bson:"_id"`
	EventID       string        `bson:"eventId"`
	Action        domain.Action `bson:"action"`
	PolicyID      string        `bson:"policyId"`
	PolicyVersion uint64        `bson:"policyVersion"`
	ProcessedAt   time.Time     `bson:"processedAt"`
}

func (s *Store) Pseudonymize(ctx context.Context, c domain.Candidate, d domain.Decision, newRef, receiptID string) error {
	id, e := bson.ObjectIDFromHex(c.ID)
	if e != nil {
		return domain.ErrInvalid
	}
	if s.applied(ctx, c.ID, d) {
		return application.ErrApplied
	}
	e = apimongo.WithTransaction(ctx, s.database.Client(), func(session context.Context) error {
		updated, x := s.events().UpdateOne(session, bson.M{"_id": id, "subjectRef": c.SubjectRef, "pseudonymizedAt": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"subjectRef": newRef, "pseudonymizedAt": d.ProcessedAt, "pseudonymKeyVersion": d.PseudonymKeyVersion}, "$unset": bson.M{"retentionLease": "", "retentionLeaseUntil": ""}})
		if x != nil {
			return x
		}
		if updated.MatchedCount != 1 {
			return application.ErrConflict
		}
		_, x = s.receipts().InsertOne(session, receipt{ID: receiptID, EventID: c.ID, Action: d.Action, PolicyID: d.PolicyID, PolicyVersion: d.PolicyVersion, ProcessedAt: d.ProcessedAt})
		return x
	})
	if e != nil && s.applied(ctx, c.ID, d) {
		return application.ErrApplied
	}
	return e
}
func (s *Store) AggregateErase(ctx context.Context, c domain.Candidate, d domain.Decision, receiptID string) error {
	id, e := bson.ObjectIDFromHex(c.ID)
	if e != nil {
		return domain.ErrInvalid
	}
	if s.applied(ctx, c.ID, d) {
		return application.ErrApplied
	}
	e = apimongo.WithTransaction(ctx, s.database.Client(), func(session context.Context) error {
		if _, x := s.receipts().InsertOne(session, receipt{ID: receiptID, EventID: c.ID, Action: d.Action, PolicyID: d.PolicyID, PolicyVersion: d.PolicyVersion, ProcessedAt: d.ProcessedAt}); x != nil {
			return x
		}
		if _, x := s.aggregates().UpdateOne(session, bson.M{"name": c.Name, "month": d.AggregateMonth}, bson.M{"$inc": bson.M{"count": 1}, "$setOnInsert": bson.M{"name": c.Name, "month": d.AggregateMonth}}, options.UpdateOne().SetUpsert(true)); x != nil {
			return x
		}
		deleted, x := s.events().DeleteOne(session, bson.M{"_id": id})
		if x != nil {
			return x
		}
		if deleted.DeletedCount != 1 {
			return application.ErrConflict
		}
		return nil
	})
	if e != nil && s.applied(ctx, c.ID, d) {
		return application.ErrApplied
	}
	return e
}
func (s *Store) applied(ctx context.Context, eventID string, d domain.Decision) bool {
	return s.receipts().FindOne(ctx, bson.M{"eventId": eventID, "action": d.Action, "policyVersion": d.PolicyVersion}).Err() == nil
}
