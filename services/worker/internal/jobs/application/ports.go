package application

import "context"

// JobSource is an inbound worker port. A durable adapter can implement leasing,
// retries, and deduplication without leaking infrastructure types inward.
type JobSource interface {
	Run(context.Context, Handler) error
}

type Handler interface {
	Handle(context.Context, Job) error
}

type Job struct {
	ID      string
	Kind    string
	Payload []byte
}
