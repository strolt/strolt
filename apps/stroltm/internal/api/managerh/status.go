package managerh

import (
	"net/http"
	"sync"
	"time"

	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/models"
	"github.com/strolt/strolt/shared/apiu"
)

type instacesStatus struct {
	*sync.Mutex
	LatestUpdatedAt time.Time
	Instances       map[string]models.TaskManagerStatus
}

var instacesStatusVar = instacesStatus{
	Mutex:           &sync.Mutex{},
	LatestUpdatedAt: time.Now(),
}

type responseInstacesStatus struct {
	LatestUpdatedAt string                              `json:"latestUpdatedAt"`
	Instances       map[string]models.TaskManagerStatus `json:"instances"`
}

// getStatus godoc
// @Id					 getStatus
// @Summary      Get status
// @Tags         manager
// @Security BasicAuth
// @success 200 {object} responseInstacesStatus
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/status [get].
func (s *ManagerHandlers) getStatus(w http.ResponseWriter, r *http.Request) {
	instacesStatusVar.Lock()
	status := responseInstacesStatus{
		LatestUpdatedAt: instacesStatusVar.LatestUpdatedAt.Format(time.RFC3339),
		Instances:       instacesStatusVar.Instances,
	}
	defer instacesStatusVar.Unlock()

	apiu.RenderJSON200(w, r, status)
}
