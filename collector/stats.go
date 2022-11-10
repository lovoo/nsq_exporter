package collector

import (
	"encoding/json"
	"net/http"
)

type stats struct {
	Version   string      `json:"version"`
	Health    string      `json:"health"`
	StartTime int64       `json:"start_time"`
	Topics    []*topic    `json:"topics"`
	Memory    memory      `json:"memory"`
	Producers []*producer `json:"producers"`
}

type producer struct {
	ClientId                      string      `json:"client_id"`
	Hostname                      string      `json:"hostname"`
	Version                       string      `json:"version"`
	RemoteAddress                 string      `json:"remote_address"`
	State                         int         `json:"state"`
	ReadyCount                    int         `json:"ready_count"`
	InFlightCount                 int         `json:"in_flight_count"`
	MessageCount                  int64       `json:"message_count"`
	FinishCount                   int64       `json:"finish_count"`
	RequeueCount                  int64       `json:"requeue_count"`
	ConnectTS                     int64       `json:"connect_ts"`
	SampleRate                    int         `json:"sample_rate"`
	Deflate                       bool        `json:"deflate"`
	Snappy                        bool        `json:"snappy"`
	UserAgent                     string      `json:"user_agent"`
	PubCounts                     []*pubCount `json:"pub_counts"`
	TLS                           bool        `json:"tls"`
	TLSCipherSuite                string      `json:"tls_cipher_suite"`
	TLSVersion                    string      `json:"tls_version"`
	TLSNegotiatedProtocol         string      `json:"tls_negotiated_protocol"`
	TLSNegotiatedProtocolIsMutual bool        `json:"tls_negotiated_protocol_is_mutual"`
}

type pubCount struct {
	Topic string `json:"topic"`
	Count int64  `json:"count"`
}

type memory struct {
	HeapObjects       int64 `json:"heap_objects"`
	HeapIdleBytes     int64 `json:"heap_idle_bytes"`
	HeapInUseBytes    int64 `json:"heap_in_use_bytes"`
	HeapReleasedBytes int64 `json:"heap_released_bytes"`
	GcPauseUsec100    int64 `json:"gc_pause_usec_100"`
	GcPauseUsec99     int64 `json:"gc_pause_usec_99"`
	GcPauseUsec95     int64 `json:"gc_pause_usec_95"`
	NextGcBytes       int64 `json:"next_gc_bytes"`
	GcTotalRuns       int64 `json:"gc_total_runs"`
}

type topic struct {
	Name         string     `json:"topic_name"`
	Channels     []*channel `json:"channels"`
	Depth        int64      `json:"depth"`
	BackendDepth int64      `json:"backend_depth"`
	MessageCount uint64     `json:"message_count"`
	MessageBytes uint64     `json:"message_bytes"`
	Paused       bool       `json:"paused"`
	E2eLatency   e2elatency `json:"e2e_processing_latency"`
}

type channel struct {
	Name          string     `json:"channel_name"`
	Depth         int64      `json:"depth"`
	BackendDepth  int64      `json:"backend_depth"`
	InFlightCount int        `json:"in_flight_count"`
	DeferredCount int        `json:"deferred_count"`
	MessageCount  uint64     `json:"message_count"`
	RequeueCount  uint64     `json:"requeue_count"`
	TimeoutCount  uint64     `json:"timeout_count"`
	ClientCount   uint64     `json:"client_count"`
	Clients       []*client  `json:"clients"`
	Paused        bool       `json:"paused"`
	E2eLatency    e2elatency `json:"e2e_processing_latency"`
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
	TLS                           bool   `json:"tls"`
	TLSCipherSuite                string `json:"tls_cipher_suite"`
	TLSVersion                    string `json:"tls_version"`
	TLSNegotiatedProtocol         string `json:"tls_negotiated_protocol"`
	TLSNegotiatedProtocolIsMutual bool   `json:"tls_negotiated_protocol_is_mutual"`
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

	var st stats
	if err = json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return nil, err
	}
	return &st, nil
}
