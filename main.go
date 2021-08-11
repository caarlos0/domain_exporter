package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/domain_exporter/internal/client"
	"github.com/caarlos0/domain_exporter/internal/collector"
	"github.com/caarlos0/domain_exporter/internal/rdap"
	"github.com/caarlos0/domain_exporter/internal/whois"
	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// nolint: gochecknoglobals
var (
	bind     = kingpin.Flag("bind", "addr to bind the server").Short('b').Default(":9222").String()
	debug    = kingpin.Flag("debug", "show debug logs").Default("false").Bool()
	format   = kingpin.Flag("logFormat", "log format to use").Default("console").Enum("json", "console")
	interval = kingpin.Flag("cache", "time to cache the result of whois calls").Default("2h").Duration()
	version  = "master"
)

func main() {
	kingpin.Version("domain_exporter version " + version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	urlPrefix, urlPrefixOK := os.LookupEnv("DOMAIN_EXPORTER_URL_PREFIX")
	if !urlPrefixOK {
		urlPrefix = ""
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("enabled debug mode")
	}

	log.Info().Msgf("starting domain_exporter %s", version)
	cache := cache.New(*interval, *interval)
	whoisClient := client.NewCachedClient(whois.NewClient(), cache)
	rdapClient := client.NewCachedClient(rdap.NewClient(), cache)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probeHandler(client.NewMultiClient(rdapClient, whoisClient)))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(
			w, `
			<html>
			<head><title>Domain Exporter</title></head>
			<body>
				<h1>Domain Exporter</h1>
				<p><a href="%[1]s/metrics">Metrics</a></p>
				<p><a href="%[1]s/probe?target=google.com">probe google.com</a></p>
			</body>
			</html>
			`, urlPrefix,
		)
	})
	log.Info().Msgf("listening on %s", *bind)
	if err := http.ListenAndServe(*bind, nil); err != nil {
		log.Fatal().Err(err).Msg("error starting server")
	}
}

func probeHandler(cli client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		target := strings.Replace(params.Get("target"), "www.", "", 1)
		if target == "" {
			log.Error().Msg("target parameter missing")
			http.Error(w, "target parameter is missing", http.StatusBadRequest)
			return
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(collector.NewDomainCollector(cli, target))

		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}
}
