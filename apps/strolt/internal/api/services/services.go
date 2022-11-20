package services

import (
	"github.com/strolt/strolt/apps/strolt/internal/api/services/tasks"

	"github.com/go-chi/chi/v5"
)

func Router(r chi.Router) {
	r.Route("/services", func(r chi.Router) {
		r.Get("/status", getStatus)

		r.Route("/{serviceName}", func(r chi.Router) {
			tasks.Router(r)
		})
	})
}
