// Package metrics registers and updates the strolt Prometheus metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds the registered Prometheus collectors.
type Metrics struct {
	operations *prometheus.CounterVec
}

var metrics = &Metrics{}

// Init registers the strolt Prometheus metrics.
func Init() {
	metrics = &Metrics{}
	metrics.registerOperations()
}
