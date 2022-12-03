package public

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/strolt/strolt/apps/strolt/internal/api/apiu"
)

func Router(r chi.Router) {
	r.Route("/", func(r chi.Router) {
		prometheusMetrics(r)
		debug(r)

		r.Get("/ping", getPing)
	})
}

// prometheusMetrics godoc
// @Tags         public
// @Id           getMetrics
// @Summary      Prometheus metrics
// @Success      200  {string}  string.
// @Router       /metrics [get].
func prometheusMetrics(r chi.Router) {
	r.Mount("/metrics", promhttp.Handler())
}

// debug godoc
// @Tags         public
// @Id           getDebug
// @Summary      Go debug info
// @Success      200  {string}  string.
// @Router       /debug [get].
func debug(r chi.Router) {
	r.Mount("/debug", middleware.Profiler())
}

type getPingResponse struct {
	Data string `json:"data"`
}

// getPing godoc
// @Tags         public
// @Id           getPing
// @Summary      Get ping
// @success 200 {object} getPingResponse
// @Router       /ping [get].
func getPing(w http.ResponseWriter, r *http.Request) {
	apiu.RenderJSON200(w, r, getPingResponse{Data: "pong"})
}
