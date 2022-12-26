package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func (m *Metrics) registerOperations() {
	m.operations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "strolt_operations_total",
			Help: "Total operations.",
		},
		[]string{"type", "operation"},
	)
	m.operations.WithLabelValues("success", "backup")
	m.operations.WithLabelValues("error", "backup")
	m.operations.WithLabelValues("success", "prune")
	m.operations.WithLabelValues("error", "prune")

	prometheus.MustRegister(m.operations)
}

type Oper struct{}

func Operations() *Oper {
	return &Oper{}
}

func (o *Oper) BackupSuccess() {
	metrics.operations.With(prometheus.Labels{
		"type":      "success",
		"operation": "backup",
	}).Inc()
}

func (o *Oper) BackupError() {
	metrics.operations.With(prometheus.Labels{
		"type":      "error",
		"operation": "backup",
	}).Inc()
}

func (o *Oper) PruneSuccess() {
	metrics.operations.With(prometheus.Labels{
		"type":      "success",
		"operation": "prune",
	}).Inc()
}

func (o *Oper) PruneError() {
	metrics.operations.With(prometheus.Labels{
		"type":      "error",
		"operation": "prune",
	}).Inc()
}
