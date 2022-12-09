package managerh

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/apps/stroltm/internal/manager"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/models"
	"github.com/strolt/strolt/shared/apiu"
)

type ManagerHandlers struct {
}

func New() *ManagerHandlers {
	return &ManagerHandlers{}
}

func (s *ManagerHandlers) Router(r chi.Router) {
	r.Get("/api/v1/manager/instances", s.getInstances)
	r.Post("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/backup", s.backup)
	r.Get("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots", s.getSnapshots)
	r.Get("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune/snapshots", s.getSnapshotsForPrune)
	r.Post("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune", s.prune)
}

type getInstancesResult struct {
	Items []getInstancesResultItem `json:"data"`
}

type getInstancesResultItem struct {
	InstanceName        string            `json:"instanceName"`
	Config              *models.APIConfig `json:"config"`
	IsOnline            bool              `json:"isOnline"`
	LatestSuccessPingAt string            `json:"latestSuccessPingAt"`
}

// getInstances godoc
// @Id					 getInstances
// @Summary      Get Instances
// @Tags         manager
// @Security BasicAuth
// @success 200 {object} getInstancesResult
// @Router       /api/v1/manager/instances [get].
func (s *ManagerHandlers) getInstances(w http.ResponseWriter, r *http.Request) {
	stroltInstances := manager.GetStroltInstances()

	items := make([]getInstancesResultItem, len(stroltInstances))

	for i, instance := range stroltInstances {
		items[i] = getInstancesResultItem{
			InstanceName:        instance.InstanceName,
			Config:              instance.Config,
			IsOnline:            instance.IsOnline,
			LatestSuccessPingAt: instance.Watch.LatestSuccessPingAt.Format(time.RFC3339),
		}
	}

	apiu.RenderJSON200(w, r, getInstancesResult{Items: items})
}
