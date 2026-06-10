// Package services provides the HTTP API handlers for service tasks.
package services

import (
	"github.com/go-chi/chi/v5"
)

// Services groups the service task API handlers.
type Services struct{}

// New creates a Services handler group.
func New() *Services {
	return &Services{}
}

// Router mounts the service task routes on the given router.
func (s *Services) Router(r chi.Router) {
	r.Get("/api/v1/services/status", s.getStatus)

	r.Post("/api/v1/services/{serviceName}/tasks/{taskName}/backup", s.backup)
	r.Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots", s.getSnapshots)
	r.Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots/prune", s.getSnapshotsForPrune)
	r.Post("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune", s.prune)
	r.Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/stats", s.getStats)
}
