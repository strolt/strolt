package managerh

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/shared/apiu"
	"github.com/strolt/strolt/shared/sdk/common"
	"github.com/strolt/strolt/shared/sdk/strolt"
	"github.com/strolt/strolt/shared/sdk/stroltp"
)

type ManagerHandlers struct {
}

func New() *ManagerHandlers {
	return &ManagerHandlers{}
}

func (s *ManagerHandlers) Router(r chi.Router) {
	r.Get("/api/v1/manager/instances", s.getInstances)

	r.Post("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/backup", s.backupDirect)
	r.Get("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots", s.getSnapshotsDirect)
	r.Get("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune/snapshots", s.getSnapshotsForPrune)
	r.Post("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune", s.prune)
	r.Get("/api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/stats", s.getStatsDirect)
	r.Post("/api/v1/manager/instances/backup-all", s.backupAll)

	r.Post("/api/v1/manager/instances/{proxyName}/{instanceName}/{serviceName}/tasks/{taskName}/backup", s.backupProxy)
	r.Get("/api/v1/manager/instances/{proxyName}/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots", s.getSnapshotsProxy)
	r.Get("/api/v1/manager/instances/{proxyName}/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/stats", s.getStatsProxy)
}

// getInstances godoc
// @Id					 getInstances
// @Summary      Get Instances
// @Tags         manager
// @Security BasicAuth
// @success 200 {object} []common.ManagerPreparedInstance
// @Router       /api/v1/manager/instances [get].
func (s *ManagerHandlers) getInstances(w http.ResponseWriter, r *http.Request) {
	instances := strolt.ManagerGetPreparedInstances()
	instancesFromProxies := stroltp.ManagerGetPreparedInstances()

	list := make([]common.ManagerPreparedInstance, len(instances)+len(instancesFromProxies))

	i := 0

	for _, instance := range instances {
		list[i] = instance
		i++
	}

	for _, instance := range instancesFromProxies {
		list[i] = instance
		i++
	}

	apiu.RenderJSON200(w, r, list)
}
