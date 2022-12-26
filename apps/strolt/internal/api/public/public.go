package public

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/shared/apiu"
)

type Public struct {
}

func New() *Public {
	return &Public{}
}

func (s *Public) Router(r chi.Router) {
	r.Get("/api/v1/ping", s.ping)

	s.prometheusMetrics(r)

	if env.IsDebug() {
		s.debug(r)
	}
}

// prometheusMetrics godoc
// @Tags         public
// @Id           getMetrics
// @Summary      Prometheus metrics
// @Success      200  {string}  string.
// @Router       /metrics [get].
func (s *Public) prometheusMetrics(r chi.Router) {
	r.Mount("/metrics", promhttp.Handler())
}

// debug godoc
// @Tags         public
// @Id           getDebug
// @Description  Available only if env `STROLT_LOG_LEVEL=DEBUG` or `STROLT_LOG_LEVEL=TRACE`
// @Summary      Go debug info
// @Success      200  {string}  string.
// @Router       /debug [get].
func (s *Public) debug(r chi.Router) {
	r.Mount("/debug", middleware.Profiler())
}

type getPingResponse struct {
	Data string `json:"data"`
}

// ping godoc
// @Tags         public
// @Id           ping
// @Summary      Ping
// @success 200 {object} getPingResponse
// @Router       /api/v1/ping [get].
func (s *Public) ping(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, getPingResponse{
		Data: "pong",
	})
}
