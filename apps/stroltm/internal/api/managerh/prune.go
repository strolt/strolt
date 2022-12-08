package managerh

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/apps/stroltm/internal/api/apiu"
)

// getSnapshotsForPrune godoc
// @Id					 getSnapshotsForPrune
// @Summary      Get snapshots for prune
// @Tags         manager
// @Security BasicAuth
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} models.ServicesGetPruneResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune/snapshots [get].
func (s *ManagerHandlers) getSnapshotsForPrune(w http.ResponseWriter, r *http.Request) {
	instanceName := chi.URLParam(r, "instanceName")
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	sdk, err := getSDK(instanceName)
	if err != nil {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})
		return
	}

	result, err := sdk.GetSnapshotsForPrune(serviceName, taskName, destinationName)
	if err != nil {
		// fmt.Println("err", err)
		// fmt.Println("IsServerError", result.IsServerError())
		// fmt.Println("IsClientError", result.IsClientError())
		// fmt.Println("result.Payload", result.Payload)
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})
		return
	}

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

// prune godoc
// @Id					 prune
// @Summary      Prune
// @Tags         manager
// @Security BasicAuth
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} models.ServicesGetPruneResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune [post].
func (s *ManagerHandlers) prune(w http.ResponseWriter, r *http.Request) {
	instanceName := chi.URLParam(r, "instanceName")
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	sdk, err := getSDK(instanceName)
	if err != nil {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})
		return
	}

	result, err := sdk.Prune(serviceName, taskName, destinationName)
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
