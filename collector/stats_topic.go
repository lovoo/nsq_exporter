package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// see https://github.com/nsqio/nsq/blob/master/nsqd/stats.go
type topic struct {
	Name         string     `json:"topic_name"`
	Channels     []*channel `json:"channels"`
	Depth        int64      `json:"depth"`
	BackendDepth int64      `json:"backend_depth"`
	MessageCount uint64     `json:"message_count"`
	Paused       bool       `json:"paused"`
}

type topicCollector []struct {
	val func(*topic) float64
	vec *prometheus.GaugeVec
}

func newTopicCollector(namespace string) topicCollector {
	labels := []string{"type", "topic", "paused"}

	return topicCollector{
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

func (c topicCollector) update(t *topic, out chan<- prometheus.Metric) {
	labels := prometheus.Labels{
		"type":   "topic",
		"topic":  t.Name,
		"paused": strconv.FormatBool(t.Paused),
	}

	for _, g := range c {
		g.vec.With(labels).Set(g.val(t))
		g.vec.Collect(out)
	}
}
