package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalid           = errors.New("invalid sow")
	ErrNotConfirmed      = errors.New("deliberate confirmation is required")
	ErrScreeningRejected = errors.New("sow failed media screening")
	ErrCommandMismatch   = errors.New("command id reused with different input")
)

type Media struct {
	Key          string
	ScreeningKey string
}

type Sow struct {
	ID             string
	ActorKey       string
	Body           string
	Media          []Media
	CommandID      string
	Fingerprint    string
	AllowanceUnits int64
	AcceptedAt     time.Time
}

func Accept(id, actorKey, body string, media []Media, commandID, fingerprint string, units int64, at time.Time) (Sow, error) {
	if strings.TrimSpace(id) == "" || actorKey == "" || strings.TrimSpace(body) == "" || commandID == "" || fingerprint == "" || units <= 0 {
		return Sow{}, ErrInvalid
	}
	if len(media) > 4 {
		return Sow{}, ErrInvalid
	}
	for _, item := range media {
		if item.Key == "" || item.ScreeningKey == "" {
			return Sow{}, ErrInvalid
		}
	}
	return Sow{ID: id, ActorKey: actorKey, Body: strings.TrimSpace(body), Media: append([]Media(nil), media...), CommandID: commandID, Fingerprint: fingerprint, AllowanceUnits: units, AcceptedAt: at.UTC()}, nil
}
