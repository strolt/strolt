package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	operations *prometheus.CounterVec
}

var metrics = &Metrics{}

func Init() {
	metrics = &Metrics{}
	metrics.registerOperations()
}
