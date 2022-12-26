package api

import (
	"net/http"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/ldflags"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/apiu"
)

var startedAt = time.Now().Format(time.RFC3339)

type getInfoResponse struct {
	Version               string `json:"version"`
	StartedAt             string `json:"startedAt"`
	UpdateConfigAt        string `json:"updateConfigAt"`
	TaskStatusesUpdatedAt string `json:"taskStatusUpdatedAt"`
}

// getInfo godoc
// @Id					 getInfo
// @Summary      Get info
// @Tags         info
// @Security BasicAuth
// @success 200 {object} getInfoResponse
// @Router       /api/v1/info [get].
func (api *API) getInfo(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, getInfoResponse{
		Version:               ldflags.GetVersion(),
		StartedAt:             startedAt,
		UpdateConfigAt:        startedAt,
		TaskStatusesUpdatedAt: task.GetLastChangedManager().Format(time.RFC3339),
	})
}
