package managerh

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/shared/apiu"
)

// backup godoc
// @Id					 backup
// @Summary      Start backup
// @Tags         manager
// @Security BasicAuth
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @success 200 {object} apiu.ResultSuccess
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{instanceName}/{serviceName}/tasks/{taskName}/backup [post].
func (s *ManagerHandlers) backup(w http.ResponseWriter, r *http.Request) {
	instanceName := chi.URLParam(r, "instanceName")
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")

	sdk, err := getSDK(instanceName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	result, err := sdk.Backup(serviceName, taskName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	if result == nil || result.Payload == nil {
		apiu.RenderJSON500(w, r, fmt.Errorf("response is empty"))
		return
	}

	apiu.RenderJSON200(w, r, apiu.ResultSuccess{Data: result.Payload.Data})
}
