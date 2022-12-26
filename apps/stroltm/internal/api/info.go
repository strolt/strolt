package api

import (
	"net/http"

	"github.com/strolt/strolt/apps/stroltm/internal/manager"
	"github.com/strolt/strolt/shared/apiu"
)

// getInfo godoc
// @Id					 getInfo
// @Summary      Get Info
// @Tags         global
// @Security BasicAuth
// @success 200 {object} manager.Info
// @Router       /api/v1/info [get].
func (api *API) getInfo(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, manager.GetInfo())
}
