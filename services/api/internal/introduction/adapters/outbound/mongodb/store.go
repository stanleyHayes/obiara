// Package mongodb persists Voice of Introduction aggregates.
//
// Documents hold references and keyed fingerprints only: an asset id, a
// transcript id, a checksum, and the audit trail. No audio, no transcript
// text and no raw member identifier is written here — the bytes live in the
// object store under their own key, and this collection could be dumped
// without disclosing what anybody said.
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	platformmongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

type Store struct {
	introductions *mongo.Collection
}

func NewStore(database *mongo.Database) *Store {
	return &Store{introductions: database.Collection("voice_introductions")}
}

type eventDocument struct {
	CommandID   string    `bson:"commandId"`
	Fingerprint string    `bson:"fingerprint"`
	Action      string    `bson:"action"`
	OccurredAt  time.Time `bson:"occurredAt"`
	Version     uint64    `bson:"version"`
}

type introductionDocument struct {
	ID               string          `bson:"_id"`
	OwnerID          string          `bson:"ownerId"`
	ConsentPurposeID string          `bson:"consentPurposeId"`
	ConsentVersion   uint64          `bson:"consentVersion"`
	ConsentEvaluated time.Time       `bson:"consentEvaluatedAt"`
	AssetID          string          `bson:"assetId"`
	ContentType      string          `bson:"contentType"`
	Size             int64           `bson:"size"`
	DurationNanos    int64           `bson:"durationNanos"`
	Checksum         string          `bson:"checksum"`
	TranscriptID     string          `bson:"transcriptId,omitempty"`
	TranscriptLang   string          `bson:"transcriptLanguage,omitempty"`
	TranscriptConf   uint8           `bson:"transcriptConfidence,omitempty"`
	Status           string          `bson:"status"`
	DataStatus       string          `bson:"dataStatus"`
	RetentionUntil   time.Time       `bson:"retentionUntil,omitempty"`
	RetentionLegal   bool            `bson:"retentionLegalHold"`
	DeletionDueAt    time.Time       `bson:"deletionDueAt,omitempty"`
	CreatedAt        time.Time       `bson:"createdAt"`
	UpdatedAt        time.Time       `bson:"updatedAt"`
	Version          uint64          `bson:"version"`
	Events           []eventDocument `bson:"events"`
	// CreationCommandID is indexed uniquely so Create is idempotent by the
	// command that opened the aggregate, which is what the port promises.
	CreationCommandID string `bson:"creationCommandId"`
}

func (store *Store) EnsureIndexes(ctx context.Context) error {
	_, err := store.introductions.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// Idempotent creation: a retried BeginUpload must return the
			// aggregate the first attempt made, not open a second one.
			Keys:    bson.D{{Key: "creationCommandId", Value: 1}},
			Options: options.Index().SetName("voice_introduction_command_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "ownerId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("voice_introduction_owner"),
		},
		{
			// The retention sweep reads this: due, still holding data, and
			// not under legal hold.
			Keys:    bson.D{{Key: "deletionDueAt", Value: 1}, {Key: "dataStatus", Value: 1}},
			Options: options.Index().SetName("voice_introduction_retention"),
		},
	})
	return err
}

func (store *Store) Create(
	ctx context.Context,
	introduction domain.Introduction,
) (domain.Introduction, bool, error) {
	document := toDocument(introduction)
	if _, err := store.introductions.InsertOne(ctx, document); err == nil {
		return introduction, false, nil
	} else if !platformmongo.IsDuplicateKey(err) {
		return domain.Introduction{}, false, application.ErrDependencyUnavailable
	}

	// Replay: the creating command already opened one. Return that rather
	// than a second aggregate for the same recording.
	var existing introductionDocument
	if err := store.introductions.FindOne(ctx, bson.M{
		"creationCommandId": document.CreationCommandID,
	}).Decode(&existing); err != nil {
		return domain.Introduction{}, false, application.ErrDependencyUnavailable
	}
	return fromDocument(existing), true, nil
}

func (store *Store) FindByID(ctx context.Context, id string) (domain.Introduction, error) {
	var document introductionDocument
	if err := store.introductions.FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Introduction{}, application.ErrNotFound
		}
		return domain.Introduction{}, application.ErrDependencyUnavailable
	}
	return fromDocument(document), nil
}

// Update writes only when the stored version is the one the caller decided
// against. Two concurrent transitions would otherwise both succeed and the
// second would erase the first's audit entry.
func (store *Store) Update(
	ctx context.Context,
	introduction domain.Introduction,
	expectedVersion uint64,
	_ string,
) error {
	document := toDocument(introduction)
	result, err := store.introductions.ReplaceOne(
		ctx,
		bson.M{"_id": document.ID, "version": expectedVersion},
		document,
	)
	if err != nil {
		return application.ErrDependencyUnavailable
	}
	if result.MatchedCount != 1 {
		return application.ErrOptimisticConflict
	}
	return nil
}

// DueForPurge returns aggregates whose retention has elapsed and that still
// hold data, excluding anything under legal hold.
func (store *Store) DueForPurge(ctx context.Context, at time.Time, limit int) ([]domain.Introduction, error) {
	if limit < 1 {
		limit = 100
	}
	cursor, err := store.introductions.Find(ctx, bson.M{
		"deletionDueAt":      bson.M{"$gt": time.Time{}, "$lte": at.UTC()},
		"dataStatus":         string(domain.DataRetained),
		"retentionLegalHold": false,
	}, options.Find().SetSort(bson.D{{Key: "deletionDueAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, application.ErrDependencyUnavailable
	}
	defer cursor.Close(ctx)

	var due []domain.Introduction
	for cursor.Next(ctx) {
		var document introductionDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, application.ErrDependencyUnavailable
		}
		due = append(due, fromDocument(document))
	}
	if cursor.Err() != nil {
		return nil, application.ErrDependencyUnavailable
	}
	return due, nil
}

func toDocument(introduction domain.Introduction) introductionDocument {
	events := make([]eventDocument, 0, len(introduction.Events()))
	creationCommand := introduction.ID()
	for index, event := range introduction.Events() {
		events = append(events, eventDocument{
			CommandID: event.CommandID(), Fingerprint: event.Fingerprint(),
			Action: string(event.Action()), OccurredAt: event.OccurredAt(),
			Version: event.Version(),
		})
		if index == 0 {
			creationCommand = event.CommandID()
		}
	}
	return introductionDocument{
		ID:      introduction.ID(),
		OwnerID: introduction.OwnerID(),

		ConsentPurposeID: introduction.Consent().PurposeID(),
		ConsentVersion:   introduction.Consent().Version(),
		ConsentEvaluated: introduction.Consent().EvaluatedAt(),

		AssetID:       introduction.Media().AssetID(),
		ContentType:   introduction.Media().ContentType(),
		Size:          introduction.Media().Size(),
		DurationNanos: int64(introduction.Media().Duration()),
		Checksum:      introduction.Media().Checksum(),

		TranscriptID:   introduction.Transcript().ID(),
		TranscriptLang: introduction.Transcript().Language(),
		TranscriptConf: introduction.Transcript().Confidence(),

		Status:         string(introduction.Status()),
		DataStatus:     string(introduction.DataStatus()),
		RetentionUntil: introduction.Retention().Until(),
		RetentionLegal: introduction.Retention().LegalHold(),
		DeletionDueAt:  introduction.DeletionDueAt(),
		CreatedAt:      introduction.CreatedAt(),
		UpdatedAt:      introduction.UpdatedAt(),
		Version:        introduction.Version(),
		Events:         events,

		CreationCommandID: creationCommand,
	}
}

func fromDocument(document introductionDocument) domain.Introduction {
	events := make([]domain.Event, 0, len(document.Events))
	for _, event := range document.Events {
		events = append(events, domain.NewEvent(
			event.CommandID, event.Fingerprint, domain.Action(event.Action),
			event.OccurredAt, event.Version,
		))
	}
	return domain.Reconstitute(domain.ReconstituteParams{
		ID:      document.ID,
		OwnerID: document.OwnerID,
		Consent: domain.ReconstituteConsentSnapshot(
			document.ConsentPurposeID, document.ConsentVersion, document.ConsentEvaluated,
		),
		Media: domain.ReconstituteMediaRef(
			document.AssetID, document.ContentType, document.Size,
			time.Duration(document.DurationNanos), document.Checksum,
		),
		Transcript: domain.ReconstituteTranscriptRef(
			document.TranscriptID, document.TranscriptLang, document.TranscriptConf,
		),
		Status:        domain.Status(document.Status),
		DataStatus:    domain.DataStatus(document.DataStatus),
		Retention:     domain.NewRetention(document.RetentionUntil, document.RetentionLegal),
		DeletionDueAt: document.DeletionDueAt,
		CreatedAt:     document.CreatedAt,
		UpdatedAt:     document.UpdatedAt,
		Version:       document.Version,
		Events:        events,
	})
}
