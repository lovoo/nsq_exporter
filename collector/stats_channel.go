package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// see https://github.com/nsqio/nsq/blob/master/nsqd/stats.go
type channel struct {
	Name          string    `json:"channel_name"`
	Depth         int64     `json:"depth"`
	BackendDepth  int64     `json:"backend_depth"`
	InFlightCount int       `json:"in_flight_count"`
	DeferredCount int       `json:"deferred_count"`
	MessageCount  uint64    `json:"message_count"`
	RequeueCount  uint64    `json:"requeue_count"`
	TimeoutCount  uint64    `json:"timeout_count"`
	Clients       []*client `json:"clients"`
	Paused        bool      `json:"paused"`
}

type channelCollector []struct {
	val func(*channel) float64
	vec *prometheus.GaugeVec
}

func newChannelCollector(namespace string) channelCollector {
	labels := []string{"type", "topic", "channel", "paused"}

	return channelCollector{
		{
			val: func(c *channel) float64 { return float64(c.Depth) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "depth",
				Help:      "Queue depth",
			}, labels),
		},
		{
			val: func(c *channel) float64 { return float64(c.BackendDepth) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "backend_depth",
				Help:      "Queue backend depth",
			}, labels),
		},
		{
			val: func(c *channel) float64 { return float64(c.MessageCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "message_count",
				Help:      "Queue message count",
			}, labels),
		},
		{
			val: func(c *channel) float64 { return float64(c.InFlightCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "in_flight_count",
				Help:      "In flight count",
			}, labels),
		},
		{
			val: func(c *channel) float64 { return float64(c.DeferredCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "deferred_count",
				Help:      "Deferred count",
			}, labels),
		},
		{
			val: func(c *channel) float64 { return float64(c.RequeueCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "requeue_count",
				Help:      "Requeue Count",
			}, labels),
		},
		{
			val: func(c *channel) float64 { return float64(c.TimeoutCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "timeout_count",
				Help:      "Timeout count",
			}, labels),
		},
	}
}

func (c channelCollector) update(topic string, ch *channel, out chan<- prometheus.Metric) {
	labels := prometheus.Labels{
		"type":    "channel",
		"topic":   topic,
		"channel": ch.Name,
		"paused":  strconv.FormatBool(ch.Paused),
	}

	for _, g := range c {
		g.vec.With(labels).Set(g.val(ch))
		g.vec.Collect(out)
	}
}
