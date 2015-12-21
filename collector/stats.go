package collector

import (
	"encoding/json"
	"net/http"
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

// see https://github.com/nsqio/nsq/blob/master/nsqd/stats.go
type topic struct {
	Name         string     `json:"topic_name"`
	Paused       bool       `json:"paused"`
	Depth        int64      `json:"depth"`
	BackendDepth int64      `json:"backend_depth"`
	MessageCount uint64     `json:"message_count"`
	Channels     []*channel `json:"channels"`
}

type channel struct {
	Name          string    `json:"channel_name"`
	Paused        bool      `json:"paused"`
	Depth         int64     `json:"depth"`
	BackendDepth  int64     `json:"backend_depth"`
	MessageCount  uint64    `json:"message_count"`
	InFlightCount int       `json:"in_flight_count"`
	DeferredCount int       `json:"deferred_count"`
	RequeueCount  uint64    `json:"requeue_count"`
	TimeoutCount  uint64    `json:"timeout_count"`
	Clients       []*client `json:"clients"`
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

func getNsqdStats(nsqdURL string) (*stats, error) {
	resp, err := http.Get(nsqdURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sr statsResponse
	if err = json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}
	return &sr.Data, nil
}
