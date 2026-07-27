package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/application"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/domain"
)

// NominationRepository persists nominations in the `nominations` collection.
type NominationRepository struct {
	col *mongo.Collection
}

// NewNominationRepository constructs the repository and ensures indexes.
func NewNominationRepository(ctx context.Context, db *mongo.Database) (*NominationRepository, error) {
	col := db.Collection("nominations")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "memberId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "memberId", Value: 1}, {Key: "kinPhone", Value: 1}, {Key: "status", Value: 1}}},
	})
	if err != nil {
		return nil, fmt.Errorf("create nomination indexes: %w", err)
	}
	return &NominationRepository{col: col}, nil
}

type nominationDoc struct {
	ID           string     `bson:"_id"`
	MemberID     string     `bson:"memberId"`
	KinName      string     `bson:"kinName"`
	KinPhone     string     `bson:"kinPhone"`
	Relationship string     `bson:"relationship"`
	Status       string     `bson:"status"`
	Version      int64      `bson:"version"`
	CreatedAt    time.Time  `bson:"createdAt"`
	RespondedAt  *time.Time `bson:"respondedAt,omitempty"`
}

func toDoc(n domain.Nomination) nominationDoc {
	return nominationDoc{
		ID: n.ID, MemberID: n.MemberID, KinName: n.KinName, KinPhone: n.KinPhone,
		Relationship: string(n.Relationship), Status: string(n.Status),
		Version: n.Version, CreatedAt: n.CreatedAt, RespondedAt: n.RespondedAt,
	}
}

func (d nominationDoc) toDomain() domain.Nomination {
	return domain.Nomination{
		ID: d.ID, MemberID: d.MemberID, KinName: d.KinName, KinPhone: d.KinPhone,
		Relationship: domain.Relationship(d.Relationship), Status: domain.Status(d.Status),
		Version: d.Version, CreatedAt: d.CreatedAt, RespondedAt: d.RespondedAt,
	}
}

// Create inserts a nomination.
func (r *NominationRepository) Create(ctx context.Context, n domain.Nomination) error {
	if _, err := r.col.InsertOne(ctx, toDoc(n)); err != nil {
		return fmt.Errorf("insert nomination: %w", err)
	}
	return nil
}

// FindByID loads a nomination.
func (r *NominationRepository) FindByID(ctx context.Context, id string) (domain.Nomination, error) {
	var d nominationDoc
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Nomination{}, application.ErrNominationNotFound
	}
	if err != nil {
		return domain.Nomination{}, fmt.Errorf("find nomination: %w", err)
	}
	return d.toDomain(), nil
}

// Update replaces a nomination with optimistic concurrency on version.
func (r *NominationRepository) Update(ctx context.Context, n domain.Nomination) error {
	d := toDoc(n)
	prev := d.Version - 1
	res, err := r.col.ReplaceOne(ctx, bson.M{"_id": d.ID, "version": prev}, d)
	if err != nil {
		return fmt.Errorf("update nomination: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("update nomination: version conflict")
	}
	return nil
}

// ListByMember lists a member's nominations, latest first.
func (r *NominationRepository) ListByMember(ctx context.Context, memberID string) ([]domain.Nomination, error) {
	cur, err := r.col.Find(ctx, bson.M{"memberId": memberID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("list nominations: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	out := []domain.Nomination{}
	for cur.Next(ctx) {
		var d nominationDoc
		if err := cur.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode nomination: %w", err)
		}
		out = append(out, d.toDomain())
	}
	return out, cur.Err()
}

// HasPendingForKin reports whether a pending nomination exists for this member+kin phone.
func (r *NominationRepository) HasPendingForKin(ctx context.Context, memberID, kinPhone string) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{
		"memberId": memberID, "kinPhone": kinPhone, "status": string(domain.StatusPending),
	})
	if err != nil {
		return false, fmt.Errorf("count pending nominations: %w", err)
	}
	return count > 0, nil
}
