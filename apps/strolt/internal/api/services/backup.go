package services

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/apiu"
	"github.com/strolt/strolt/shared/logger"

	"github.com/go-chi/chi/v5"
)

// backup godoc
// @Id					 backup
// @Summary      Start backup
// @Tags         services
// @Security BasicAuth
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @success 200 {object} apiu.ResultSuccess
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/services/{serviceName}/tasks/{taskName}/backup [post].
func (s *Services) backup(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")

	taskOperation := task.ControllerOperation{
		ServiceName: serviceName,
		TaskName:    taskName,
	}

	if taskOperation.IsWorking() {
		apiu.RenderJSON500(w, r, apiu.ErrTaskAlreadyWorking)
		return
	}

	t, err := task.New(serviceName, taskName, sctxt.TApi, sctxt.OpTypeBackup)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	go func() {
		defer t.Close()

		if err := t.Backup(); err != nil {
			logger.New().Error(err)
		}
	}()

	// go func() {
	// 	defer t.Close()

	// 	if err = t.BackupSourceToWorkDir(); err != nil {
	// 		log.Error(err)
	// 		apiu.RenderJSON500(w, r, err)

	// 		return
	// 	}

	// 	for destinationName := range t.TaskConfig.Destinations {
	// 		log := log.WithField(destinationName, destinationName)
	// 		log.Info("backup destination")

	// 		// log.Debug("Backup source to destination sleep started...")
	// 		// time.Sleep(time.Minute * 1)
	// 		// log.Debug("Backup source to destination sleep finished...")

	// 		_, err := t.BackupWorkDirToDestination(destinationName)
	// 		if err != nil {
	// 			log.Error(err)
	// 		}
	// 	}
	// }()

	apiu.RenderJSON200(w, r, apiu.ResultSuccess{Data: "success started"})
}
