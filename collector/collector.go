package collector

import "github.com/prometheus/client_golang/prometheus"

// Collector defines the interface for collecting all metrics for Prometheus.
type Collector interface {
	Collect(out chan<- prometheus.Metric) error
}
