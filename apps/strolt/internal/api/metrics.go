package api

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/api/apiu"
	"github.com/strolt/strolt/apps/strolt/internal/metrics"
)

type getStroltMetricsResponse struct {
	Operations metrics.OperationsData `json:"operations"`
}

// getStroltMetrics godoc
// @Id					 getStroltMetrics
// @Summary      Get strolt metrics
// @Security BasicAuth
// @success 200 {object} getStroltMetricsResponse
// @Router       /api/v1/metrics [get].
func (api *API) getStroltMetrics(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, getStroltMetricsResponse{
		Operations: metrics.Operations().Get(),
	})
}
