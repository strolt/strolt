package api

import (
	"net/http"

	"github.com/strolt/strolt/apps/stroltm/internal/api/apiu"
)

// getConfig godoc
// @Summary      Get config
// @Tags         services
// @Accept       json
// @Produce      json
// @success 200 {object} apiu.ResultSuccess
// @Router       /api/config [get].
func getConfig(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, apiu.ResultSuccess{Data: "config"})
}
