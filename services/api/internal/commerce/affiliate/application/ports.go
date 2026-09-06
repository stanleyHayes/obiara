package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/domain"
)

var (
	ErrNotFound    = errors.New("affiliate not found")
	ErrConflict    = errors.New("affiliate changed concurrently")
	ErrCodeTaken   = errors.New("that affiliate code is already in use")
	ErrUnavailable = errors.New("affiliate service unavailable")
	// ErrMemberAffiliate refuses an affiliate that is a member.
	//
	// The owner's decision, and the one with the least engineering in it and
	// the most consequence: paying somebody inside the community to recruit
	// changes what "why is this person talking to me" means, in a product
	// whose premise is that the answer is not money (agent_plan.md §41).
	ErrMemberAffiliate = errors.New("a member cannot be an affiliate")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Affiliate) error
	FindByID(context.Context, string) (domain.Affiliate, error)
	FindByCode(context.Context, string) (domain.Affiliate, error)
	Append(ctx context.Context, affiliate domain.Affiliate, expected uint64, commandID string) error
	List(ctx context.Context, limit int) ([]domain.Affiliate, error)
}

// Referrals remembers which affiliate brought which member, and whether that
// referral has been qualified yet.
//
// Kept apart from the affiliate aggregate because a referral is pending for
// thirty days before it earns anything, and a sweep has to be able to find the
// ones that are due without loading every affiliate.
type Referrals interface {
	Record(context.Context, Referral) error
	Find(ctx context.Context, memberKey string) (Referral, error)
	// DueForQualification lists referrals whose waiting period has elapsed
	// and that have not been settled either way.
	DueForQualification(ctx context.Context, at time.Time, limit int) ([]Referral, error)
	MarkSettled(ctx context.Context, memberKey, outcome string) error
}

// Referral is one member somebody brought.
type Referral struct {
	// MemberKey is a digest. Nothing reads it back to a person: it exists so
	// one referral counts once and so a sweep can ask the qualification
	// questions about somebody without naming them.
	MemberKey string
	// MemberID is raw, because the qualification questions are asked about a
	// person thirty days later and a one-way digest cannot be asked anything.
	// It is the same reason a purchase order keeps one (agent_plan.md §75),
	// and it is safe for the same reason: no affiliate ever reads this row.
	// What an affiliate is told is a count.
	MemberID    string
	AffiliateID string
	Code        string
	RecordedAt  time.Time
	// QualifiesAt is when the waiting period ends.
	QualifiesAt time.Time
	// Settled is empty while pending, then "qualified" or "refused".
	Settled string
}

// Qualification answers the questions that decide whether a referral has
// converted: is this member verified, and are they clean.
//
// Commission never accrues on a signup. A scheme paying per signup rewards
// exactly the bulk recruitment the tier ladder, age assurance and Sentinel
// exist to slow down.
type Qualification interface {
	// Verified reports whether the member reached Tier 1 and is still there.
	Verified(ctx context.Context, memberID string) (bool, error)
	// Clean reports whether the member has no upheld safety finding.
	Clean(ctx context.Context, memberID string) (bool, error)
}

// Ledger records what is owed. Commission is a liability the moment it
// accrues, whether or not anybody has asked to be paid.
type Ledger interface {
	RecordAccrual(ctx context.Context, reference string, minor int64, at time.Time) error
	RecordClawback(ctx context.Context, reference string, minor int64, at time.Time) error
}

type Keyer interface {
	MemberKey(memberID string) (string, error)
}

type IDSource interface{ NewID() string }
