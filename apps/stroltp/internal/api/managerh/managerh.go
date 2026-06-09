// Package managerh provides HTTP handlers for managing strolt instances.
package managerh

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/shared/apiu"
	_ "github.com/strolt/strolt/shared/sdk/common" // register swagger models
	"github.com/strolt/strolt/shared/sdk/strolt"
)

func getSDK(instanceName string) (*strolt.SDK, error) {
	sdk, ok := strolt.ManagerGetSDKByInstanceName(instanceName)
	if !ok {
		return nil, fmt.Errorf("not exists instance '%s'", instanceName)
	}

	return sdk, nil
}

// ManagerHandlers groups HTTP handlers for strolt instance management.
type ManagerHandlers struct{}

// New creates a ManagerHandlers instance.
func New() *ManagerHandlers {
	return &ManagerHandlers{}
}

// Router registers manager routes on the given router.
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
//
//	@Id			getInstances
//	@Summary	Get Instances
//	@Tags		manager
//	@Security	BasicAuth
//	@success	200	{object}	[]common.ManagerPreparedInstance
//	@Router		/api/v1/manager/instances [get].
func (s *ManagerHandlers) getInstances(w http.ResponseWriter, r *http.Request) {
	instances := strolt.ManagerGetPreparedInstances()

	apiu.RenderJSON200(w, r, instances)
}
