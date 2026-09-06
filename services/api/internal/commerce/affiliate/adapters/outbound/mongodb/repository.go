// Package mongodb stores affiliates and the referrals they brought.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/domain"
)

type Repository struct {
	affiliates *mongo.Collection
	referrals  *mongo.Collection
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{
		affiliates: database.Collection("affiliates"),
		referrals:  database.Collection("affiliate_referrals"),
	}
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := repository.affiliates.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("affiliate_code"),
		},
		{
			Keys:    bson.D{{Key: "commands.id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("affiliate_command"),
		},
	}); err != nil {
		return err
	}
	_, err := repository.referrals.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// One referral per member. The first code keeps them.
			Keys:    bson.D{{Key: "_id", Value: 1}},
			Options: options.Index().SetName("affiliate_referral_member"),
		},
		{
			// The sweep: what is due and not yet settled either way.
			Keys:    bson.D{{Key: "settled", Value: 1}, {Key: "qualifiesAt", Value: 1}},
			Options: options.Index().SetName("affiliate_referral_due"),
		},
	})
	return err
}

type eventDocument struct {
	Sequence      uint64        `bson:"sequence"`
	CommandID     string        `bson:"commandId"`
	ReasonCode    string        `bson:"reasonCode"`
	Action        domain.Action `bson:"action"`
	AmountPesewas int64         `bson:"amountPesewas"`
	At            time.Time     `bson:"at"`
}

type commandDocument struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}

type document struct {
	ID           string        `bson:"_id"`
	Name         string        `bson:"name"`
	Code         string        `bson:"code"`
	ContactEmail string        `bson:"contactEmail"`
	Status       domain.Status `bson:"status"`
	// CountedKeys are digests. The row says how many referrals counted and
	// never which members they were.
	CountedKeys    []string          `bson:"countedKeys"`
	AccruedPesewas int64             `bson:"accruedPesewas"`
	PaidPesewas    int64             `bson:"paidPesewas"`
	Revision       uint64            `bson:"revision"`
	RegisteredAt   time.Time         `bson:"registeredAt"`
	Events         []eventDocument   `bson:"events"`
	Commands       []commandDocument `bson:"commands"`
}

func (repository *Repository) Create(ctx context.Context, affiliate domain.Affiliate) error {
	_, err := repository.affiliates.InsertOne(ctx, toDocument(affiliate))
	if err == nil {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		return application.ErrCodeTaken
	}
	return application.ErrUnavailable
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Affiliate, error) {
	return repository.find(ctx, bson.M{"_id": id})
}

func (repository *Repository) FindByCode(ctx context.Context, code string) (domain.Affiliate, error) {
	return repository.find(ctx, bson.M{"code": code})
}

func (repository *Repository) find(ctx context.Context, filter bson.M) (domain.Affiliate, error) {
	var stored document
	if err := repository.affiliates.FindOne(ctx, filter).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Affiliate{}, application.ErrNotFound
		}
		return domain.Affiliate{}, application.ErrUnavailable
	}
	return toDomain(stored)
}

// Append writes one transition onto the revision it was decided against.
//
// The revision guard is what keeps money straight under load: two sweeps
// accruing at once both read revision N, and only one write lands.
func (repository *Repository) Append(
	ctx context.Context, affiliate domain.Affiliate, expected uint64, commandID string,
) error {
	events, commands := affiliate.Events(), affiliate.Commands()
	if len(events) != int(expected+1) || len(commands) != int(expected+1) {
		return domain.ErrInvalidAffiliate
	}
	result, err := repository.affiliates.UpdateOne(ctx,
		bson.M{"_id": affiliate.ID(), "revision": expected},
		bson.M{
			"$set": bson.M{
				"status":         affiliate.Status(),
				"countedKeys":    affiliate.CountedKeys(),
				"accruedPesewas": affiliate.AccruedPesewas(),
				"paidPesewas":    affiliate.PaidPesewas(),
				"revision":       affiliate.Revision(),
			},
			"$push": bson.M{
				"events":   toEvent(events[len(events)-1]),
				"commands": toCommand(commands[len(commands)-1]),
			},
		})
	if err != nil {
		return application.ErrUnavailable
	}
	if result.MatchedCount == 0 {
		if repository.affiliates.FindOne(ctx, bson.M{"commands.id": commandID}).Err() == nil {
			// Already applied. The caller's aggregate said the same, so this
			// is a retry rather than a conflict.
			return nil
		}
		return application.ErrConflict
	}
	return nil
}

func (repository *Repository) List(ctx context.Context, limit int) ([]domain.Affiliate, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	cursor, err := repository.affiliates.Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "name", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, application.ErrUnavailable
	}
	defer cursor.Close(ctx)

	affiliates := make([]domain.Affiliate, 0, limit)
	for cursor.Next(ctx) {
		var stored document
		if err := cursor.Decode(&stored); err != nil {
			return nil, application.ErrUnavailable
		}
		affiliate, err := toDomain(stored)
		if err != nil {
			continue
		}
		affiliates = append(affiliates, affiliate)
	}
	if cursor.Err() != nil {
		return nil, application.ErrUnavailable
	}
	return affiliates, nil
}

type referralDocument struct {
	MemberKey   string    `bson:"_id"`
	MemberID    string    `bson:"memberId"`
	AffiliateID string    `bson:"affiliateId"`
	Code        string    `bson:"code"`
	RecordedAt  time.Time `bson:"recordedAt"`
	QualifiesAt time.Time `bson:"qualifiesAt"`
	Settled     string    `bson:"settled"`
}

// Record writes a referral, and refuses to move one that already exists: the
// first code keeps the member.
func (repository *Repository) Record(ctx context.Context, referral application.Referral) error {
	_, err := repository.referrals.InsertOne(ctx, referralDocument{
		MemberKey: referral.MemberKey, MemberID: referral.MemberID,
		AffiliateID: referral.AffiliateID,
		Code:        referral.Code, RecordedAt: referral.RecordedAt.UTC(),
		QualifiesAt: referral.QualifiesAt.UTC(), Settled: referral.Settled,
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}
		return application.ErrUnavailable
	}
	return nil
}

func (repository *Repository) Find(
	ctx context.Context, memberKey string,
) (application.Referral, error) {
	var stored referralDocument
	if err := repository.referrals.FindOne(ctx, bson.M{"_id": memberKey}).Decode(&stored); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return application.Referral{}, application.ErrNotFound
		}
		return application.Referral{}, application.ErrUnavailable
	}
	return application.Referral{
		MemberKey: stored.MemberKey, MemberID: stored.MemberID,
		AffiliateID: stored.AffiliateID, Code: stored.Code,
		RecordedAt: stored.RecordedAt, QualifiesAt: stored.QualifiesAt, Settled: stored.Settled,
	}, nil
}

func (repository *Repository) DueForQualification(
	ctx context.Context, at time.Time, limit int,
) ([]application.Referral, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	cursor, err := repository.referrals.Find(ctx, bson.M{
		"settled":     "",
		"qualifiesAt": bson.M{"$lte": at.UTC()},
	}, options.Find().SetSort(bson.D{{Key: "qualifiesAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, application.ErrUnavailable
	}
	defer cursor.Close(ctx)

	due := make([]application.Referral, 0, limit)
	for cursor.Next(ctx) {
		var stored referralDocument
		if err := cursor.Decode(&stored); err != nil {
			return nil, application.ErrUnavailable
		}
		due = append(due, application.Referral{
			MemberKey: stored.MemberKey, MemberID: stored.MemberID,
			AffiliateID: stored.AffiliateID, Code: stored.Code,
			RecordedAt: stored.RecordedAt, QualifiesAt: stored.QualifiesAt, Settled: stored.Settled,
		})
	}
	return due, cursor.Err()
}

func (repository *Repository) MarkSettled(
	ctx context.Context, memberKey, outcome string,
) error {
	_, err := repository.referrals.UpdateOne(ctx,
		bson.M{"_id": memberKey},
		bson.M{"$set": bson.M{"settled": outcome}})
	if err != nil {
		return application.ErrUnavailable
	}
	return nil
}

func toDocument(affiliate domain.Affiliate) document {
	stored := document{
		ID: affiliate.ID(), Name: affiliate.Name(), Code: affiliate.Code(),
		ContactEmail: affiliate.ContactEmail(), Status: affiliate.Status(),
		CountedKeys: affiliate.CountedKeys(), AccruedPesewas: affiliate.AccruedPesewas(),
		PaidPesewas: affiliate.PaidPesewas(), Revision: affiliate.Revision(),
		RegisteredAt: affiliate.RegisteredAt(),
	}
	for _, event := range affiliate.Events() {
		stored.Events = append(stored.Events, toEvent(event))
	}
	for _, command := range affiliate.Commands() {
		stored.Commands = append(stored.Commands, toCommand(command))
	}
	return stored
}

func toDomain(stored document) (domain.Affiliate, error) {
	state := domain.State{
		ID: stored.ID, Name: stored.Name, Code: stored.Code,
		ContactEmail: stored.ContactEmail, Status: stored.Status,
		CountedKeys: stored.CountedKeys, AccruedPesewas: stored.AccruedPesewas,
		PaidPesewas: stored.PaidPesewas, Revision: stored.Revision,
		RegisteredAt: stored.RegisteredAt,
	}
	for _, event := range stored.Events {
		state.Events = append(state.Events, domain.Event{
			Sequence: event.Sequence, CommandID: event.CommandID, ReasonCode: event.ReasonCode,
			Action: event.Action, AmountPesewas: event.AmountPesewas, At: event.At,
		})
	}
	for _, command := range stored.Commands {
		state.Commands = append(state.Commands, domain.AppliedCommand{
			ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
		})
	}
	return domain.Rehydrate(state)
}

func toEvent(event domain.Event) eventDocument {
	return eventDocument{
		Sequence: event.Sequence, CommandID: event.CommandID, ReasonCode: event.ReasonCode,
		Action: event.Action, AmountPesewas: event.AmountPesewas, At: event.At,
	}
}

func toCommand(command domain.AppliedCommand) commandDocument {
	return commandDocument{
		ID: command.ID, Fingerprint: command.Fingerprint, Revision: command.Revision,
	}
}
