package managerh

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/shared/apiu"
	"github.com/strolt/strolt/shared/sdk/strolt"
	_ "github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_models"
	"github.com/strolt/strolt/shared/sdk/stroltp"
)

func getSDK(instanceName string) (*strolt.SDK, error) {
	sdk, ok := strolt.ManagerGetSDKByInstanceName(instanceName)
	if !ok {
		return nil, fmt.Errorf("not exists instance '%s'", instanceName)
	}

	return sdk, nil
}

func getProxySDK(proxyInstanceName string) (*stroltp.SDK, error) {
	sdk, ok := stroltp.ManagerGetSDKByInstanceName(proxyInstanceName)
	if !ok {
		return nil, fmt.Errorf("not exists proxy instance '%s'", proxyInstanceName)
	}

	return sdk, nil
}

// getSnapshotsDirect godoc
// @Id					 getSnapshotsDirect
// @Summary      Get snapshots
// @Tags         manager-direct
// @Security BasicAuth
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} strolt_models.ServicesGetSnapshotsResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots [get].
func (s *ManagerHandlers) getSnapshotsDirect(w http.ResponseWriter, r *http.Request) {
	instanceName := chi.URLParam(r, "instanceName")
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	sdk, err := getSDK(instanceName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	result, err := sdk.GetSnapshots(serviceName, taskName, destinationName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	if result == nil || result.Payload == nil {
		apiu.RenderJSON500(w, r, errors.New("response is empty"))
		return
	}

	apiu.RenderJSON200(w, r, result.Payload)
}

// getSnapshotsProxy godoc
// @Id					 getSnapshotsProxy
// @Summary      Get snapshots proxy
// @Tags         manager-proxy
// @Security BasicAuth
// @Param   proxyName           path    string     true        "Proxy name"
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} strolt_models.ServicesGetSnapshotsResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{proxyName}/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots [get].
func (s *ManagerHandlers) getSnapshotsProxy(w http.ResponseWriter, r *http.Request) {
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

	result, err := sdk.GetSnapshots(instanceName, serviceName, taskName, destinationName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	if result == nil || result.Payload == nil {
		apiu.RenderJSON500(w, r, errors.New("response is empty"))
		return
	}

	apiu.RenderJSON200(w, r, result.Payload)
}
