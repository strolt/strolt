package api

import (
	"net/http"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/ldflags"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/apiu"
)

var startedAt = time.Now().Format(time.RFC3339)

type GetInfoResponse struct {
	Version               string `json:"version"`
	StartedAt             string `json:"startedAt"`
	ConfigUpdatedAt       string `json:"configUpdatedAt"`
	TaskStatusesUpdatedAt string `json:"taskStatusUpdatedAt"`
}

// getInfo godoc
// @Id					 getInfo
// @Summary      Get info
// @Tags         info
// @Security BasicAuth
// @success 200 {object} GetInfoResponse
// @Router       /api/v1/info [get].
func (api *API) getInfo(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, GetInfoResponse{
		Version:               ldflags.GetVersion(),
		StartedAt:             startedAt,
		ConfigUpdatedAt:       startedAt,
		TaskStatusesUpdatedAt: task.GetLastChangedManager().Format(time.RFC3339),
	})
}
