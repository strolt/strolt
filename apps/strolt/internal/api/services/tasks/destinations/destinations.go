package destinations

import "github.com/go-chi/chi/v5"

func Router(r chi.Router) {
	r.Route("/destinations", func(r chi.Router) {
		r.Route("/{destinationName}", func(r chi.Router) {
			r.Get("/snapshots", getSnapshots)
			r.Get("/prune", getPrune)
			r.Post("/prune", postPrune)
		})
	})
}
