// Package services provides HTTP handlers for the services API.
package services

import (
	"github.com/go-chi/chi/v5"
)

// Services groups HTTP handlers for the services API.
type Services struct{}

// New creates a Services handlers instance.
func New() *Services {
	return &Services{}
}

// Router registers services routes on the given router.
func (s *Services) Router(r chi.Router) {
	r.Get("/api/v1/services", s.getList)
}

// r.Post("/api/v1/services/{serviceName}/tasks/{taskName}/backup", s.backup)
// r.Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots", s.getSnapshots)
// r.Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots/prune", s.getSnapshotsForPrune)
// r.Post("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune", s.prune)
// r.Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/stats", s.getStats)
