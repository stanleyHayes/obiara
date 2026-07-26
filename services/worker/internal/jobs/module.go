package jobs

import (
	"context"

	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

type Module struct {
	source  application.JobSource
	handler application.Handler
}

func NewModule(source application.JobSource, handler application.Handler) Module {
	return Module{source: source, handler: handler}
}

func (module Module) Run(ctx context.Context) error {
	return module.source.Run(ctx, module.handler)
}
