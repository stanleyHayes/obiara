// Package mongodb persists sow screening reviews awaiting a person.
package mongodb

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/domain"
)

var (
	// ErrReviewNotFound reports a reference no review answers to.
	ErrReviewNotFound = errors.New("screening review not found")
	// ErrActorRequired refuses an anonymous read of members' words.
	ErrActorRequired = errors.New("reading the screening queue requires an actor")
)

// The store exists to be this port. Asserting it here means a signature drift
// fails the build rather than the composition root.
var _ application.HumanReview = (*ReviewStore)(nil)

// ReviewStore implements the screening context's HumanReview port and serves
// the queue a reviewer works from.
//
// It holds the sow's words. That is unavoidable — a person cannot review
// something they cannot read — so the content lives here beside the decision
// rather than in the aggregate, and the retention table deletes it once the
// decision is old enough that nobody needs to see it again.
type ReviewStore struct {
	database *mongo.Database
	now      func() time.Time
}

func NewReviewStore(database *mongo.Database, now func() time.Time) *ReviewStore {
	if now == nil {
		now = time.Now
	}
	return &ReviewStore{database: database, now: now}
}

func (store *ReviewStore) reviews() *mongo.Collection {
	return store.database.Collection("seed_screening_reviews")
}

// accessLog records who read members' words and when.
//
// The safety context audits every evidence read, and this queue holds the
// same kind of thing: a member's own writing, shown to staff. Reading it is
// the job, so the record is not a suspicion — it is what makes an
// insider-access review possible at all.
func (store *ReviewStore) accessLog() *mongo.Collection {
	return store.database.Collection("seed_screening_access_log")
}

type accessDocument struct {
	ActorID  string    `bson:"actorId"`
	Reviewed int       `bson:"reviewed"`
	ReadAt   time.Time `bson:"readAt"`
}

type mediaDocument struct {
	MIME       string `bson:"mime"`
	Bytes      int64  `bson:"bytes"`
	DurationMs int64  `bson:"durationMs"`
}

type reviewDocument struct {
	ID     string `bson:"_id"`
	Reason string `bson:"reason"`
	Status string `bson:"status"`
	// Text is the member's own words, held only until the decision ages out.
	Text          string          `bson:"text"`
	LocaleTag     string          `bson:"localeTag"`
	LocaleVersion uint64          `bson:"localeVersion"`
	Media         []mediaDocument `bson:"media"`
	// AdvisoryStatus and AdvisoryReasons are what the automated pass thought.
	// They are shown to the reviewer as an opinion, never as a decision.
	AdvisoryStatus     string     `bson:"advisoryStatus,omitempty"`
	AdvisoryReasons    []string   `bson:"advisoryReasons,omitempty"`
	AdvisoryConfidence int        `bson:"advisoryConfidence,omitempty"`
	RoutedAt           time.Time  `bson:"routedAt"`
	DecidedAt          *time.Time `bson:"decidedAt,omitempty"`
	DecidedBy          string     `bson:"decidedBy,omitempty"`
	CommandID          string     `bson:"commandId,omitempty"`
}

func (store *ReviewStore) EnsureIndexes(ctx context.Context) error {
	if _, err := store.accessLog().Indexes().CreateOne(ctx, mongo.IndexModel{
		// Insider-access reviews read this by agent and by time.
		Keys:    bson.D{{Key: "actorId", Value: 1}, {Key: "readAt", Value: -1}},
		Options: options.Index().SetName("seed_screening_access_by_agent"),
	}); err != nil {
		return err
	}
	_, err := store.reviews().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// The queue: oldest pending first, so nothing waits behind
			// something newer.
			Keys:    bson.D{{Key: "status", Value: 1}, {Key: "routedAt", Value: 1}},
			Options: options.Index().SetName("seed_screening_review_queue"),
		},
		{
			// One decision per request, so two reviewers clicking at once
			// cannot both settle the same review.
			Keys: bson.D{{Key: "commandId", Value: 1}},
			Options: options.Index().SetName("seed_screening_review_command_unique").
				SetUnique(true).SetPartialFilterExpression(bson.M{"commandId": bson.M{"$type": "string"}}),
		},
	})
	return err
}

// Route persists a review and returns the reference screening hands back.
//
// The reference is the review's own id: screening validates it as a 64-hex
// digest and the sow stores it, so a held sow and its review are findable
// from each other with nothing in between to drift.
func (store *ReviewStore) Route(ctx context.Context, reviewCase application.ReviewCase) (string, error) {
	review, err := domain.Route(reviewCase.ID, string(reviewCase.Reason), store.now().UTC())
	if err != nil {
		return "", err
	}
	document := reviewDocument{
		ID: review.ID(), Reason: review.Reason(), Status: string(review.Status()),
		Text: reviewCase.Input.Text, LocaleTag: reviewCase.Input.LocaleTag,
		LocaleVersion: reviewCase.Input.LocaleVersion, RoutedAt: review.RoutedAt(),
	}
	for _, item := range reviewCase.Input.Media {
		document.Media = append(document.Media, mediaDocument{
			MIME: item.MIME, Bytes: item.Bytes, DurationMs: item.DurationMs,
		})
	}
	if reviewCase.Advisory != nil {
		document.AdvisoryStatus = string(reviewCase.Advisory.Status)
		document.AdvisoryConfidence = reviewCase.Advisory.Confidence
		for _, reason := range reviewCase.Advisory.Reasons {
			document.AdvisoryReasons = append(document.AdvisoryReasons, string(reason))
		}
	}
	if _, err := store.reviews().InsertOne(ctx, document); err != nil {
		// A retried screening pass routes the same review id again. That is
		// the same review, not a second one.
		if mongo.IsDuplicateKeyError(err) {
			return review.ID(), nil
		}
		return "", err
	}
	return review.ID(), nil
}

// Pending is the queue a reviewer works, oldest first.
type Pending struct {
	Review          domain.Review
	Text            string
	LocaleTag       string
	Media           []application.MediaMetadata
	AdvisoryStatus  string
	AdvisoryReasons []string
}

// The actor is required rather than optional: the queue cannot be read
// anonymously, because a record of who saw a member's words is only worth
// anything if there is no way to read them without leaving one.
func (store *ReviewStore) Pending(ctx context.Context, actorID string, limit int) ([]Pending, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, ErrActorRequired
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	cursor, err := store.reviews().Find(ctx,
		bson.M{"status": string(domain.StatusPending)},
		options.Find().SetSort(bson.D{{Key: "routedAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	pending := make([]Pending, 0, limit)
	for cursor.Next(ctx) {
		var document reviewDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		pending = append(pending, toPending(document))
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	// Written after the read succeeded and before the words are handed back,
	// so the log cannot claim an access that did not happen and the caller
	// cannot receive one that was never logged.
	if _, err := store.accessLog().InsertOne(ctx, accessDocument{
		ActorID: actorID, Reviewed: len(pending), ReadAt: store.now().UTC(),
	}); err != nil {
		return nil, err
	}
	return pending, nil
}

// Find returns one review by its reference.
func (store *ReviewStore) Find(ctx context.Context, reference string) (Pending, error) {
	var document reviewDocument
	err := store.reviews().FindOne(ctx, bson.M{"_id": reference}).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Pending{}, ErrReviewNotFound
	}
	if err != nil {
		return Pending{}, err
	}
	return toPending(document), nil
}

// Decide records the reviewer's judgement.
//
// The update is guarded on the review still being pending, so two reviewers
// deciding at once cannot both write — the second is told the review is
// already decided rather than quietly overwriting the first.
func (store *ReviewStore) Decide(ctx context.Context, reference string, status domain.Status, actorID, commandID string) (domain.Review, error) {
	current, err := store.Find(ctx, reference)
	if err != nil {
		return domain.Review{}, err
	}
	decided, err := current.Review.Decide(status, actorID, commandID, store.now().UTC())
	if err != nil {
		return domain.Review{}, err
	}
	result, err := store.reviews().UpdateOne(ctx,
		bson.M{"_id": reference, "status": string(domain.StatusPending)},
		bson.M{"$set": bson.M{
			"status":    string(decided.Status()),
			"decidedAt": decided.DecidedAt(),
			"decidedBy": decided.DecidedBy(),
			"commandId": decided.CommandID(),
		}})
	if err != nil {
		return domain.Review{}, err
	}
	if result.MatchedCount == 0 {
		return domain.Review{}, domain.ErrNotPending
	}
	return decided, nil
}

func toPending(document reviewDocument) Pending {
	media := make([]application.MediaMetadata, 0, len(document.Media))
	for _, item := range document.Media {
		media = append(media, application.MediaMetadata{
			MIME: item.MIME, Bytes: item.Bytes, DurationMs: item.DurationMs,
		})
	}
	return Pending{
		Review: domain.Reconstitute(document.ID, document.Reason, domain.Status(document.Status),
			document.RoutedAt, document.DecidedAt, document.DecidedBy, document.CommandID),
		Text:            document.Text,
		LocaleTag:       document.LocaleTag,
		Media:           media,
		AdvisoryStatus:  document.AdvisoryStatus,
		AdvisoryReasons: document.AdvisoryReasons,
	}
}
