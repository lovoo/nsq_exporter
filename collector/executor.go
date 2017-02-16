package collector

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// NsqExecutor collects all NSQ metrics from the registered collectors.
// This type implements the prometheus.Collector interface and can be
// registered in the metrics collection.
//
// The executor takes the time needed for scraping nsqd stat endpoint and
// provides an extra metric for this. This metric is labeled with the
// scrape result ("success" or "error").
type NsqExecutor struct {
	nsqdURL    string
	httpClient *http.Client

	collectors []StatsCollector
	summary    *prometheus.SummaryVec
	mutex      sync.RWMutex
}

// NewNsqExecutor creates a new executor for collecting NSQ metrics.
func NewNsqExecutor(namespace, nsqdURL string, timeout time.Duration) *NsqExecutor {
	sum := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: "exporter",
		Name:      "scrape_duration_seconds",
		Help:      "Duration of a scrape job of the NSQ exporter",
	}, []string{"result"})
	prometheus.MustRegister(sum)
	return &NsqExecutor{
		nsqdURL: nsqdURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		summary: sum,
	}
}

// Use configures a specific stats collector, so the stats could be
// exposed to the Prometheus system.
func (e *NsqExecutor) Use(c StatsCollector) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.collectors = append(e.collectors, c)
}

// Describe implements the prometheus.Collector interface.
func (e *NsqExecutor) Describe(ch chan<- *prometheus.Desc) {
	for _, c := range e.collectors {
		c.describe(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (e *NsqExecutor) Collect(out chan<- prometheus.Metric) {
	start := time.Now()
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// reset state, because metrics can gone
	for _, c := range e.collectors {
		c.reset()
	}

	stats, err := getNsqdStats(e.httpClient, e.nsqdURL)
	tScrape := time.Since(start).Seconds()

	result := "success"
	if err != nil {
		result = "error"
	}

	e.summary.WithLabelValues(result).Observe(tScrape)

	if err == nil {
		for _, c := range e.collectors {
			c.set(stats)
		}
		for _, c := range e.collectors {
			c.collect(out)
		}
	}
}
