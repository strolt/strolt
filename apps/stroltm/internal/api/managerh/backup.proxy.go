package managerh

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/shared/apiu"
)

// backupProxy godoc
// @Id					 backupProxy
// @Summary      Start backup proxy
// @Tags         manager
// @Security BasicAuth
// @Param   proxyName           path    string     true        "Proxy name"
// @Param   instanceName        path    string     true        "Instance name"
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @success 200 {object} apiu.ResultSuccess
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/manager/instances/{proxyName}/{instanceName}/{serviceName}/tasks/{taskName}/backup [post].
func (s *ManagerHandlers) backupProxy(w http.ResponseWriter, r *http.Request) {
	proxyName := chi.URLParam(r, "proxyName")
	instanceName := chi.URLParam(r, "instanceName")
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")

	if err := backupProxy(proxyName, instanceName, serviceName, taskName); err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	apiu.RenderJSON200(w, r, apiu.ResultSuccess{Data: "success started"})
}

func backupProxy(proxyName, instanceName, serviceName, taskName string) error {
	sdk, err := getProxySDK(proxyName)
	if err != nil {
		return fmt.Errorf("instance not exists")
	}

	if _, err := sdk.Backup(instanceName, serviceName, taskName); err != nil {
		return err
	}

	return nil
}
