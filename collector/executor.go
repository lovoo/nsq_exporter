package collector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// NsqExecutor collects all NSQ metrics from the registered collectors.
// This type implements the prometheus.Collector interface and can be
// registered in the metrics collection.
//
// The executor takes the time needed by each registered collector and
// provides an extra metric for this. This metric is labeled with the
// result ("success" or "error") and the collector.
type NsqExecutor struct {
	collectors map[string]Collector
	summary    *prometheus.SummaryVec
}

// NewNsqExecutor creates a new executor for the NSQ metrics.
func NewNsqExecutor(namespace string) *NsqExecutor {
	return &NsqExecutor{
		collectors: make(map[string]Collector),
		summary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: "exporter",
			Name:      "scape_duration_seconds",
			Help:      "Duration of a scrape job of the NSQ exporter",
		}, []string{"collector", "result"}),
	}
}

// AddCollector adds a new collector for the metrics collection.
// Each collector needs a unique name which is used as a label
// for the executor metric.
func (e *NsqExecutor) AddCollector(name string, c Collector) {
	e.collectors[name] = c
}

// Describe implements the prometheus.Collector interface.
func (e *NsqExecutor) Describe(out chan<- *prometheus.Desc) {
	e.summary.Describe(out)
}

// Collect implements the prometheus.Collector interface.
func (e *NsqExecutor) Collect(out chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	wg.Add(len(e.collectors))
	for name, coll := range e.collectors {
		go func(name string, coll Collector) {
			e.exec(name, coll, out)
			wg.Done()
		}(name, coll)
	}
	wg.Wait()
}

func (e *NsqExecutor) exec(name string, coll Collector, out chan<- prometheus.Metric) {
	start := time.Now()
	err := coll.Collect(out)
	dur := time.Since(start)

	labels := prometheus.Labels{"collector": name}
	if err != nil {
		labels["result"] = "error"
	} else {
		labels["result"] = "success"
	}

	e.summary.With(labels).Observe(dur.Seconds())
}
