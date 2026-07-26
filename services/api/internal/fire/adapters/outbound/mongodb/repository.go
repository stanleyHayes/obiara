// Package mongodb persists fires and attendance with transactional
// capacity semantics (agent_plan.md §7.4: counter and record commit
// together; optimistic concurrency pins every counter move).
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/domain"
)

type Repository struct {
	database *mongo.Database
}

func NewRepository(database *mongo.Database) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) fires() *mongo.Collection {
	return repository.database.Collection("fires")
}

func (repository *Repository) attendance() *mongo.Collection {
	return repository.database.Collection("fire_attendance")
}

type fireDocument struct {
	ID            string    `bson:"_id"`
	HostID        string    `bson:"hostId"`
	CircleID      string    `bson:"circleId,omitempty"`
	Title         string    `bson:"title"`
	StartsAt      time.Time `bson:"startsAt"`
	Capacity      int       `bson:"capacity"`
	GoingCount    int       `bson:"goingCount"`
	WaitlistCount int       `bson:"waitlistCount"`
	Status        string    `bson:"status"`
	Version       int64     `bson:"version"`
	CreatedAt     time.Time `bson:"createdAt"`
}

type rsvpDocument struct {
	ID        string    `bson:"_id"`
	FireID    string    `bson:"fireId"`
	MemberID  string    `bson:"memberId"`
	Status    string    `bson:"status"`
	Position  int       `bson:"position"`
	Version   int64     `bson:"version"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (repository *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := repository.fires().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "status", Value: 1}, {Key: "startsAt", Value: 1}},
		Options: options.Index().SetName("fires_upcoming"),
	}); err != nil {
		return err
	}
	_, err := repository.attendance().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "fireId", Value: 1}, {Key: "memberId", Value: 1}},
			Options: options.Index().SetName("attendance_fire_member_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "fireId", Value: 1}, {Key: "status", Value: 1}, {Key: "position", Value: 1}},
			Options: options.Index().SetName("attendance_waitlist"),
		},
	})
	return err
}

func (repository *Repository) Create(ctx context.Context, fire domain.Fire) error {
	_, err := repository.fires().InsertOne(ctx, toFireDocument(fire))
	return err
}

func (repository *Repository) FindByID(ctx context.Context, id string) (domain.Fire, error) {
	var document fireDocument
	if err := repository.fires().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Fire{}, application.ErrFireNotFound
		}
		return domain.Fire{}, err
	}
	return toFireDomain(document), nil
}

// AdmitTx claims a seat or waitlist position atomically. The seat claim is
// a single conditional increment (goingCount < capacity, encoded with
// $expr) so concurrent admissions cannot overshoot capacity; the waitlist
// position comes from the waitlistCount counter incremented in the same
// transaction as the attendance insert. Transient transaction conflicts
// are retried by the driver.
func (repository *Repository) AdmitTx(ctx context.Context, fireID, memberID string, now time.Time) (domain.RSVP, error) {
	for attempt := 0; attempt < 8; attempt++ {
		admitted, err := repository.admitOnce(ctx, fireID, memberID, now)
		if errors.Is(err, errWaitlistRace) {
			continue
		}
		return admitted, err
	}
	return domain.RSVP{}, errWaitlistRace
}

func (repository *Repository) admitOnce(ctx context.Context, fireID, memberID string, now time.Time) (domain.RSVP, error) {
	var admitted domain.RSVP
	err := apimongo.WithTransaction(ctx, repository.database.Client(), func(sc context.Context) error {
		open := bson.M{"$in": bson.A{string(domain.StatusScheduled), string(domain.StatusLive)}}

		// Atomic seat claim.
		claimed, err := repository.fires().UpdateOne(sc,
			bson.M{
				"_id":    fireID,
				"status": open,
				"$expr":  bson.M{"$lt": bson.A{"$goingCount", "$capacity"}},
			},
			bson.M{"$inc": bson.M{"goingCount": 1, "version": 1}})
		if err != nil {
			return err
		}
		if claimed.MatchedCount == 1 {
			admitted = domain.ReconstituteRSVP(fireID, memberID, domain.RSVPGoing, 0, 1, now.UTC())
			return repository.insertRSVP(sc, admitted)
		}

		// No seat: confirm the fire is open before waitlisting (the claim
		// also misses closed/ended fires).
		var fire fireDocument
		if err := repository.fires().FindOne(sc, bson.M{"_id": fireID, "status": open}).Decode(&fire); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return domain.ErrFireNotOpen
			}
			return err
		}
		bumped, err := repository.fires().UpdateOne(sc,
			bson.M{"_id": fireID, "version": fire.Version},
			bson.M{"$inc": bson.M{"waitlistCount": 1, "version": 1}})
		if err != nil {
			return err
		}
		if bumped.MatchedCount == 0 {
			return errWaitlistRace
		}
		admitted = domain.ReconstituteRSVP(fireID, memberID, domain.RSVPWaitlisted, fire.WaitlistCount+1, 1, now.UTC())
		return repository.insertRSVP(sc, admitted)
	})
	if errors.Is(err, errWaitlistRace) {
		return domain.RSVP{}, errWaitlistRace
	}
	return admitted, err
}

var errWaitlistRace = errors.New("waitlist counter changed concurrently")

func (repository *Repository) insertRSVP(ctx context.Context, rsvp domain.RSVP) error {
	_, err := repository.attendance().InsertOne(ctx, toRSVPDocument(rsvp))
	if apimongo.IsDuplicateKey(err) {
		return application.ErrAlreadyRSVPed
	}
	return err
}

// CancelTx removes the RSVP and promotes the first waitlisted member when
// a going seat freed, all in one transaction.
func (repository *Repository) CancelTx(ctx context.Context, fireID, memberID string, now time.Time) (*domain.RSVP, error) {
	var promoted *domain.RSVP
	err := apimongo.WithTransaction(ctx, repository.database.Client(), func(sc context.Context) error {
		var existing rsvpDocument
		if err := repository.attendance().FindOne(sc, bson.M{"_id": rsvpKey(fireID, memberID)}).Decode(&existing); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return application.ErrRSVPNotFound
			}
			return err
		}
		if _, err := repository.attendance().DeleteOne(sc, bson.M{"_id": existing.ID}); err != nil {
			return err
		}
		if existing.Status != string(domain.RSVPGoing) {
			return nil // waitlisted cancellation frees no seat
		}

		var next rsvpDocument
		err := repository.attendance().FindOne(sc,
			bson.M{"fireId": fireID, "status": string(domain.RSVPWaitlisted)},
			options.FindOne().SetSort(bson.D{{Key: "position", Value: 1}}),
		).Decode(&next)
		if errors.Is(err, mongo.ErrNoDocuments) {
			// No waitlist: the seat simply frees.
			_, err = repository.fires().UpdateOne(sc,
				bson.M{"_id": fireID},
				bson.M{"$inc": bson.M{"goingCount": -1, "version": 1}})
			return err
		}
		if err != nil {
			return err
		}
		promotedRSVP := domain.ReconstituteRSVP(fireID, next.MemberID, domain.RSVPGoing, 0, next.Version+1, next.CreatedAt)
		if _, err := repository.attendance().UpdateOne(sc,
			bson.M{"_id": next.ID, "version": next.Version},
			bson.M{"$set": bson.M{"status": string(domain.RSVPGoing), "position": 0, "version": next.Version + 1}}); err != nil {
			return err
		}
		if _, err := repository.fires().UpdateOne(sc,
			bson.M{"_id": fireID},
			bson.M{"$inc": bson.M{"version": 1}}); err != nil {
			return err
		}
		promoted = &promotedRSVP
		return nil
	})
	return promoted, err
}

func (repository *Repository) FindRSVP(ctx context.Context, fireID, memberID string) (domain.RSVP, error) {
	var document rsvpDocument
	if err := repository.attendance().FindOne(ctx, bson.M{"_id": rsvpKey(fireID, memberID)}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.RSVP{}, application.ErrRSVPNotFound
		}
		return domain.RSVP{}, err
	}
	return toRSVPDomain(document), nil
}

func (repository *Repository) ListUpcoming(ctx context.Context, now time.Time, limit int) ([]domain.Fire, error) {
	if limit < 1 {
		limit = 20
	}
	cursor, err := repository.fires().Find(ctx,
		bson.M{"status": string(domain.StatusScheduled), "startsAt": bson.M{"$gte": now.UTC()}},
		options.Find().SetSort(bson.D{{Key: "startsAt", Value: 1}}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var fires []domain.Fire
	for cursor.Next(ctx) {
		var document fireDocument
		if err := cursor.Decode(&document); err != nil {
			return nil, err
		}
		fires = append(fires, toFireDomain(document))
	}
	return fires, cursor.Err()
}

func rsvpKey(fireID, memberID string) string {
	return fireID + "|" + memberID
}

func toFireDocument(fire domain.Fire) fireDocument {
	return fireDocument{
		ID:            fire.ID(),
		HostID:        fire.HostID(),
		CircleID:      fire.CircleID(),
		Title:         fire.Title(),
		StartsAt:      fire.StartsAt(),
		Capacity:      fire.Capacity(),
		GoingCount:    fire.GoingCount(),
		WaitlistCount: 0,
		Status:        string(fire.Status()),
		Version:       fire.Version(),
		CreatedAt:     fire.CreatedAt(),
	}
}

func toFireDomain(document fireDocument) domain.Fire {
	return domain.ReconstituteFire(
		document.ID, document.HostID, document.CircleID, document.Title,
		document.StartsAt, document.Capacity, document.GoingCount,
		domain.FireStatus(document.Status), document.Version, document.CreatedAt,
	)
}

func toRSVPDocument(rsvp domain.RSVP) rsvpDocument {
	return rsvpDocument{
		ID:        rsvpKey(rsvp.FireID(), rsvp.MemberID()),
		FireID:    rsvp.FireID(),
		MemberID:  rsvp.MemberID(),
		Status:    string(rsvp.Status()),
		Position:  rsvp.Position(),
		Version:   rsvp.Version(),
		CreatedAt: rsvp.CreatedAt(),
	}
}

func toRSVPDomain(document rsvpDocument) domain.RSVP {
	return domain.ReconstituteRSVP(document.FireID, document.MemberID, domain.RSVPStatus(document.Status), document.Position, document.Version, document.CreatedAt)
}
