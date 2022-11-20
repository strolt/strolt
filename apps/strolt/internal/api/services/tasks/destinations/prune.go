package destinations

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/api/apiu"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"

	"github.com/go-chi/chi/v5"
)

type getPruneResult struct {
	Data task.SnapshotList `json:"data"`
}

// getPrune godoc
// @Summary      Get snapshots for prune
// @Tags         services
// @Accept       json
// @Produce      json
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} getPruneResult
// @success 500 {object} apiu.ResultError
// @Router       /api/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune [get].
func getPrune(w http.ResponseWriter, r *http.Request) {
	prune(w, r, true)
}

// postPrune godoc
// @Summary      Prune snapshots
// @Tags         services
// @Accept       json
// @Produce      json
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} getPruneResult
// @success 500 {object} apiu.ResultError
// @Router       /api/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune [post].
func postPrune(w http.ResponseWriter, r *http.Request) {
	prune(w, r, false)
}

func prune(w http.ResponseWriter, r *http.Request, idDryRun bool) {
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	taskOperation := task.ControllerOperation{
		ServiceName:     serviceName,
		TaskName:        taskName,
		DestinationName: destinationName,
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

	snapshotList, err := t.DestinationPrune(destinationName, idDryRun)
	if err != nil {
		apiu.RenderJSON500(w, r, apiu.ResultError{Error: err.Error()})
		return
	}

	apiu.RenderJSON500(w, r, getPruneResult{
		Data: snapshotList,
	})
}
