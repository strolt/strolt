package metrics

import (
	"sync"

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

type OperationsData struct {
	BackupSuccessCount int `json:"backupSuccess"`
	BackupErrorCount   int `json:"backupError"`
	PruneSuccessCount  int `json:"pruneSuccess"`
	PruneErrorCount    int `json:"pruneError"`
}

type operDataMutex struct {
	BackupSuccessCount int `json:"backupSuccess"`
	BackupErrorCount   int `json:"backupError"`
	PruneSuccessCount  int `json:"pruneSuccess"`
	PruneErrorCount    int `json:"pruneError"`

	sync.Mutex
}

var operData = &operDataMutex{}

func Operations() *Oper {
	return &Oper{}
}

func (o *Oper) BackupSuccess() {
	metrics.operations.With(prometheus.Labels{
		"type":      "success",
		"operation": "backup",
	}).Inc()

	operData.Lock()
	defer operData.Unlock()

	operData.BackupSuccessCount++

	if operData.PruneErrorCount < 0 {
		operData.PruneErrorCount = 0
	}
}

func (o *Oper) BackupError() {
	metrics.operations.With(prometheus.Labels{
		"type":      "error",
		"operation": "backup",
	}).Inc()

	operData.Lock()
	defer operData.Unlock()

	operData.BackupErrorCount++

	if operData.PruneErrorCount < 0 {
		operData.PruneErrorCount = 0
	}
}

func (o *Oper) PruneSuccess() {
	metrics.operations.With(prometheus.Labels{
		"type":      "success",
		"operation": "prune",
	}).Inc()

	operData.Lock()
	defer operData.Unlock()

	operData.PruneSuccessCount++

	if operData.PruneErrorCount < 0 {
		operData.PruneErrorCount = 0
	}
}

func (o *Oper) PruneError() {
	metrics.operations.With(prometheus.Labels{
		"type":      "error",
		"operation": "prune",
	}).Inc()

	operData.Lock()
	defer operData.Unlock()

	operData.PruneErrorCount++

	if operData.PruneErrorCount < 0 {
		operData.PruneErrorCount = 0
	}
}

func (o *Oper) Get() OperationsData {
	data := OperationsData{
		BackupSuccessCount: operData.BackupSuccessCount,
		BackupErrorCount:   operData.BackupErrorCount,
		PruneSuccessCount:  operData.PruneSuccessCount,
		PruneErrorCount:    operData.PruneErrorCount,
	}

	return data
}
