// Package mongodb persists admin principals, MFA challenges, sessions and
// the immutable admin-access audit (FR-801).
package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/application"
	"github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

// PrincipalRepository persists admin principals.
type PrincipalRepository struct {
	database *mongo.Database
}

func NewPrincipalRepository(database *mongo.Database) *PrincipalRepository {
	return &PrincipalRepository{database: database}
}

// ChallengeRepository persists MFA challenges.
type ChallengeRepository struct {
	database *mongo.Database
}

func NewChallengeRepository(database *mongo.Database) *ChallengeRepository {
	return &ChallengeRepository{database: database}
}

// SessionRepository persists admin sessions.
type SessionRepository struct {
	database *mongo.Database
}

func NewSessionRepository(database *mongo.Database) *SessionRepository {
	return &SessionRepository{database: database}
}

// AuditStore persists the immutable admin-access audit.
type AuditStore struct {
	database *mongo.Database
}

func NewAuditStore(database *mongo.Database) *AuditStore {
	return &AuditStore{database: database}
}

// Repository bundles the per-aggregate repositories for shared index setup.
type Repository = PrincipalRepository

func (repository *PrincipalRepository) principals() *mongo.Collection {
	return repository.database.Collection("admin_principals")
}

func (repository *ChallengeRepository) challenges() *mongo.Collection {
	return repository.database.Collection("admin_mfa_challenges")
}

func (repository *SessionRepository) sessions() *mongo.Collection {
	return repository.database.Collection("admin_sessions")
}

func (repository *AuditStore) audit() *mongo.Collection {
	return repository.database.Collection("admin_access")
}

type principalDocument struct {
	ID        string    `bson:"_id"`
	Email     string    `bson:"email"`
	Roles     []string  `bson:"roles"`
	Status    string    `bson:"status"`
	Version   int64     `bson:"version"`
	CreatedAt time.Time `bson:"createdAt"`
}

type challengeDocument struct {
	ID          string     `bson:"_id"`
	PrincipalID string     `bson:"principalId"`
	CodeHash    string     `bson:"codeHash"`
	ExpiresAt   time.Time  `bson:"expiresAt"`
	Attempts    int        `bson:"attempts"`
	CreatedAt   time.Time  `bson:"createdAt"`
	ConsumedAt  *time.Time `bson:"consumedAt,omitempty"`
}

type sessionDocument struct {
	ID          string    `bson:"_id"`
	PrincipalID string    `bson:"principalId"`
	Roles       []string  `bson:"roles"`
	SteppedUp   bool      `bson:"steppedUp"`
	ExpiresAt   time.Time `bson:"expiresAt"`
	Revoked     bool      `bson:"revoked"`
	Version     int64     `bson:"version"`
	CreatedAt   time.Time `bson:"createdAt"`
}

func (repository *PrincipalRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.principals().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetName("admin_principals_email_unique").SetUnique(true),
	})
	return err
}

// EnsureIndexes declares the challenge indexes.
func (repository *ChallengeRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.challenges().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "principalId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("admin_mfa_principal_latest"),
		},
		{
			Keys:    bson.D{{Key: "createdAt", Value: 1}},
			Options: options.Index().SetName("admin_mfa_ttl").SetExpireAfterSeconds(3600),
		},
	})
	return err
}

// EnsureIndexes declares the session indexes.
func (repository *SessionRepository) EnsureIndexes(ctx context.Context) error {
	_, err := repository.sessions().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "principalId", Value: 1}, {Key: "createdAt", Value: -1}},
		Options: options.Index().SetName("admin_sessions_principal"),
	})
	return err
}

func (repository *PrincipalRepository) Create(ctx context.Context, principal domain.Principal) error {
	_, err := repository.principals().InsertOne(ctx, toPrincipalDocument(principal))
	if apimongo.IsDuplicateKey(err) {
		return application.ErrPrincipalExists
	}
	return err
}

func (repository *PrincipalRepository) FindByEmail(ctx context.Context, email string) (domain.Principal, error) {
	var document principalDocument
	if err := repository.principals().FindOne(ctx, bson.M{"email": email}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Principal{}, application.ErrPrincipalNotFound
		}
		return domain.Principal{}, err
	}
	return toPrincipalDomain(document), nil
}

func (repository *PrincipalRepository) FindByID(ctx context.Context, id string) (domain.Principal, error) {
	var document principalDocument
	if err := repository.principals().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Principal{}, application.ErrPrincipalNotFound
		}
		return domain.Principal{}, err
	}
	return toPrincipalDomain(document), nil
}

func (repository *ChallengeRepository) Create(ctx context.Context, challenge domain.Challenge) error {
	_, err := repository.challenges().InsertOne(ctx, challengeDocument{
		ID:          challenge.ID(),
		PrincipalID: challenge.PrincipalID(),
		CodeHash:    challenge.CodeHash(),
		ExpiresAt:   challenge.ExpiresAt(),
		Attempts:    challenge.Attempts(),
		CreatedAt:   challenge.CreatedAt(),
		ConsumedAt:  challenge.ConsumedAt(),
	})
	return err
}

func (repository *ChallengeRepository) LatestForPrincipal(ctx context.Context, principalID string) (domain.Challenge, error) {
	var document challengeDocument
	err := repository.challenges().FindOne(ctx,
		bson.M{"principalId": principalID},
		options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	).Decode(&document)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Challenge{}, application.ErrChallengeNotFound
		}
		return domain.Challenge{}, err
	}
	return domain.ReconstituteChallenge(document.ID, document.PrincipalID, document.CodeHash, document.ExpiresAt, document.Attempts, document.CreatedAt, document.ConsumedAt), nil
}

func (repository *ChallengeRepository) Update(ctx context.Context, challenge domain.Challenge) error {
	_, err := repository.challenges().UpdateOne(ctx,
		bson.M{"_id": challenge.ID()},
		bson.M{"$set": bson.M{"attempts": challenge.Attempts(), "consumedAt": challenge.ConsumedAt()}})
	return err
}

func (repository *SessionRepository) Create(ctx context.Context, session domain.Session) error {
	_, err := repository.sessions().InsertOne(ctx, toSessionDocument(session))
	return err
}

func (repository *SessionRepository) FindByID(ctx context.Context, id string) (domain.Session, error) {
	var document sessionDocument
	if err := repository.sessions().FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Session{}, application.ErrSessionNotFound
		}
		return domain.Session{}, err
	}
	return toSessionDomain(document), nil
}

func (repository *SessionRepository) Update(ctx context.Context, session domain.Session) error {
	result, err := repository.sessions().UpdateOne(ctx,
		bson.M{"_id": session.ID(), "version": session.Version() - 1},
		bson.M{"$set": bson.M{"steppedUp": session.SteppedUp(), "revoked": session.Revoked(), "version": session.Version()}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return application.ErrSessionNotFound
	}
	return nil
}

func (repository *AuditStore) Append(ctx context.Context, actorID, action, target string, at time.Time) error {
	_, err := repository.audit().InsertOne(ctx, bson.M{
		"actorId": actorID, "action": action, "target": target, "at": at.UTC(),
	})
	return err
}

// CountAudit supports admin access reviews.
func (repository *AuditStore) CountFor(ctx context.Context, filter bson.M) (int, error) {
	count, err := repository.audit().CountDocuments(ctx, filter)
	return int(count), err
}

func toPrincipalDocument(principal domain.Principal) principalDocument {
	roles := make([]string, 0, len(principal.Roles()))
	for _, role := range principal.Roles() {
		roles = append(roles, string(role))
	}
	return principalDocument{
		ID: principal.ID(), Email: principal.Email(), Roles: roles,
		Status: string(principal.Status()), Version: principal.Version(), CreatedAt: principal.CreatedAt(),
	}
}

func toPrincipalDomain(document principalDocument) domain.Principal {
	roles := make([]domain.Role, 0, len(document.Roles))
	for _, role := range document.Roles {
		roles = append(roles, domain.Role(role))
	}
	return domain.ReconstitutePrincipal(document.ID, document.Email, roles, domain.Status(document.Status), document.Version, document.CreatedAt)
}

func toSessionDocument(session domain.Session) sessionDocument {
	roles := make([]string, 0, len(session.Roles()))
	for _, role := range session.Roles() {
		roles = append(roles, string(role))
	}
	return sessionDocument{
		ID: session.ID(), PrincipalID: session.PrincipalID(), Roles: roles,
		SteppedUp: session.SteppedUp(), ExpiresAt: session.ExpiresAt(),
		Revoked: session.Revoked(), Version: session.Version(), CreatedAt: session.CreatedAt(),
	}
}

func toSessionDomain(document sessionDocument) domain.Session {
	roles := make([]domain.Role, 0, len(document.Roles))
	for _, role := range document.Roles {
		roles = append(roles, domain.Role(role))
	}
	return domain.ReconstituteSession(document.ID, document.PrincipalID, roles, document.SteppedUp, document.ExpiresAt, document.Revoked, document.Version, document.CreatedAt)
}
