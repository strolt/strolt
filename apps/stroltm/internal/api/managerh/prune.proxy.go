package managerh

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/shared/apiu"
	_ "github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_models"
)

// getSnapshotsForPruneProxy godoc
// @Id					 getSnapshotsForPruneProxy
// @Summary      Get snapshots for prune
// @Tags         manager-proxy
// @Security BasicAuth
// @Param   proxyName           path    string     true        "Proxy name"
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} strolt_models.ServicesGetPruneResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{proxyName}/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune/snapshots [get].
func (s *ManagerHandlers) getSnapshotsForPruneProxy(w http.ResponseWriter, r *http.Request) {
	proxyName := chi.URLParam(r, "proxyName")
	instanceName := chi.URLParam(r, "instanceName")
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	sdk, err := getProxySDK(proxyName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	result, err := sdk.GetSnapshotsForPrune(instanceName, serviceName, taskName, destinationName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	if result == nil || result.Payload == nil {
		apiu.RenderJSON500(w, r, fmt.Errorf("response is empty"))
		return
	}

	apiu.RenderJSON200(w, r, result.Payload)
}

// pruneProxy godoc
// @Id					 pruneProxy
// @Summary      Prune
// @Tags         manager-proxy
// @Security BasicAuth
// @Param   proxyName           path    string     true        "Proxy name"
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} strolt_models.ServicesGetPruneResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{proxyName}/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune [post].
func (s *ManagerHandlers) pruneProxy(w http.ResponseWriter, r *http.Request) {
	proxyName := chi.URLParam(r, "proxyName")
	instanceName := chi.URLParam(r, "instanceName")
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	sdk, err := getProxySDK(proxyName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	result, err := sdk.Prune(instanceName, serviceName, taskName, destinationName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	if result == nil || result.Payload == nil {
		apiu.RenderJSON500(w, r, fmt.Errorf("response is empty"))
		return
	}

	apiu.RenderJSON200(w, r, result.Payload)
}
