package managerh

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/apps/stroltm/internal/manager"
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
	r.Get("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/stats", s.getStats)
	r.Post("/api/v1/manager/instances/backup-all", s.backupAll)
}

// getInstances godoc
// @Id					 getInstances
// @Summary      Get Instances
// @Tags         manager
// @Security BasicAuth
// @success 200 {object} []manager.PreparedInstance
// @Router       /api/v1/manager/instances [get].
func (s *ManagerHandlers) getInstances(w http.ResponseWriter, r *http.Request) {
	instances := manager.GetPreparedInstances()

	apiu.RenderJSON200(w, r, instances)
}
