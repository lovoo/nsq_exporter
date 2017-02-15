package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type clientStats []struct {
	val func(*client) float64
	vec *prometheus.GaugeVec
}

// ClientStats creates a new stats collector which is able to
// expose the client metrics of a nsqd node to Prometheus. The
// client metrics are reported per topic and per channel.
//
// If there are too many clients, it could cause a timeout of the
// Prometheus collection process. So be sure the number of clients
// is small enough when using this collector.
func ClientStats(namespace string) StatsCollector {
	labels := []string{"topic", "channel", "deflate", "snappy", "tls", "client_id", "hostname", "version", "remote_address"}
	namespace += "_client"

	return clientStats{
		{
			// TODO: Give state a descriptive name instead of a number.
			val: func(c *client) float64 { return float64(c.State) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "state",
				Help:      "State of client",
			}, labels),
		},
		{
			val: func(c *client) float64 { return float64(c.FinishCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "finish_count",
				Help:      "Finish count",
			}, labels),
		},
		{
			val: func(c *client) float64 { return float64(c.MessageCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "message_count",
				Help:      "Queue message count",
			}, labels),
		},
		{
			val: func(c *client) float64 { return float64(c.ReadyCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "ready_count",
				Help:      "Ready count",
			}, labels),
		},
		{
			val: func(c *client) float64 { return float64(c.InFlightCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "in_flight_count",
				Help:      "In flight count",
			}, labels),
		},
		{
			val: func(c *client) float64 { return float64(c.RequeueCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "requeue_count",
				Help:      "Requeue count",
			}, labels),
		},
		{
			val: func(c *client) float64 { return float64(c.ConnectTime) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "connect_ts",
				Help:      "Connect timestamp",
			}, labels),
		},
		{
			val: func(c *client) float64 { return float64(c.SampleRate) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "sample_rate",
				Help:      "Sample Rate",
			}, labels),
		},
	}
}

func (cs clientStats) set(s *stats) {
	for _, topic := range s.Topics {
		for _, channel := range topic.Channels {
			for _, client := range channel.Clients {
				labels := prometheus.Labels{
					"topic":          topic.Name,
					"channel":        channel.Name,
					"deflate":        strconv.FormatBool(client.Deflate),
					"snappy":         strconv.FormatBool(client.Snappy),
					"tls":            strconv.FormatBool(client.TLS),
					"client_id":      client.ID,
					"hostname":       client.Hostname,
					"version":        client.Version,
					"remote_address": client.RemoteAddress,
				}

				for _, c := range cs {
					c.vec.With(labels).Set(c.val(client))
				}
			}
		}
	}
}

func (cs clientStats) collect(out chan<- prometheus.Metric) {
	for _, c := range cs {
		c.vec.Collect(out)
	}
}

func (cs clientStats) describe(ch chan<- *prometheus.Desc) {
	for _, c := range cs {
		c.vec.Describe(ch)
	}
}

func (cs clientStats) reset() {
	for _, c := range cs {
		c.vec.Reset()
	}
}
