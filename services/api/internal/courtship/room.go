// Package courtship composes the courtship room slices: the rhythm a pair
// keeps (pace), the pause either may call (pause), the honesty ribbon,
// closure, and the in-room safety actions.
//
// They are composed together because they are one surface to a member — the
// room they are courting in — and because they share a single privacy secret.
// Every member reference is keyed before it reaches storage, so a room's
// participants are never legible at rest.
package courtship

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	closuremongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/adapters/outbound/mongodb"
	closureprivacy "github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/adapters/outbound/privacy"
	closureapp "github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/application"
	closuredomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/domain"
	honestymongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/adapters/outbound/mongodb"
	honestyprivacy "github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/adapters/outbound/privacy"
	honestyapp "github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/application"
	honestydomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/domain"
	pacemongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/adapters/outbound/mongodb"
	paceprivacy "github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/adapters/outbound/privacy"
	paceapp "github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/application"
	pacedomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/domain"
	pausemongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/adapters/outbound/mongodb"
	pauseprivacy "github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/adapters/outbound/privacy"
	pauseapp "github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/application"
	pausedomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/domain"
	safetymongo "github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/adapters/outbound/mongodb"
	safetyprivacy "github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/adapters/outbound/privacy"
	safetyapp "github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/application"
	safetydomain "github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/domain"
)

// ErrSecretRequired reports a module built without a privacy secret.
var ErrSecretRequired = errors.New("courtship room module requires a privacy secret")

// inactivityThreshold is how long a room may sit untouched before it can be
// closed for inactivity. A courtship that has gone quiet for a month has
// ended whether or not either party said so.
const inactivityThreshold = 30 * 24 * time.Hour

// RoomModule bundles the room slices.
type RoomModule struct {
	Pace    paceapp.Service
	Pause   pauseapp.Service
	Closure closureapp.Service
	Honesty honestyapp.Service
	Safety  safetyapp.Service
}

// NewRoomModule builds every room slice against one database and secret.
func NewRoomModule(ctx context.Context, database *mongo.Database, secret string) (RoomModule, error) {
	if len(secret) < 32 {
		return RoomModule{}, ErrSecretRequired
	}
	key := []byte(secret)

	paceRepository := pacemongo.NewRepository(database)
	if err := paceRepository.EnsureIndexes(ctx); err != nil {
		return RoomModule{}, err
	}
	paceKeyer, err := paceprivacy.NewKeyer(key)
	if err != nil {
		return RoomModule{}, err
	}

	pauseRepository := pausemongo.NewRepository(database)
	if err := pauseRepository.EnsureIndexes(ctx); err != nil {
		return RoomModule{}, err
	}
	pauseKeyer, err := pauseprivacy.NewKeyer(key)
	if err != nil {
		return RoomModule{}, err
	}

	closureRepository := closuremongo.NewRepository(database)
	if err := closureRepository.EnsureIndexes(ctx); err != nil {
		return RoomModule{}, err
	}
	closureKeyer, err := closureprivacy.NewKeyer(key)
	if err != nil {
		return RoomModule{}, err
	}

	honestyRepository := honestymongo.NewRepository(database)
	if err := honestyRepository.EnsureIndexes(ctx); err != nil {
		return RoomModule{}, err
	}
	honestyKeyer, err := honestyprivacy.NewKeyer(key)
	if err != nil {
		return RoomModule{}, err
	}

	safetyRepository := safetymongo.NewRepository(database)
	if err := safetyRepository.EnsureIndexes(ctx); err != nil {
		return RoomModule{}, err
	}
	safetyKeyer, err := safetyprivacy.NewKeyer(key)
	if err != nil {
		return RoomModule{}, err
	}

	return RoomModule{
		Pace:    paceapp.NewService(paceRepository, paceKeyer, time.Now),
		Pause:   pauseapp.NewService(pauseRepository, pauseKeyer, time.Now),
		Closure: closureapp.NewService(closureRepository, closureKeyer, time.Now, inactivityThreshold),
		Honesty: honestyapp.NewService(honestyRepository, honestyKeyer, time.Now),
		Safety:  safetyapp.NewService(safetyRepository, safetyKeyer, time.Now),
	}, nil
}

// notFound reports a store miss. A room that does not exist and one the
// caller is not part of must be indistinguishable, so both become the
// slice's own denial rather than reaching the transport as an internal
// fault — which would both leak that the room is absent and answer 500 to a
// perfectly ordinary request.
func notFound(err error) bool {
	return errors.Is(err, mongo.ErrNoDocuments)
}

// Room adapts the five room slices to the single inbound port the transport
// layer sees. Each slice keeps its own aggregate and store; this only spares
// the composition root from threading five services through one surface that
// a member experiences as one room.
type Room struct{ module RoomModule }

func NewRoom(module RoomModule) Room { return Room{module: module} }

// Start opens a room between two members.
//
// Each facet is its own aggregate with its own store, so opening a room means
// opening all of them. Opening only the pace left the room existing in name
// while every other action failed on a missing document.
//
// The facets are opened after the pace, and a failure part-way through is
// returned to the caller: the command id makes the whole call safe to retry,
// and a half-open room is better surfaced than hidden.
func (room Room) Start(ctx context.Context, roomID string, members []string, commandID, actorID string) (pacedomain.Pace, error) {
	pace, err := room.module.Pace.Start(ctx, roomID, members, commandID, actorID)
	if err != nil {
		return pacedomain.Pace{}, err
	}
	if _, err := room.module.Honesty.Open(ctx, roomID, members, commandID); err != nil {
		return pacedomain.Pace{}, err
	}
	if _, err := room.module.Pause.Open(ctx, roomID, members, commandID); err != nil {
		return pacedomain.Pace{}, err
	}
	if _, err := room.module.Closure.Open(ctx, roomID, members, commandID); err != nil {
		return pacedomain.Pace{}, err
	}
	if _, err := room.module.Safety.Open(ctx, roomID, members, commandID); err != nil {
		return pacedomain.Pace{}, err
	}
	return pace, nil
}

func (room Room) Advance(ctx context.Context, roomID, commandID string, expected uint64) (pacedomain.Pace, error) {
	pace, err := room.module.Pace.Advance(ctx, roomID, commandID, expected)
	if notFound(err) {
		return pacedomain.Pace{}, pacedomain.ErrDenied
	}
	return pace, err
}

func (room Room) Relight(ctx context.Context, roomID, memberID, commandID string, expected uint64) (pacedomain.Pace, error) {
	pace, err := room.module.Pace.Relight(ctx, roomID, memberID, commandID, expected)
	if notFound(err) {
		return pacedomain.Pace{}, pacedomain.ErrDenied
	}
	return pace, err
}

func (room Room) ApplyPause(ctx context.Context, roomID, memberID, commandID string, action pausedomain.Action, expected uint64) (pausedomain.Stone, error) {
	stone, err := room.module.Pause.Apply(ctx, roomID, memberID, commandID, action, expected)
	if notFound(err) {
		return pausedomain.Stone{}, pausedomain.ErrDenied
	}
	return stone, err
}

func (room Room) SetHonesty(ctx context.Context, roomID, memberID, commandID string, grant bool, expected uint64) (honestydomain.Ribbon, error) {
	ribbon, err := room.module.Honesty.Set(ctx, roomID, memberID, commandID, grant, expected)
	if notFound(err) {
		return honestydomain.Ribbon{}, honestydomain.ErrDenied
	}
	return ribbon, err
}

func (room Room) Close(ctx context.Context, roomID, memberID, commandID string, expected uint64) (closuredomain.Closure, error) {
	closure, err := room.module.Closure.Close(ctx, roomID, memberID, commandID, expected)
	if notFound(err) {
		return closuredomain.Closure{}, closuredomain.ErrDenied
	}
	return closure, err
}

func (room Room) Block(ctx context.Context, roomID, memberID, commandID string, expected uint64) (safetydomain.Safety, error) {
	safety, err := room.module.Safety.Block(ctx, roomID, memberID, commandID, expected)
	if notFound(err) {
		return safetydomain.Safety{}, safetydomain.ErrDenied
	}
	return safety, err
}

func (room Room) Report(ctx context.Context, roomID, memberID, commandID string, category safetydomain.Category, evidenceRef string, expected uint64) (safetydomain.Safety, error) {
	safety, err := room.module.Safety.Report(ctx, roomID, memberID, commandID, category, evidenceRef, expected)
	if notFound(err) {
		return safetydomain.Safety{}, safetydomain.ErrDenied
	}
	return safety, err
}
