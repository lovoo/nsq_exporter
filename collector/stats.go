package collector

import (
	"encoding/json"
	"net/http"
)

type stats struct {
	Version   string   `json:"version"`
	Health    string   `json:"health"`
	StartTime int64    `json:"start_time"`
	Topics    []*topic `json:"topics"`
}

// see https://github.com/nsqio/nsq/blob/master/nsqd/stats.go
type topic struct {
	Name         string     `json:"topic_name"`
	Paused       bool       `json:"paused"`
	Depth        int64      `json:"depth"`
	BackendDepth int64      `json:"backend_depth"`
	MessageCount uint64     `json:"message_count"`
	E2eLatency   e2elatency `json:"e2e_processing_latency"`
	Channels     []*channel `json:"channels"`
}

type channel struct {
	Name          string     `json:"channel_name"`
	Paused        bool       `json:"paused"`
	Depth         int64      `json:"depth"`
	BackendDepth  int64      `json:"backend_depth"`
	MessageCount  uint64     `json:"message_count"`
	InFlightCount int        `json:"in_flight_count"`
	DeferredCount int        `json:"deferred_count"`
	RequeueCount  uint64     `json:"requeue_count"`
	TimeoutCount  uint64     `json:"timeout_count"`
	E2eLatency    e2elatency `json:"e2e_processing_latency"`
	Clients       []*client  `json:"clients"`
}

type e2elatency struct {
	Count       int                  `json:"count"`
	Percentiles []map[string]float64 `json:"percentiles"`
}

func (e *e2elatency) percentileValue(idx int) float64 {
	if idx >= len(e.Percentiles) {
		return 0
	}
	return e.Percentiles[idx]["value"]
}

type client struct {
	ID            string `json:"client_id"`
	Hostname      string `json:"hostname"`
	Version       string `json:"version"`
	RemoteAddress string `json:"remote_address"`
	State         int32  `json:"state"`
	FinishCount   uint64 `json:"finish_count"`
	MessageCount  uint64 `json:"message_count"`
	ReadyCount    int64  `json:"ready_count"`
	InFlightCount int64  `json:"in_flight_count"`
	RequeueCount  uint64 `json:"requeue_count"`
	ConnectTime   int64  `json:"connect_ts"`
	SampleRate    int32  `json:"sample_rate"`
	Deflate       bool   `json:"deflate"`
	Snappy        bool   `json:"snappy"`
	TLS           bool   `json:"tls"`
}

func getPercentile(t *topic, percentile int) float64 {
	if len(t.E2eLatency.Percentiles) > 0 {
		if percentile == 99 {
			return t.E2eLatency.Percentiles[0]["value"]
		} else if percentile == 95 {
			return t.E2eLatency.Percentiles[1]["value"]
		}
	}
	return 0
}

func getNsqdStats(client *http.Client, nsqdURL string) (*stats, error) {
	resp, err := client.Get(nsqdURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sr stats
	if err = json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}
	return &sr, nil
}
