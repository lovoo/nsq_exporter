package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type topicStats []struct {
	val func(*topic) float64
	vec *prometheus.GaugeVec
}

// TopicStats creates a new stats collector which is able to
// expose the topic metrics of a nsqd node to Prometheus.
func TopicStats(namespace string) StatsCollector {
	labels := []string{"topic", "paused"}
	namespace += "_topic"

	return topicStats{
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
			val: func(t *topic) float64 { return getPercentile(t, 99) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "e2e_latency_99_percentile",
				Help:      "Queue e2e latency 99th percentile",
			}, labels),
		},
		{
			val: func(t *topic) float64 { return getPercentile(t, 95) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "e2e_latency_95_percentile",
				Help:      "Queue e2e latency 95th percentile",
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

func (ts topicStats) set(s *stats) {
	for _, topic := range s.Topics {
		labels := prometheus.Labels{
			"topic":  topic.Name,
			"paused": strconv.FormatBool(topic.Paused),
		}

		for _, c := range ts {
			c.vec.With(labels).Set(c.val(topic))
		}
	}
}
func (ts topicStats) collect(out chan<- prometheus.Metric) {
	for _, c := range ts {
		c.vec.Collect(out)
	}
}

func (ts topicStats) describe(ch chan<- *prometheus.Desc) {
	for _, c := range ts {
		c.vec.Describe(ch)
	}
}

func (ts topicStats) reset() {
	for _, c := range ts {
		c.vec.Reset()
	}
}
