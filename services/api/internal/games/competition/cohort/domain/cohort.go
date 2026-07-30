package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"
)

type Status string

const (
	StatusOpen    Status = "open"
	StatusLocked  Status = "locked"
	StatusStarted Status = "started"
)

var (
	ErrInvalid    = errors.New("invalid competition cohort")
	ErrTransition = errors.New("invalid competition cohort transition")
	ErrStale      = errors.New("stale competition cohort")
)

var (
	idPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

type Command struct {
	ID               string
	ExpectedRevision uint64
	At               time.Time
}

type Applied struct {
	ID       string `bson:"id"`
	Action   string `bson:"action"`
	Revision uint64 `bson:"revision"`
}

type Cohort struct {
	id          string
	capacity    int
	memberKeys  []string
	status      Status
	competition string
	revision    uint64
	commands    []Applied
}

type State struct {
	ID            string
	Capacity      int
	MemberKeys    []string
	Status        Status
	CompetitionID string
	Revision      uint64
	Commands      []Applied
}

func Create(id string, capacity int, command Command) (Cohort, error) {
	cohort := Cohort{id: strings.TrimSpace(id), capacity: capacity, status: StatusOpen}
	if !idPattern.MatchString(cohort.id) || !validCapacity(capacity) {
		return Cohort{}, ErrInvalid
	}
	if err := cohort.apply(command, "create"); err != nil {
		return Cohort{}, err
	}
	return cohort, nil
}

func Rehydrate(state State) (Cohort, error) {
	cohort := Cohort{
		id: state.ID, capacity: state.Capacity,
		memberKeys: append([]string(nil), state.MemberKeys...),
		status:     state.Status, competition: state.CompetitionID,
		revision: state.Revision, commands: append([]Applied(nil), state.Commands...),
	}
	if !idPattern.MatchString(cohort.id) || !validCapacity(cohort.capacity) ||
		!slices.IsSorted(cohort.memberKeys) || len(cohort.memberKeys) > cohort.capacity ||
		len(cohort.commands) != int(cohort.revision) || cohort.revision == 0 ||
		(cohort.status != StatusOpen && cohort.status != StatusLocked && cohort.status != StatusStarted) ||
		(cohort.status != StatusOpen && len(cohort.memberKeys) != cohort.capacity) ||
		(cohort.status == StatusStarted && !idPattern.MatchString(cohort.competition)) {
		return Cohort{}, ErrInvalid
	}
	for index, key := range cohort.memberKeys {
		if !keyPattern.MatchString(key) || (index > 0 && cohort.memberKeys[index-1] == key) {
			return Cohort{}, ErrInvalid
		}
	}
	return cohort, nil
}

func (cohort Cohort) Join(memberKey string, command Command) (Cohort, error) {
	if cohort.status != StatusOpen || !keyPattern.MatchString(memberKey) ||
		slices.Contains(cohort.memberKeys, memberKey) || len(cohort.memberKeys) >= cohort.capacity {
		return Cohort{}, ErrTransition
	}
	next := cohort.clone()
	next.memberKeys = append(next.memberKeys, memberKey)
	slices.Sort(next.memberKeys)
	if len(next.memberKeys) == next.capacity {
		next.status = StatusLocked
	}
	if err := next.apply(command, "join"); err != nil {
		return Cohort{}, err
	}
	return next, nil
}

func (cohort Cohort) Leave(memberKey string, command Command) (Cohort, error) {
	index := slices.Index(cohort.memberKeys, memberKey)
	if cohort.status != StatusOpen || index < 0 {
		return Cohort{}, ErrTransition
	}
	next := cohort.clone()
	next.memberKeys = append(next.memberKeys[:index], next.memberKeys[index+1:]...)
	if err := next.apply(command, "leave"); err != nil {
		return Cohort{}, err
	}
	return next, nil
}

func (cohort Cohort) Start(competitionID string, command Command) (Cohort, error) {
	if cohort.status != StatusLocked || !idPattern.MatchString(competitionID) {
		return Cohort{}, ErrTransition
	}
	next := cohort.clone()
	next.status, next.competition = StatusStarted, competitionID
	if err := next.apply(command, "start"); err != nil {
		return Cohort{}, err
	}
	return next, nil
}

func (cohort *Cohort) apply(command Command, action string) error {
	if !idPattern.MatchString(strings.TrimSpace(command.ID)) || command.At.IsZero() ||
		command.ExpectedRevision != cohort.revision {
		return ErrStale
	}
	cohort.revision++
	cohort.commands = append(cohort.commands, Applied{
		ID: strings.TrimSpace(command.ID), Action: action, Revision: cohort.revision,
	})
	return nil
}

func (cohort Cohort) clone() Cohort {
	cohort.memberKeys = append([]string(nil), cohort.memberKeys...)
	cohort.commands = append([]Applied(nil), cohort.commands...)
	return cohort
}

func validCapacity(capacity int) bool {
	return capacity >= 4 && capacity <= 16 && capacity&(capacity-1) == 0
}

func (cohort Cohort) ID() string            { return cohort.id }
func (cohort Cohort) Capacity() int         { return cohort.capacity }
func (cohort Cohort) MemberKeys() []string  { return append([]string(nil), cohort.memberKeys...) }
func (cohort Cohort) Status() Status        { return cohort.status }
func (cohort Cohort) CompetitionID() string { return cohort.competition }
func (cohort Cohort) Revision() uint64      { return cohort.revision }
func (cohort Cohort) Commands() []Applied   { return append([]Applied(nil), cohort.commands...) }
