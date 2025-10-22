package apiu

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

var (
	ErrTaskAlreadyWorking = errors.New("task already working")
)

type ResultError struct {
	Error string `json:"error"`
}

type ResultSuccess struct {
	Data string `json:"data"`
}

func RenderJSON200(w http.ResponseWriter, r *http.Request, v interface{}) {
	render.Status(r, 200) //nolint:mnd
	render.JSON(w, r, v)
}

func RenderJSON400(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, 400) //nolint:mnd
	render.JSON(w, r, ResultError{Error: err.Error()})
}

func RenderJSON401(w http.ResponseWriter, r *http.Request) {
	render.Status(r, 401) //nolint:mnd
	render.JSON(w, r, ResultError{Error: ""})
}

func RenderJSON500(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, 500) //nolint:mnd
	render.JSON(w, r, ResultError{Error: err.Error()})
}
