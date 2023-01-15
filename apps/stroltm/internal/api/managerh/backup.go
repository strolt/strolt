package managerh

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/strolt/strolt/apps/stroltm/internal/manager"
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

	if err := backup(instanceName, serviceName, taskName); err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	apiu.RenderJSON200(w, r, apiu.ResultSuccess{Data: "success started"})
}

func backup(instanceName, serviceName, taskName string) error {
	sdk, err := getSDK(instanceName)
	if err != nil {
		return fmt.Errorf("instance not exists")
	}

	if _, err := sdk.Backup(serviceName, taskName); err != nil {
		return err
	}

	return nil
}

type backupAllResponse struct {
	SuccessStarted []backupAllStatusItem `json:"successStarted"`
	ErrorStarted   []backupAllStatusItem `json:"errorStarted"`
	*sync.Mutex
}

type backupAllStatusItem struct {
	InstanceName string `json:"instanceName"`
	ServiceName  string `json:"serviceName,omitempty"`
	TaskName     string `json:"taskName,omitempty"`
}

// backupAll godoc
// @Id					 backupAll
// @Summary      Start all backup
// @Tags         manager
// @Security BasicAuth
// @success 200 {object} backupAllResponse
// @Router       /api/v1/manager/instances/backup-all [post].
func (s *ManagerHandlers) backupAll(w http.ResponseWriter, r *http.Request) {
	items := []backupAllStatusItem{}
	itemsError := []backupAllStatusItem{}

	for _, instance := range manager.GetPreparedInstances() {
		if !instance.IsOnline || instance.Config == nil {
			itemsError = append(itemsError, backupAllStatusItem{
				InstanceName: instance.Name,
			})

			continue
		}

		for serviceName, service := range instance.Config.Services {
			for taskName := range service {
				items = append(items, backupAllStatusItem{
					InstanceName: instance.Name,
					ServiceName:  serviceName,
					TaskName:     taskName,
				})
			}
		}
	}

	response := backupAllResponse{
		SuccessStarted: []backupAllStatusItem{},
		ErrorStarted:   itemsError,
		Mutex:          &sync.Mutex{},
	}

	wg := sync.WaitGroup{}

	for _, item := range items {
		wg.Add(1)

		go func(item backupAllStatusItem) {
			err := backup(item.InstanceName, item.ServiceName, item.TaskName)

			response.Lock()
			if err == nil {
				response.SuccessStarted = append(response.SuccessStarted, item)
			} else {
				response.ErrorStarted = append(response.ErrorStarted, item)
			}
			response.Unlock()

			wg.Done()
		}(item)
	}

	wg.Wait()

	apiu.RenderJSON200(w, r, response)
}
