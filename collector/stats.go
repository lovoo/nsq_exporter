package collector

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type statsResponse struct {
	StatusCode int    `json:"status_code"`
	StatusText string `json:"status_text"`
	Data       stats  `json:"data"`
}

type stats struct {
	Version   string   `json:"version"`
	Health    string   `json:"health"`
	StartTime int64    `json:"start_time"`
	Topics    []*topic `json:"topics"`
}

type statsCollector struct {
	nsqUrl string

	topicCount   prometheus.Gauge
	channelCount *prometheus.GaugeVec
	clientCount  *prometheus.GaugeVec

	topics   topicCollector
	channels channelCollector
	clients  clientCollector
}

// NewStatsCollector create a collector which collects all NSQ metrics
// from the /stats route of the NSQ host.
func NewStatsCollector(nsqUrl string) Collector {
	const namespace = "nsq"
	return &statsCollector{
		nsqUrl: nsqUrl,

		topicCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "topics_total",
			Help:      "The total number of topics",
		}),
		channelCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "channels_total",
			Help:      "The total number of channels",
		}, []string{"topic"}),
		clientCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "channels_total",
			Help:      "The total number of channels",
		}, []string{"topic", "channel"}),

		topics:   newTopicCollector(namespace),
		channels: newChannelCollector(namespace),
		clients:  newClientCollector(namespace),
	}
}

func (c *statsCollector) Collect(out chan<- prometheus.Metric) error {
	s, err := c.fetchStats()
	if err != nil {
		return err
	}

	c.topicCount.Set(float64(len(s.Topics)))

	for _, topic := range s.Topics {
		c.channelCount.With(prometheus.Labels{
			"topic": topic.Name,
		}).Set(float64(len(topic.Channels)))

		c.topics.update(topic, out)
		for _, channel := range topic.Channels {
			c.clientCount.With(prometheus.Labels{
				"topic":   topic.Name,
				"channel": channel.Name,
			}).Set(float64(len(channel.Clients)))

			c.channels.update(topic.Name, channel, out)
			for _, client := range channel.Clients {
				c.clients.update(topic.Name, channel.Name, client, out)
			}
		}
	}
	return nil
}

func (c *statsCollector) fetchStats() (*stats, error) {
	resp, err := http.Get(c.nsqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var s statsResponse
	if err = json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	return &s.Data, nil
}
