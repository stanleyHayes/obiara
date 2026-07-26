// Package jobs is the composition root of the worker's job context.
package jobs

import (
	"context"

	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

type Module struct {
	scheduler *application.Scheduler
}

func NewModule(scheduler *application.Scheduler) Module {
	return Module{scheduler: scheduler}
}

func (module Module) Run(ctx context.Context) error {
	return module.scheduler.Start(ctx)
}
