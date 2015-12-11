package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/lovoo/nsq_exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// Version of nsq_exporter. Set at build time.
const Version = "0.0.0.dev"

var (
	listenAddress     = flag.String("web.listen", ":9117", "Address on which to expose metrics and web interface.")
	metricsPath       = flag.String("web.path", "/metrics", "Path under which to expose metrics.")
	nsqUrl            = flag.String("nsq.addr", "http://localhost:4151/stats", "Address of the NSQ host.")
	enabledCollectors = flag.String("collectors", "nsqstats", "Comma-separated list of collectors to use.")
	namespace         = flag.String("namespace", "nsq", "Namespace for the NSQ metrics.")

	collectorRegistry = map[string]func(name string, x *collector.NsqExecutor) error{
		"nsqstats": addStatsCollector,
	}
)

func main() {
	flag.Parse()

	ex, err := createNsqExecutor()
	if err != nil {
		log.Fatalf("error creating nsq executor: %v", err)
	}
	prometheus.MustRegister(ex)

	handler := prometheus.Handler()
	if *metricsPath == "" || *metricsPath == "/" {
		http.Handle(*metricsPath, handler)
	} else {
		http.Handle(*metricsPath, handler)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
			<head><title>NSQ Exporter</title></head>
			<body>
			<h1>NSQ Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		})
	}

	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func createNsqExecutor() (*collector.NsqExecutor, error) {
	ex := collector.NewNsqExecutor(*namespace)
	for _, name := range strings.Split(*enabledCollectors, ",") {
		name = strings.TrimSpace(name)
		addCollector, has := collectorRegistry[name]
		if !has {
			return nil, fmt.Errorf("unknown collector: %s", name)
		}

		if err := addCollector(name, ex); err != nil {
			return nil, err
		}
	}
	return ex, nil
}

func addStatsCollector(name string, ex *collector.NsqExecutor) error {
	u, err := url.Parse(normalizeURL(*nsqUrl))
	if err != nil {
		return err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Path == "" {
		u.Path = "/stats"
	}
	u.RawQuery = "format=json"
	ex.AddCollector(name, collector.NewStatsCollector(u.String()))
	return nil
}

func normalizeURL(u string) string {
	u = strings.ToLower(u)
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
		return "http://" + u
	}
	return u
}
