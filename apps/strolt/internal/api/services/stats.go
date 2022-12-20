package services

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/apiu"

	"github.com/go-chi/chi/v5"
)

// getStats godoc
// @Id					 getStats
// @Summary      Get stats
// @Tags         services
// @Security BasicAuth
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} getStatsResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/stats [get].
func (s *Services) getStats(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	taskOperation := task.ControllerOperation{
		ServiceName:     serviceName,
		TaskName:        taskName,
		DestinationName: destinationName,
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
	defer t.Close()

	stats, err := t.GetStats(destinationName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	apiu.RenderJSON200(w, r, getStatsResult{
		Data: stats,
	})
}

type getStatsResult struct {
	Data interfaces.FormattedStats `json:"data"`
}
