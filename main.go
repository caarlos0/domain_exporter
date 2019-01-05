package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/domain_exporter/client"
	"github.com/caarlos0/domain_exporter/collector"
	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

// nolint: gochecknoglobals
var (
	bind     = kingpin.Flag("bind", "addr to bind the server").Short('b').Default(":9222").String()
	debug    = kingpin.Flag("debug", "show debug logs").Default("false").Bool()
	interval = kingpin.Flag("refresh.interval", "time between refreshes with whois api").Default("30m").Duration()
	version  = "master"
)

func main() {
	kingpin.Version("domain_exporter version " + version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if *debug {
		_ = log.Base().SetLevel("debug")
		log.Debug("enabled debug mode")
	}

	log.Info("starting domain_exporter", version)
	var cache = cache.New(*interval, *interval)
	var cli = client.NewCachedClient(client.NewWhoisClient(), cache)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probeHandler(cli))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(
			w, `
			<html>
			<head><title>Domain Exporter</title></head>
			<body>
				<h1>Domain Exporter</h1>
				<p><a href="/metrics">Metrics</a></p>
				<p><a href="/probe?target=google.com">probe google.com</a></p>
			</body>
			</html>
			`,
		)
	})
	log.Info("listening on", *bind)
	if err := http.ListenAndServe(*bind, nil); err != nil {
		log.Fatalf("error starting server: %s", err)
	}
}

func probeHandler(cli client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params = r.URL.Query()
		var target = strings.Replace(params.Get("target"), "www.", "", 1)
		if target == "" {
			log.Error("target parameter missing")
			http.Error(w, "target parameter is missing", http.StatusBadRequest)
			return
		}

		var registry = prometheus.NewRegistry()
		registry.MustRegister(collector.NewDomainCollector(cli, target))
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}
}
