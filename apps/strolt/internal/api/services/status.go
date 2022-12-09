package services

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/apiu"
)

type getStatusResult struct {
	Data []task.ControllerOperation `json:"data"`
}

// getStatus godoc
// @Id					 getStatus
// @Summary      Get services status
// @Tags         services
// @Security BasicAuth
// @success 200 {object} getStatusResult
// @Router       /api/v1/services/status [get].
func (s *Services) getStatus(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, getStatusResult{Data: task.GetOperations()})
}
