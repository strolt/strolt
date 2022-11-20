package tasks

import (
	"github.com/strolt/strolt/apps/strolt/internal/api/services/tasks/destinations"

	"github.com/go-chi/chi/v5"
)

func Router(r chi.Router) {
	r.Route("/tasks", func(r chi.Router) {
		r.Route("/{taskName}", func(r chi.Router) {
			r.Post("/backup", postBackup)

			destinations.Router(r)
		})
	})
}
