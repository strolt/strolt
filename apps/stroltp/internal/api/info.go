package api

import (
	"net/http"

	"github.com/strolt/strolt/apps/stroltp/internal/ldflags"
	"github.com/strolt/strolt/shared/apiu"
	_ "github.com/strolt/strolt/shared/sdk/common"
	"github.com/strolt/strolt/shared/sdk/strolt"
)

// getInfo godoc
// @Id					 getInfo
// @Summary      Get info
// @Tags         info
// @Security BasicAuth
// @success 200 {object} common.ManagerInfo
// @Router       /api/v1/info [get].
func (api *API) getInfo(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, strolt.ManagerGetInfo(ldflags.GetVersion()))
}
