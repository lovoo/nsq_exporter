package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// see https://github.com/nsqio/nsq/blob/master/nsqd/stats.go
type client struct {
	ID                            string `json:"client_id"`
	Hostname                      string `json:"hostname"`
	Version                       string `json:"version"`
	RemoteAddress                 string `json:"remote_address"`
	State                         int32  `json:"state"`
	ReadyCount                    int64  `json:"ready_count"`
	InFlightCount                 int64  `json:"in_flight_count"`
	MessageCount                  uint64 `json:"message_count"`
	FinishCount                   uint64 `json:"finish_count"`
	RequeueCount                  uint64 `json:"requeue_count"`
	ConnectTime                   int64  `json:"connect_ts"`
	SampleRate                    int32  `json:"sample_rate"`
	Deflate                       bool   `json:"deflate"`
	Snappy                        bool   `json:"snappy"`
	UserAgent                     string `json:"user_agent"`
	Authed                        bool   `json:"authed,omitempty"`
	AuthIdentity                  string `json:"auth_identity,omitempty"`
	AuthIdentityURL               string `json:"auth_identity_url,omitempty"`
	TLS                           bool   `json:"tls"`
	CipherSuite                   string `json:"tls_cipher_suite"`
	TLSVersion                    string `json:"tls_version"`
	TLSNegotiatedProtocol         string `json:"tls_negotiated_protocol"`
	TLSNegotiatedProtocolIsMutual bool   `json:"tls_negotiated_protocol_is_mutual"`
}

type clientCollector []struct {
	val func(*client) float64
	vec *prometheus.GaugeVec
}

func newClientCollector(namespace string) clientCollector {
	labels := []string{"type", "topic", "channel", "deflate", "snappy", "tls", "client_id", "hostname", "version", "remote_address"}

	return clientCollector{
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

func (c clientCollector) update(topic, channel string, cl *client, out chan<- prometheus.Metric) {
	labels := prometheus.Labels{
		"type":           "client",
		"topic":          topic,
		"channel":        channel,
		"deflate":        strconv.FormatBool(cl.Deflate),
		"snappy":         strconv.FormatBool(cl.Snappy),
		"tls":            strconv.FormatBool(cl.TLS),
		"client_id":      cl.ID,
		"hostname":       cl.Hostname,
		"version":        cl.Version,
		"remote_address": cl.RemoteAddress,
	}

	for _, g := range c {
		g.vec.With(labels).Set(g.val(cl))
		g.vec.Collect(out)
	}
}
