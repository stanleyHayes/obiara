// Package health implements the service health semantics required by
// agent_plan.md §12: /live reports process liveness only, while /ready
// reports dependency-degraded readiness and fails when a required
// dependency (MongoDB) is unreachable.
package health

import (
	"context"
	"net/http"
	"time"
)

const readyTimeout = 2 * time.Second

// Live reports that the process is up. It never checks dependencies so
// orchestrators do not restart a process that is merely degraded.
func Live() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

// Ready reports whether the service can serve traffic. It returns 503 when
// the ping function fails, signalling dependency-degraded operation.
func Ready(ping func(context.Context) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readyTimeout)
		defer cancel()
		if err := ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("dependency unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}
