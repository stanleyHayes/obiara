package domain

import (
	"errors"
	"sort"
	"strings"
	"time"
)

var (
	ErrInvalid         = errors.New("invalid sprout")
	ErrNotParticipant  = errors.New("actor is not a doorway participant")
	ErrOutOfTurn       = errors.New("doorway exchange is out of turn")
	ErrDoorwaySealed   = errors.New("doorway has completed exactly three exchanges")
	ErrCommandMismatch = errors.New("command id reused with different input")
)

const MaxExchanges = 3

type Intent struct {
	ID, ActorKey, TargetKey, SeedKey, CommandID, Fingerprint string
	At                                                       time.Time
}

func NewIntent(id, actorKey, targetKey, seedKey, commandID, fingerprint string, at time.Time) (Intent, error) {
	if id == "" || actorKey == "" || targetKey == "" || actorKey == targetKey || seedKey == "" || commandID == "" || fingerprint == "" {
		return Intent{}, ErrInvalid
	}
	return Intent{id, actorKey, targetKey, seedKey, commandID, fingerprint, at.UTC()}, nil
}

type Exchange struct {
	Number      int
	ActorKey    string
	MessageKey  string
	CommandID   string
	Fingerprint string
	At          time.Time
}

type Doorway struct {
	id           string
	participants [2]string
	nextActorKey string
	exchanges    []Exchange
	revision     uint64
	openedAt     time.Time
	sealedAt     time.Time
}

func Open(id, firstActorKey, secondActorKey string, at time.Time) (Doorway, error) {
	if id == "" || firstActorKey == "" || secondActorKey == "" || firstActorKey == secondActorKey {
		return Doorway{}, ErrInvalid
	}
	participants := []string{firstActorKey, secondActorKey}
	sort.Strings(participants)
	return Doorway{id: id, participants: [2]string{participants[0], participants[1]}, nextActorKey: firstActorKey, revision: 1, openedAt: at.UTC()}, nil
}

func Rehydrate(id string, participants [2]string, next string, exchanges []Exchange, revision uint64, openedAt, sealedAt time.Time) (Doorway, error) {
	if id == "" || participants[0] == "" || participants[1] == "" || participants[0] >= participants[1] || revision != uint64(len(exchanges)+1) || len(exchanges) > MaxExchanges {
		return Doorway{}, ErrInvalid
	}
	copyExchanges := append([]Exchange(nil), exchanges...)
	for i, e := range copyExchanges {
		if e.Number != i+1 || e.ActorKey != participants[i%2] && e.ActorKey != participants[(i+1)%2] || e.MessageKey == "" {
			return Doorway{}, ErrInvalid
		}
		if i > 0 && e.ActorKey == copyExchanges[i-1].ActorKey {
			return Doorway{}, ErrInvalid
		}
	}
	if len(copyExchanges) == MaxExchanges && sealedAt.IsZero() {
		return Doorway{}, ErrInvalid
	}
	return Doorway{id, participants, next, copyExchanges, revision, openedAt.UTC(), sealedAt.UTC()}, nil
}

func (d Doorway) Exchange(actorKey, messageKey, commandID, fingerprint string, at time.Time) (Doorway, bool, error) {
	for _, exchange := range d.exchanges {
		if exchange.CommandID == commandID {
			if exchange.Fingerprint != fingerprint {
				return d, false, ErrCommandMismatch
			}
			return d, false, nil
		}
	}
	if !d.sealedAt.IsZero() || len(d.exchanges) >= MaxExchanges {
		return d, false, ErrDoorwaySealed
	}
	if actorKey != d.participants[0] && actorKey != d.participants[1] {
		return d, false, ErrNotParticipant
	}
	if actorKey != d.nextActorKey {
		return d, false, ErrOutOfTurn
	}
	if strings.TrimSpace(messageKey) == "" || commandID == "" || fingerprint == "" {
		return d, false, ErrInvalid
	}
	next := d
	next.exchanges = append([]Exchange(nil), d.exchanges...)
	next.exchanges = append(next.exchanges, Exchange{len(next.exchanges) + 1, actorKey, messageKey, commandID, fingerprint, at.UTC()})
	next.revision++
	if len(next.exchanges) == MaxExchanges {
		next.nextActorKey = ""
		next.sealedAt = at.UTC()
	} else if actorKey == next.participants[0] {
		next.nextActorKey = next.participants[1]
	} else {
		next.nextActorKey = next.participants[0]
	}
	return next, true, nil
}

func (d Doorway) ID() string              { return d.id }
func (d Doorway) Participants() [2]string { return d.participants }
func (d Doorway) NextActorKey() string    { return d.nextActorKey }
func (d Doorway) Exchanges() []Exchange   { return append([]Exchange(nil), d.exchanges...) }
func (d Doorway) Revision() uint64        { return d.revision }
func (d Doorway) OpenedAt() time.Time     { return d.openedAt }
func (d Doorway) SealedAt() time.Time     { return d.sealedAt }
func (d Doorway) Sealed() bool            { return !d.sealedAt.IsZero() }
