package services

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/apiu"

	"github.com/go-chi/chi/v5"
)

type getPruneResult struct {
	Data task.SnapshotList `json:"data"`
}

// getSnapshotsForPrune godoc
// @Id					 getSnapshotsForPrune
// @Summary      Get snapshots for prune
// @Tags         services
// @Security BasicAuth
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} getPruneResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots/prune [get].
func (s *Services) getSnapshotsForPrune(w http.ResponseWriter, r *http.Request) {
	prune(w, r, true)
}

// prune godoc
// @Id					 prune
// @Summary      Prune snapshots
// @Tags         services
// @Security BasicAuth
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} getPruneResult
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/prune [post].
func (s *Services) prune(w http.ResponseWriter, r *http.Request) {
	prune(w, r, false)
}

func prune(w http.ResponseWriter, r *http.Request, idDryRun bool) {
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	t, err := task.New(serviceName, taskName, sctxt.TApi, sctxt.OpTypePrune)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}
	defer t.Close()

	if t.IsRunning() {
		apiu.RenderJSON400(w, r, apiu.ErrTaskAlreadyWorking)
		return
	}

	snapshotList, err := t.Prune(destinationName, idDryRun)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	apiu.RenderJSON200(w, r, getPruneResult{
		Data: snapshotList,
	})
}
