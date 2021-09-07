package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/domain_exporter/internal/client"
	"github.com/caarlos0/domain_exporter/internal/collector"
	"github.com/caarlos0/domain_exporter/internal/rdap"
	"github.com/caarlos0/domain_exporter/internal/refresher"
	"github.com/caarlos0/domain_exporter/internal/safeconfig"
	"github.com/caarlos0/domain_exporter/internal/whois"
	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// nolint: gochecknoglobals
var (
	bind       = kingpin.Flag("bind", "addr to bind the server").Short('b').Default(":9222").String()
	debug      = kingpin.Flag("debug", "show debug logs").Default("false").Bool()
	format     = kingpin.Flag("logFormat", "log format to use").Default("console").Enum("json", "console")
	interval   = kingpin.Flag("cache", "time to cache the result of whois calls").Default("2h").Duration()
	configFile = kingpin.Flag("config", "configuration file").String()
	version    = "master"
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
	cfg, err := safeconfig.New(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("error to create config")
	}

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache := cache.New(*interval, *interval)
	cli := client.NewMultiClient(rdap.NewClient(), whois.NewClient())
	cachedClient := client.NewCachedClient(cli, cache)

	if len(cfg.Domains) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			refresher.New(*interval, cachedClient, cfg.Domains...).Run(ctx)
		}()

		domainCollector := collector.NewDomainCollector(cachedClient, cfg.Domains...)
		prometheus.DefaultRegisterer.MustRegister(domainCollector)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probeHandler(cli))
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

	if err := runServerWithGracefullyShutdown(wg); err != nil {
		log.Fatal().Err(err).Msg("error starting server")
	}

	log.Info().Msg("domain exporter is finished")
}

func runServerWithGracefullyShutdown(wg *sync.WaitGroup) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	signal.Notify(signalChan, syscall.SIGINT)

	server := &http.Server{Addr: *bind}

	wg.Add(1)
	go func() {
		defer wg.Done()

		sig := <-signalChan

		log.Warn().Msgf("got %s signal. Shutdown", sig)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed to shutdown http server")
		}
	}()

	log.Info().Msgf("listening on %s", *bind)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
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
