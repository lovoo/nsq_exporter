package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// StatsCollector defines an interface for collecting specific stats
// from a nsqd node.
type StatsCollector interface {
	collect(s *stats, out chan<- prometheus.Metric)
}

// NsqdStats represents a Collector which collects all the configured stats
// from a nsqd node. Besides the configured stats it will also expose a
// metric for the total number of existing topics.
type NsqdStats struct {
	nsqdURL    string
	collectors []StatsCollector
	topicCount prometheus.Gauge
}

// NewNsqdStats creates a new stats collector which uses the given namespace
// and reads the stats from the given URL of the nsqd.
func NewNsqdStats(namespace, nsqdURL string) *NsqdStats {
	return &NsqdStats{
		nsqdURL: nsqdURL,
		topicCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "topics_total",
			Help:      "The total number of topics",
		}),
	}
}

// Use configures a specific stats collector, so the stats could be
// exposed to the Prometheus system.
func (s *NsqdStats) Use(c StatsCollector) {
	s.collectors = append(s.collectors, c)
}

// Collect collects all the registered stats metrics from the nsqd node.
func (s *NsqdStats) Collect(out chan<- prometheus.Metric) error {
	stats, err := getNsqdStats(s.nsqdURL)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(s.collectors))
	for _, coll := range s.collectors {
		go func(coll StatsCollector) {
			coll.collect(stats, out)
			wg.Done()
		}(coll)
	}

	s.topicCount.Set(float64(len(stats.Topics)))
	wg.Wait()
	return nil
}
