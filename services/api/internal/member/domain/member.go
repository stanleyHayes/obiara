package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidID    = errors.New("member id is required")
	ErrInvalidEmail = errors.New("member email is required")
)

// Member is the domain representation of an Obiara member.
// Transport and persistence concerns deliberately stay outside this package.
type Member struct {
	id        string
	email     string
	createdAt time.Time
}

func NewMember(id, email string, createdAt time.Time) (Member, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Member{}, ErrInvalidID
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return Member{}, ErrInvalidEmail
	}

	return Member{id: id, email: email, createdAt: createdAt.UTC()}, nil
}

func (m Member) ID() string           { return m.id }
func (m Member) Email() string        { return m.email }
func (m Member) CreatedAt() time.Time { return m.createdAt }
