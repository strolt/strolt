package services

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/api/apiu"
	"github.com/strolt/strolt/apps/strolt/internal/task"
)

type getStatusResult struct {
	Data []task.ControllerOperation `json:"data"`
}

// getStatus godoc
// @Id					 getStatus
// @Summary      Get task statuses
// @Tags         services
// @Accept       json
// @Produce      json
// @success 200 {object} getStatusResult
// @Router       /api/services/status [get].
func getStatus(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, getStatusResult{Data: task.GetOperations()})
}
