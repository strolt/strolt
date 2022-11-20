package tasks

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/api/apiu"
	"github.com/strolt/strolt/apps/strolt/internal/logger"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"

	"github.com/go-chi/chi/v5"
)

// postBackup godoc
// @Summary      Start backup
// @Tags         services
// @Accept       json
// @Produce      json
// @Param   serviceName  path    string     true        "Service name"
// @Param   taskName     path    string     true        "Task name"
// @success 200 {object} apiu.ResultSuccess
// @success 500 {object} apiu.ResultError
// @Router       /api/services/{serviceName}/tasks/{taskName}/backup [post].
func postBackup(w http.ResponseWriter, r *http.Request) {
	log := logger.New()

	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")

	taskOperation := task.ControllerOperation{
		ServiceName: serviceName,
		TaskName:    taskName,
	}

	if taskOperation.IsWorking() {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: apiu.ErrTaskAlreadyWorking.Error()})
		return
	}

	t, err := task.New(serviceName, taskName, sctxt.TApi, sctxt.OpTypeBackup)
	if err != nil {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})
		return
	}
	defer t.Close()

	go func() {
		if err = t.BackupSourceToWorkDir(); err != nil {
			log.Error(err)
			apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})

			return
		}

		for destinationName := range t.TaskConfig.Destinations {
			log := log.WithField(destinationName, destinationName)
			log.Info("backup destination")

			_, err := t.BackupWorkDirToDestination(destinationName)
			if err != nil {
				log.Error(err)
			}
		}
	}()

	apiu.RenderJSON200(w, r, apiu.ResultSuccess{Data: "success started"})
}
