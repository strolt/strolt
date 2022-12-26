package services

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/apiu"

	"github.com/go-chi/chi/v5"
)

// getSnapshots godoc
// @Id					 getSnapshots
// @Summary      Get snapshots
// @Tags         services
// @Security BasicAuth
// @Param   serviceName         path    string     true        "Service name"
// @Param   taskName            path    string     true        "Task name"
// @Param   destinationName     path    string     true        "Destination name"
// @success 200 {object} getSnapshotsResult
// @success 400 {object} apiu.ResultError
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/services/{serviceName}/tasks/{taskName}/destinations/{destinationName}/snapshots [get].
func (s *Services) getSnapshots(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")
	destinationName := chi.URLParam(r, "destinationName")

	t, err := task.New(serviceName, taskName, sctxt.TApi, sctxt.OpTypeSnapshots)
	if err != nil {
		apiu.RenderJSON400(w, r, err)
		return
	}
	defer t.Close()

	if t.IsRunning() {
		apiu.RenderJSON400(w, r, apiu.ErrTaskAlreadyWorking)
		return
	}

	snapshotList, err := t.GetSnapshotList(destinationName)
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}

	apiu.RenderJSON200(w, r, getSnapshotsResult{
		Items: snapshotList,
	})
}

type getSnapshotsResult struct {
	Items task.SnapshotList `json:"items"`
}
