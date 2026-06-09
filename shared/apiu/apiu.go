// Package apiu provides HTTP API utilities for rendering JSON responses.
package apiu

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

// ErrTaskAlreadyWorking is returned when a task is already in progress.
var ErrTaskAlreadyWorking = errors.New("task already working")

// ResultError is a JSON response body containing an error message.
type ResultError struct {
	Error string `json:"error"`
}

// ResultSuccess is a JSON response body containing a data string.
type ResultSuccess struct {
	Data string `json:"data"`
}

// RenderJSON200 writes v as a JSON response with status 200.
func RenderJSON200(w http.ResponseWriter, r *http.Request, v any) {
	render.Status(r, 200) //nolint:mnd
	render.JSON(w, r, v)
}

// RenderJSON400 writes err as a JSON error response with status 400.
func RenderJSON400(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, 400) //nolint:mnd
	render.JSON(w, r, ResultError{Error: err.Error()})
}

// RenderJSON401 writes an empty JSON error response with status 401.
func RenderJSON401(w http.ResponseWriter, r *http.Request) {
	render.Status(r, 401) //nolint:mnd
	render.JSON(w, r, ResultError{Error: ""})
}

// RenderJSON500 writes err as a JSON error response with status 500.
func RenderJSON500(w http.ResponseWriter, r *http.Request, err error) {
	render.Status(r, 500) //nolint:mnd
	render.JSON(w, r, ResultError{Error: err.Error()})
}
