package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type topicsCollector []struct {
	val func(*topic) float64
	vec *prometheus.GaugeVec
}

// TopicsCollector creates a new stats collector which is able to
// expose the topic metrics of a nsqd node to Prometheus.
func TopicsCollector(namespace string) StatsCollector {
	labels := []string{"type", "topic", "paused"}

	return topicsCollector{
		{
			val: func(t *topic) float64 { return float64(len(t.Channels)) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "channel_count",
				Help:      "Number of channels",
			}, labels),
		},
		{
			val: func(t *topic) float64 { return float64(t.Depth) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "depth",
				Help:      "Queue depth",
			}, labels),
		},
		{
			val: func(t *topic) float64 { return float64(t.BackendDepth) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "backend_depth",
				Help:      "Queue backend depth",
			}, labels),
		},
		{
			val: func(t *topic) float64 { return float64(t.MessageCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "message_count",
				Help:      "Queue message count",
			}, labels),
		},
	}
}

func (coll topicsCollector) collect(s *stats, out chan<- prometheus.Metric) {
	for _, topic := range s.Topics {
		labels := prometheus.Labels{
			"type":   "topic",
			"topic":  topic.Name,
			"paused": strconv.FormatBool(topic.Paused),
		}

		for _, c := range coll {
			c.vec.With(labels).Set(c.val(topic))
			c.vec.Collect(out)
		}
	}
}
