package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/strolt/strolt/apps/stroltm/internal/ui"
)

// @title           Strolt Manager API
// @version         1.0
// @BasePath  /
// @securityDefinitions.basic  BasicAuth
func Start() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5)) //nolint:gomnd

	r.Get("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api welcome")) //nolint
	})

	r.Route("/", ui.Router)

	http.ListenAndServe(":8090", r) //nolint
}
