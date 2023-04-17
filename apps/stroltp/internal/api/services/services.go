package services

import (
	"github.com/go-chi/chi/v5"
)

type Services struct {
}

func New() *Services {
	return &Services{}
}

func (s *Services) Router(r chi.Router) {
	r.Get("/api/v1/services", s.getList)
}

// r.Post("/api/v1/services/{serviceName}/tasks/{taskName}/backup", s.backup)
// r.Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots", s.getSnapshots)
// r.Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots/prune", s.getSnapshotsForPrune)
// r.Post("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune", s.prune)
// r.Get("/api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/stats", s.getStats)
