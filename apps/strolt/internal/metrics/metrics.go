package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/strolt/strolt/apps/strolt/internal/logger"
)

type Metrics struct {
	apiHandler *prometheus.CounterVec
	operations *prometheus.CounterVec
}

func Init() {
	metrics = &Metrics{}
	metrics.registerAPIHandler()
	metrics.registerOperations()
}

func (m *Metrics) registerAPIHandler() {
	m.apiHandler = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "strolt_metric_api_handler_requests_total",
			Help: "Total number of scrapes by HTTP status code.",
		},
		[]string{"code"},
	)
	m.apiHandler.WithLabelValues("200")
	m.apiHandler.WithLabelValues("400")
	m.apiHandler.WithLabelValues("401")
	m.apiHandler.WithLabelValues("404")
	m.apiHandler.WithLabelValues("500")

	prometheus.MustRegister(m.apiHandler)
}

var metrics = &Metrics{}

func APIHandlerRequestsInc(code int) {
	if code == 200 || code == 400 || code == 401 || code == 404 || code == 500 {
		metrics.apiHandler.With(prometheus.Labels{
			"code": strconv.Itoa(code),
		}).Add(1)
	} else {
		logger.New().Warnf("metrics api handler request not allowed code - %d", code)
	}
}
