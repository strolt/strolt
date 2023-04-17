package managerh

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/shared/apiu"
	_ "github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_models"
)

// getStats godoc
// @Id					 getStats
// @Summary      Get stats
// @Tags         manager
// @Security BasicAuth
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} strolt_models.ServicesGetStatsResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/stats [get].
func (s *ManagerHandlers) getStats(w http.ResponseWriter, r *http.Request) {
	instanceName := chi.URLParam(r, "instanceName")
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	sdk, err := getSDK(instanceName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	result, err := sdk.GetStats(serviceName, taskName, destinationName)
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
