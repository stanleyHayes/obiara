package domain

import (
	"errors"
	"regexp"
	"time"
)

type Operation string

const (
	OperationSow       Operation = "sow"
	OperationCandidate Operation = "candidate"
)

var (
	ErrInvalidRequest = errors.New("invalid seed safety request")
	ErrThrottled      = errors.New("seed action unavailable")
	ErrStaleRevision  = errors.New("stale seed safety revision")
)

var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Bucket struct {
	ActorKey      string
	WindowStarted time.Time
	Sows          int
	Candidates    int
	Denials       int
	Revision      uint64
}

type Decision struct {
	Allowed    bool
	RetryAfter time.Time
	CareSignal bool
}

func New(actorKey string, at time.Time) (Bucket, error) {
	if !keyPattern.MatchString(actorKey) || at.IsZero() {
		return Bucket{}, ErrInvalidRequest
	}
	return Bucket{ActorKey: actorKey, WindowStarted: window(at), Revision: 1}, nil
}

func Rehydrate(bucket Bucket) (Bucket, error) {
	if !keyPattern.MatchString(bucket.ActorKey) || bucket.WindowStarted.IsZero() ||
		bucket.Sows < 0 || bucket.Candidates < 0 || bucket.Denials < 0 || bucket.Revision == 0 {
		return Bucket{}, ErrInvalidRequest
	}
	bucket.WindowStarted = bucket.WindowStarted.UTC()
	return bucket, nil
}

func (bucket Bucket) Evaluate(operation Operation, expected uint64, at time.Time) (Bucket, Decision, error) {
	if expected != bucket.Revision {
		return Bucket{}, Decision{}, ErrStaleRevision
	}
	if at.IsZero() || at.Before(bucket.WindowStarted) || (operation != OperationSow && operation != OperationCandidate) {
		return Bucket{}, Decision{}, ErrInvalidRequest
	}
	currentWindow := window(at)
	if currentWindow.After(bucket.WindowStarted) {
		bucket.WindowStarted = currentWindow
		bucket.Sows, bucket.Candidates, bucket.Denials = 0, 0, 0
	}
	allowed := operation == OperationSow && bucket.Sows < 6 ||
		operation == OperationCandidate && bucket.Candidates < 30
	if allowed {
		if operation == OperationSow {
			bucket.Sows++
		} else {
			bucket.Candidates++
		}
	} else {
		bucket.Denials++
	}
	bucket.Revision++
	return bucket, Decision{
		Allowed: allowed, RetryAfter: bucket.WindowStarted.Add(10 * time.Minute),
		CareSignal: !allowed && bucket.Denials >= 3,
	}, nil
}

func window(at time.Time) time.Time {
	at = at.UTC()
	return at.Truncate(10 * time.Minute)
}
