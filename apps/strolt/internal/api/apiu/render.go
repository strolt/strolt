package apiu

import (
	"net/http"

	"github.com/go-chi/render"
)

func RenderJSON200(w http.ResponseWriter, r *http.Request, v interface{}) {
	render.Status(r, 200) //nolint:gomnd
	render.JSON(w, r, v)
}

func RenderJSON500(w http.ResponseWriter, r *http.Request, v interface{}) {
	render.Status(r, 500) //nolint:gomnd
	render.JSON(w, r, v)
}
