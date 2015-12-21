package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tsne/nsq_exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
)

// Version of nsq_exporter. Set at build time.
const Version = "0.0.0.dev"

var (
	listenAddress     = flag.String("web.listen", ":9117", "Address on which to expose metrics and web interface.")
	metricsPath       = flag.String("web.path", "/metrics", "Path under which to expose metrics.")
	nsqdURL           = flag.String("nsqd.addr", "http://localhost:4151/stats", "Address of the nsqd node.")
	enabledCollectors = flag.String("collect", "stats.topics,stats.channels", "Comma-separated list of collectors to use.")
	namespace         = flag.String("namespace", "nsq", "Namespace for the NSQ metrics.")

	collectorRegistry = map[string]func(names []string) (collector.Collector, error){
		"stats": createNsqdStats,
	}

	// stats.* collectors
	statsRegistry = map[string]func(namespace string) collector.StatsCollector{
		"topics":   collector.TopicsCollector,
		"channels": collector.ChannelsCollector,
		"clients":  collector.ClientsCollector,
	}
)

func main() {
	flag.Parse()

	ex, err := createNsqExecutor()
	if err != nil {
		log.Fatalf("error creating nsq executor: %v", err)
	}
	prometheus.MustRegister(ex)

	http.Handle(*metricsPath, prometheus.Handler())
	if *metricsPath != "" && *metricsPath != "/" {
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

	log.Print("listening to ", *listenAddress)
	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func createNsqExecutor() (*collector.NsqExecutor, error) {
	collectors := make(map[string][]string)
	for _, name := range strings.Split(*enabledCollectors, ",") {
		name = strings.TrimSpace(name)
		parts := strings.SplitN(name, ".", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid collector name: %s", name)
		}
		collectors[parts[0]] = append(collectors[parts[0]], parts[1])
	}

	ex := collector.NewNsqExecutor(*namespace)
	for collector, subcollectors := range collectors {
		newCollector, has := collectorRegistry[collector]
		if !has {
			return nil, fmt.Errorf("invalid collector: %s", collector)
		}

		c, err := newCollector(subcollectors)
		if err != nil {
			return nil, err
		}
		ex.AddCollector(collector, c)
	}
	return ex, nil
}

func createNsqdStats(statsCollectors []string) (collector.Collector, error) {
	nsqdURL, err := normalizeURL(*nsqdURL)
	if err != nil {
		return nil, err
	}

	stats := collector.NewNsqdStats(*namespace, nsqdURL)
	for _, c := range statsCollectors {
		newStatsCollector, has := statsRegistry[c]
		if !has {
			return nil, fmt.Errorf("unknown stats collector: %s", c)
		}
		stats.Use(newStatsCollector(*namespace))
	}
	return stats, nil
}

func normalizeURL(ustr string) (string, error) {
	ustr = strings.ToLower(ustr)
	if !strings.HasPrefix(ustr, "https://") && !strings.HasPrefix(ustr, "http://") {
		ustr = "http://" + ustr
	}

	u, err := url.Parse(ustr)
	if err != nil {
		return "", err
	}
	if u.Path == "" {
		u.Path = "/stats"
	}
	u.RawQuery = "format=json"
	return u.String(), nil
}
