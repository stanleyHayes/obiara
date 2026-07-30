package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
)

type Catalog struct {
	database   *mongo.Database
	collection *mongo.Collection
}

func NewCatalog(database *mongo.Database) *Catalog {
	return &Catalog{database: database, collection: database.Collection("matchmaker_licenses")}
}

type profileDocument struct {
	MatchmakerKey        string    `bson:"_id"`
	LicenseID            string    `bson:"licenseId"`
	Jurisdiction         string    `bson:"jurisdiction"`
	Version              uint64    `bson:"version"`
	ValidFrom            time.Time `bson:"validFrom"`
	ValidUntil           time.Time `bson:"validUntil"`
	MinimumFeePesewas    uint64    `bson:"minimumFeePesewas"`
	MaximumFeePesewas    uint64    `bson:"maximumFeePesewas"`
	DisplayName          string    `bson:"displayName"`
	Languages            []string  `bson:"languages"`
	Specialties          []string  `bson:"specialties"`
	CompletedEngagements uint64    `bson:"completedEngagements"`
	RatingBasisPoints    uint16    `bson:"ratingBasisPoints"`
}

func (catalog *Catalog) EnsureIndexes(ctx context.Context) error {
	_, err := catalog.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "licenseId", Value: 1}},
			Options: options.Index().SetName("matchmaker_license_id_unique").SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "jurisdiction", Value: 1},
				{Key: "validFrom", Value: 1},
				{Key: "validUntil", Value: 1},
			},
			Options: options.Index().SetName("matchmaker_license_current"),
		},
	})
	return err
}

func (catalog *Catalog) Current(ctx context.Context, matchmakerKey string) (domain.License, error) {
	var document profileDocument
	if err := catalog.collection.FindOne(ctx, bson.M{"_id": matchmakerKey}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.License{}, application.ErrUnavailable
		}
		return domain.License{}, err
	}
	return toProfile(document).License, nil
}

func (catalog *Catalog) ListCurrent(ctx context.Context, at time.Time) ([]domain.LicensedProfile, error) {
	cursor, err := catalog.collection.Find(
		ctx,
		bson.M{"validFrom": bson.M{"$lte": at.UTC()}, "validUntil": bson.M{"$gt": at.UTC()}},
		options.Find().SetSort(bson.D{{Key: "displayName", Value: 1}, {Key: "_id", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var documents []profileDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	profiles := make([]domain.LicensedProfile, 0, len(documents))
	for _, document := range documents {
		profile := toProfile(document)
		if !profile.Valid(at) {
			return nil, application.ErrUnavailable
		}
		profiles = append(profiles, profile.Clone())
	}
	return profiles, nil
}

func (catalog *Catalog) ListAll(ctx context.Context) ([]domain.LicensedProfile, error) {
	cursor, err := catalog.collection.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(bson.D{{Key: "displayName", Value: 1}, {Key: "_id", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var documents []profileDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	profiles := make([]domain.LicensedProfile, 0, len(documents))
	for _, document := range documents {
		profiles = append(profiles, toProfile(document).Clone())
	}
	return profiles, nil
}

// Put creates a license at expectedVersion zero or replaces the exact prior
// version. This is reserved for the MFA-gated admin licensing boundary.
func (catalog *Catalog) Put(ctx context.Context, profile domain.LicensedProfile, expectedVersion uint64, actorID string) error {
	if !profile.Valid(profile.License.ValidFrom) || profile.License.Version != expectedVersion+1 {
		return application.ErrInvalid
	}
	document := toProfileDocument(profile)
	return apimongo.WithTransaction(ctx, catalog.database.Client(), func(tx context.Context) error {
		if expectedVersion == 0 {
			if _, err := catalog.collection.InsertOne(tx, document); err != nil {
				if mongo.IsDuplicateKeyError(err) {
					return application.ErrConflict
				}
				return err
			}
		} else {
			result, err := catalog.collection.ReplaceOne(
				tx,
				bson.M{"_id": profile.License.MatchmakerKey, "version": expectedVersion},
				document,
			)
			if err != nil {
				if mongo.IsDuplicateKeyError(err) {
					return application.ErrConflict
				}
				return err
			}
			if result.MatchedCount == 0 {
				return application.ErrConflict
			}
		}
		_, err := catalog.database.Collection("admin_access").InsertOne(tx, bson.M{
			"actorId": actorID, "action": "admin.matchmaker.license",
			"target": profile.License.MatchmakerKey, "at": time.Now().UTC(),
			"version": profile.License.Version,
		})
		return err
	})
}

func toProfile(document profileDocument) domain.LicensedProfile {
	return domain.LicensedProfile{
		License: domain.License{
			ID: document.LicenseID, MatchmakerKey: document.MatchmakerKey,
			Jurisdiction: document.Jurisdiction, Version: document.Version,
			ValidFrom: document.ValidFrom, ValidUntil: document.ValidUntil,
			MinimumFeePesewas: document.MinimumFeePesewas,
			MaximumFeePesewas: document.MaximumFeePesewas,
		},
		DisplayName: document.DisplayName, Languages: append([]string(nil), document.Languages...),
		Specialties:          append([]string(nil), document.Specialties...),
		CompletedEngagements: document.CompletedEngagements,
		RatingBasisPoints:    document.RatingBasisPoints,
	}
}

func toProfileDocument(profile domain.LicensedProfile) profileDocument {
	return profileDocument{
		MatchmakerKey: profile.License.MatchmakerKey, LicenseID: profile.License.ID,
		Jurisdiction: profile.License.Jurisdiction, Version: profile.License.Version,
		ValidFrom: profile.License.ValidFrom.UTC(), ValidUntil: profile.License.ValidUntil.UTC(),
		MinimumFeePesewas: profile.License.MinimumFeePesewas,
		MaximumFeePesewas: profile.License.MaximumFeePesewas,
		DisplayName:       profile.DisplayName, Languages: append([]string(nil), profile.Languages...),
		Specialties:          append([]string(nil), profile.Specialties...),
		CompletedEngagements: profile.CompletedEngagements,
		RatingBasisPoints:    profile.RatingBasisPoints,
	}
}
