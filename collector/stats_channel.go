package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type channelStats []struct {
	val func(*channel) float64
	vec *prometheus.GaugeVec
}

// ChannelStats creates a new stats collector which is able to
// expose the channel metrics of a nsqd node to Prometheus. The
// channel metrics are reported per topic.
func ChannelStats(namespace string) StatsCollector {
	labels := []string{"topic", "channel", "paused"}
	namespace += "_channel"

	return channelStats{
		{
			val: func(c *channel) float64 { return float64(len(c.Clients)) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "client_count",
				Help:      "Number of clients",
			}, labels),
		},
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
			val: func(c *channel) float64 { return c.E2eLatency.percentileValue(0) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "e2e_latency_99p",
				Help:      "e2e latency 99th percentile",
			}, labels),
		},
		{
			val: func(c *channel) float64 { return c.E2eLatency.percentileValue(1) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "e2e_latency_95p",
				Help:      "e2e latency 95th percentile",
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

func (cs channelStats) set(s *stats) {
	for _, topic := range s.Topics {
		for _, channel := range topic.Channels {
			labels := prometheus.Labels{
				"topic":   topic.Name,
				"channel": channel.Name,
				"paused":  strconv.FormatBool(channel.Paused),
			}

			for _, c := range cs {
				c.vec.With(labels).Set(c.val(channel))
			}
		}
	}
}

func (cs channelStats) collect(out chan<- prometheus.Metric) {
	for _, c := range cs {
		c.vec.Collect(out)
	}
}

func (cs channelStats) describe(ch chan<- *prometheus.Desc) {
	for _, c := range cs {
		c.vec.Describe(ch)
	}
}

func (cs channelStats) reset() {
	for _, c := range cs {
		c.vec.Reset()
	}
}
