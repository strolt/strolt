package api

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/apps/stroltm/internal/api/instances"
	"github.com/strolt/strolt/apps/stroltm/internal/ui"
)

// @title           Strolt Manager API
// @version         1.0
// @BasePath  /
// @securityDefinitions.basic  BasicAuth.
func Start() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5)) //nolint:gomnd

	r.Route("/api", func(r chi.Router) {
		instances.Router(r)
	})

	r.Route("/", ui.Router)

	http.ListenAndServe(":8090", r) //nolint
}
