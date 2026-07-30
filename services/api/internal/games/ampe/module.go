package ampe

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/adapters/outbound/privacy"
)

type Module struct {
	Rounds   application.Service
	Presence Presence
}

func NewModule(
	ctx context.Context,
	database *mongo.Database,
	secret string,
	authority application.Authority,
) (Module, error) {
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.NewKeyer([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	presenceRepository := newPresenceRepository(database)
	if err := presenceRepository.ensureIndexes(ctx); err != nil {
		return Module{}, err
	}
	rounds := application.NewService(repository, authority, keyer, idSource{})
	return Module{
		Rounds: rounds,
		Presence: Presence{
			rounds: rounds, repository: presenceRepository, keyer: keyer, now: time.Now,
		},
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "ampe_" + base64.RawURLEncoding.EncodeToString(value)
}
