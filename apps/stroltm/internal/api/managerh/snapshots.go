package managerh

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/apps/stroltm/internal/api/apiu"
	"github.com/strolt/strolt/apps/stroltm/internal/manager"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt"
)

func getSDK(instanceName string) (*strolt.Sdk, error) {
	instance, ok := manager.GetStroltInstanceByInstanceName(instanceName)
	if !ok {
		return nil, fmt.Errorf("not exists instance '%s'", instanceName)
	}

	sdk := strolt.New(instance.URL, instance.Username, instance.Password)

	return sdk, nil
}

// getSnapshots godoc
// @Id					 getSnapshots
// @Summary      Get snapshots
// @Tags         manager
// @Security BasicAuth
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} models.ServicesGetSnapshotsResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots [get].
func (s *ManagerHandlers) getSnapshots(w http.ResponseWriter, r *http.Request) {
	instanceName := chi.URLParam(r, "instanceName")
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	sdk, err := getSDK(instanceName)
	if err != nil {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})
		return
	}

	result, err := sdk.GetSnapshots(serviceName, taskName, destinationName)
	if err != nil {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})
		return
	}

	if result == nil || result.Payload == nil {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: "response is empty"})
		return
	}

	apiu.RenderJSON200(w, r, result.Payload)
}
