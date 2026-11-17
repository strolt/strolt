// Package services provides the HTTP API handlers for service tasks.
package services

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Services groups the service task API handlers.
type Services struct {
	processContext context.Context //nolint:containedctx // Operation context lives for the service request or background task.
	beginOperation func() (func(), bool)
}

// New creates a Services handler group.
func New(processContext context.Context, beginOperation func() (func(), bool)) *Services {
	return &Services{processContext: processContext, beginOperation: beginOperation}
}

// Router mounts the service task routes on the given router.
func (s *Services) Router(r chi.Router) {
	r.Get("/api/v1/services/status", s.getStatus)

	r.With(s.trackOperation).Post("/api/v1/services/{serviceName}/tasks/{taskName}/backup", s.backup)
	r.With(s.trackOperation).Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots", s.getSnapshots)
	r.With(s.trackOperation).Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots/prune", s.getSnapshotsForPrune)
	r.With(s.trackOperation).Post("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune", s.prune)
	r.With(s.trackOperation).Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/stats", s.getStats)
}

func (s *Services) trackOperation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release, ok := s.beginOperation()
		if !ok {
			http.Error(w, "server is shutting down", http.StatusServiceUnavailable)
			return
		}
		defer release()
		next.ServeHTTP(w, r)
	})
}

func (s *Services) operationContext(requestContext context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(requestContext)
	stop := context.AfterFunc(s.processContext, func() {
		cancel(context.Cause(s.processContext))
	})
	if s.processContext.Err() != nil {
		cancel(context.Cause(s.processContext))
	}
	return ctx, func() {
		stop()
		cancel(context.Canceled)
	}
}
