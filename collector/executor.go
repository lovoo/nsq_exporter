package collector

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// NsqExecutor collects all NSQ metrics from the registered collectors.
// This type implements the prometheus.Collector interface and can be
// registered in the metrics collection.
//
// The executor takes the time needed for scraping nsqd stat endpoint and
// provides an extra metric for this. This metric is labeled with the
// scrape result ("success" or "error").
type NsqExecutor struct {
	nsqdURL string

	collectors []StatsCollector
	summary    *prometheus.SummaryVec
	client     *http.Client
	mutex      sync.RWMutex
}

// NewNsqExecutor creates a new executor for collecting NSQ metrics.
func NewNsqExecutor(namespace, nsqdURL, tlsCACert, tlsCert, tlsKey string) (*NsqExecutor, error) {
	sum := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: "exporter",
		Name:      "scrape_duration_seconds",
		Help:      "Duration of a scrape job of the NSQ exporter",
	}, []string{"result"})
	prometheus.MustRegister(sum)

	transport := &http.Transport{}
	if tlsCert != "" && tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if tlsCACert != "" {
			caCert, err := ioutil.ReadFile(tlsCACert)
			if err != nil {
				return nil, err
			}
			caCertPool.AppendCertsFromPEM(caCert)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		transport.TLSClientConfig = tlsConfig
	}
	return &NsqExecutor{
		nsqdURL: nsqdURL,
		summary: sum,
		client:  &http.Client{Transport: transport},
	}, nil
}

// Use configures a specific stats collector, so the stats could be
// exposed to the Prometheus system.
func (e *NsqExecutor) Use(c StatsCollector) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.collectors = append(e.collectors, c)
}

// Describe implements the prometheus.Collector interface.
func (e *NsqExecutor) Describe(ch chan<- *prometheus.Desc) {
	for _, c := range e.collectors {
		c.describe(ch)
	}
}

// Collect implements the prometheus.Collector interface.
func (e *NsqExecutor) Collect(out chan<- prometheus.Metric) {
	start := time.Now()
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// reset state, because metrics can gone
	for _, c := range e.collectors {
		c.reset()
	}

	stats, err := getNsqdStats(e.client, e.nsqdURL)
	tScrape := time.Since(start).Seconds()

	fmt.Printf("[nsq_exporter] url %s, err %v, get NSQ stats %v\n", e.nsqdURL, err, stats)

	result := "success"
	if err != nil {
		result = "error"
	}

	e.summary.WithLabelValues(result).Observe(tScrape)

	if err == nil {
		for _, c := range e.collectors {
			c.set(stats)
		}
		for _, c := range e.collectors {
			c.collect(out)
		}
	}
}
