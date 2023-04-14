package managerh

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/shared/apiu"
	"github.com/strolt/strolt/shared/sdk/strolt"
	"github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_models"
	"github.com/strolt/strolt/shared/sdk/stroltp"
	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_models"
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

	r.Post("/api/v1/manager/instances/{proxyName}/{instanceName}/{serviceName}/tasks/{taskName}/backup", s.backupProxy)
	r.Get("/api/v1/manager/instances/{proxyName}/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots", s.getSnapshotsProxy)
}

// // getInstances godoc
// //	@Id			getInstances
// //	@Summary	Get Instances
// //	@Tags		manager
// //	@Security	BasicAuth
// //	@success	200	{object}	[]strolt.ManagerPreparedInstance
// //	@Router		/api/v1/manager/instances [get].
// func (s *ManagerHandlers) getInstances(w http.ResponseWriter, r *http.Request) {
// 	instances := strolt.ManagerGetPreparedInstances()

// 	apiu.RenderJSON200(w, r, instances)
// }

type ManagerPreparedInstance struct {
	ProxyName  *string                                 `json:"proxyName,omitempty"`
	Name       string                                  `json:"name"`
	Config     *stroltp_models.StroltAPIConfig         `json:"config"`
	TaskStatus *stroltp_models.StroltTaskManagerStatus `json:"taskStatus"`
	IsOnline   bool                                    `json:"isOnline"`
}

func convertConfig(src *strolt_models.APIConfig) *stroltp_models.StroltAPIConfig {
	if src == nil {
		return nil
	}

	var config stroltp_models.StroltAPIConfig

	// TODO: rewrite this

	srcJSON, err := json.Marshal(*src)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(srcJSON, &config); err != nil {
		return nil
	}

	return &config
}

func convertTaskStatus(src *strolt_models.TaskManagerStatus) *stroltp_models.StroltTaskManagerStatus {
	if src == nil {
		return nil
	}

	var config stroltp_models.StroltTaskManagerStatus

	// TODO: rewrite this

	srcJSON, err := json.Marshal(*src)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(srcJSON, &config); err != nil {
		return nil
	}

	return &config
}

// getInstances godoc
// @Id					 getInstances
// @Summary      Get Instances
// @Tags         manager
// @Security BasicAuth
// @success 200 {object} []ManagerPreparedInstance
// @Router       /api/v1/manager/instances [get].
func (s *ManagerHandlers) getInstances(w http.ResponseWriter, r *http.Request) {
	instances := strolt.ManagerGetPreparedInstances()

	instancesFromProxies := stroltp.ManagerGetPreparedInstances()

	list := make([]ManagerPreparedInstance, len(instances)+len(instancesFromProxies))

	i := 0

	for _, instance := range instances {
		list[i] = ManagerPreparedInstance{
			Name:       instance.Name,
			Config:     convertConfig(instance.Config),
			TaskStatus: convertTaskStatus(instance.TaskStatus),
			IsOnline:   instance.IsOnline,
		}
		i++
	}

	for _, instance := range instancesFromProxies {
		proxyName := instance.ProxyName
		list[i] = ManagerPreparedInstance{
			ProxyName:  &proxyName,
			Name:       instance.Name,
			Config:     instance.Config,
			TaskStatus: instance.TaskStatus,
			IsOnline:   instance.IsOnline,
		}
		i++
	}

	apiu.RenderJSON200(w, r, list)
}
