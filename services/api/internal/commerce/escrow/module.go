package escrow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/application"
)

type Module struct {
	Escrows application.Service
}

type ids struct{}

func (ids) NewID() string {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}

type clock struct{}

func (clock) Now() time.Time { return time.Now().UTC() }

func NewModule(ctx context.Context, database *mongo.Database, settlementSecret []byte) (Module, error) {
	repository, err := mongodb.NewWithSettlementSecret(database, settlementSecret)
	if err != nil {
		return Module{}, err
	}
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{Escrows: application.New(repository, ids{}, clock{})}, nil
}
