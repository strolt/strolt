package api

import (
	"net/http"

	"github.com/strolt/strolt/apps/stroltm/internal/manager2"
	"github.com/strolt/strolt/shared/apiu"
)

// getInfo godoc
// @Id					 getInfo
// @Summary      Get Info
// @Tags         global
// @Security BasicAuth
// @success 200 {object} manager2.Info
// @Router       /api/v1/info [get].
func (api *API) getInfo(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, manager2.GetInfo())
}
