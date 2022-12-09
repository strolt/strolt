package apiu

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

var (
	ErrTaskAlreadyWorking = fmt.Errorf("task already working")
)

type ResultError struct {
	Error string `json:"error"`
}

type ResultSuccess struct {
	Data string `json:"data"`
}

func RenderJSON200(w http.ResponseWriter, r *http.Request, v interface{}) {
	render.Status(r, 200) //nolint:gomnd
	render.JSON(w, r, v)
}

func RenderJSON400(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, 400) //nolint:gomnd
	render.JSON(w, r, ResultError{Error: err.Error()})
}

func RenderJSON500(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, 500) //nolint:gomnd
	render.JSON(w, r, ResultError{Error: err.Error()})
}
