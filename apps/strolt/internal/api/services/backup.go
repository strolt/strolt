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
//
//	@Id			backup
//	@Summary	Start backup
//	@Tags		services
//	@Security	BasicAuth
//	@Param		serviceName	path		string	true	"Service name"
//	@Param		taskName	path		string	true	"Task name"
//	@success	200			{object}	apiu.ResultSuccess
//	@success	500			{object}	apiu.ResultError
//	@Router		/api/v1/services/{serviceName}/tasks/{taskName}/backup [post].
func (s *Services) backup(w http.ResponseWriter, r *http.Request) {
	serviceName := chi.URLParam(r, "serviceName")
	taskName := chi.URLParam(r, "taskName")

	releaseOperation, ok := s.beginOperation()
	if !ok {
		http.Error(w, "server is shutting down", http.StatusServiceUnavailable)
		return
	}
	transferred := false
	defer func() {
		if !transferred {
			releaseOperation()
		}
	}()
	t, err := task.New(s.processContext, serviceName, taskName, sctxt.TApi, sctxt.OpTypeBackup) //nolint:contextcheck // Task owns the daemon-scoped operation context.
	if err != nil {
		apiu.RenderJSON500(w, r, err)
		return
	}
	if t.IsRunning() {
		_ = t.Close()
		apiu.RenderJSON400(w, r, apiu.ErrTaskAlreadyWorking)
		return
	}

	transferred = true
	//nolint:contextcheck // The task retains the daemon-scoped context after this handler returns.
	go func() {
		defer releaseOperation()
		defer func() {
			if err := t.Close(); err != nil {
				logger.New().Error(err)
			}
		}()
		if err := t.Backup(); err != nil {
			logger.New().Error(err)
		}
	}()

	apiu.RenderJSON200(w, r, apiu.ResultSuccess{Data: "success started"})
}
