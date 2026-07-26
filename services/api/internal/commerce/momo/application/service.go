package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/domain"
	"strings"
)

var (
	ErrInvalid     = errors.New("invalid momo request")
	ErrNotFound    = errors.New("momo intent not found")
	ErrConflict    = errors.New("momo intent conflict")
	ErrApplied     = errors.New("momo command already applied")
	ErrUnavailable = errors.New("momo unavailable")
)

type Service struct {
	repo     Repository
	provider Provider
	ids      IDSource
	clock    Clock
	secret   []byte
}

func New(r Repository, p Provider, ids IDSource, c Clock, secret []byte) Service {
	return Service{r, p, ids, c, append([]byte(nil), secret...)}
}
func PhoneRef(secret []byte, phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if len(secret) < 32 || len(phone) < 8 || len(phone) > 20 {
		return "", ErrInvalid
	}
	for _, r := range phone {
		if (r < '0' || r > '9') && r != '+' {
			return "", ErrInvalid
		}
	}
	m := hmac.New(sha256.New, secret)
	m.Write([]byte("momo-phone:v1:" + phone))
	return hex.EncodeToString(m.Sum(nil)), nil
}
func (s Service) Create(ctx context.Context, memberKey, phoneRef string, amount uint64, command string) (domain.Intent, error) {
	i, e := domain.Create(s.ids.NewID(), memberKey, phoneRef, amount, command, s.clock.Now())
	if e != nil {
		return domain.Intent{}, ErrInvalid
	}
	if e = s.repo.Create(ctx, i); e != nil {
		return domain.Intent{}, e
	}
	return i, nil
}
func (s Service) Confirm(ctx context.Context, id, command string) (domain.Intent, error) {
	i, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Intent{}, e
	}
	n, e := i.Confirm(command, s.clock.Now())
	if e != nil {
		return domain.Intent{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, n, i.Revision(), command); e != nil {
		return domain.Intent{}, e
	}
	ref := s.ids.NewID()
	r, e := s.provider.RequestCollection(ctx, ProviderRequest{ref, n.State().PhoneRef, n.State().AmountPesewas, "GHS"})
	if e != nil || r != ref {
		return domain.Intent{}, ErrUnavailable
	}
	requested, e := n.MarkRequested(command+":provider", ref, s.clock.Now())
	if e != nil {
		return domain.Intent{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, requested, n.Revision(), command+":provider"); e != nil {
		return domain.Intent{}, e
	}
	return requested, nil
}

type Callback struct {
	CallbackID, IntentID, ProviderRef string
	Success                           bool
	OccurredUnix                      int64
	Signature                         string
}

func (s Service) Callback(ctx context.Context, c Callback) (domain.Intent, error) {
	if len(s.secret) < 32 || !s.valid(c) {
		return domain.Intent{}, ErrInvalid
	}
	i, e := s.repo.Find(ctx, c.IntentID)
	if e != nil {
		return domain.Intent{}, e
	}
	n, e := i.ApplyProvider(c.CallbackID, c.ProviderRef, c.Success, s.clock.Now())
	if e != nil {
		return domain.Intent{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, n, i.Revision(), c.CallbackID); e != nil {
		return domain.Intent{}, e
	}
	return n, nil
}
func (s Service) valid(c Callback) bool {
	payload := c.CallbackID + "|" + c.IntentID + "|" + c.ProviderRef + "|" + boolText(c.Success) + "|" + fmtInt(c.OccurredUnix)
	m := hmac.New(sha256.New, s.secret)
	m.Write([]byte(payload))
	got, e := hex.DecodeString(c.Signature)
	return e == nil && hmac.Equal(got, m.Sum(nil))
}
func SignCallback(secret []byte, c Callback) string {
	payload := c.CallbackID + "|" + c.IntentID + "|" + c.ProviderRef + "|" + boolText(c.Success) + "|" + fmtInt(c.OccurredUnix)
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(payload))
	return hex.EncodeToString(m.Sum(nil))
}
func boolText(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
func fmtInt(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
