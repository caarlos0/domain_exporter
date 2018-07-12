package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/domainr/whois"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

var (
	bind    = kingpin.Flag("bind", "addr to bind the server").Short('b').Default(":9222").String()
	debug   = kingpin.Flag("debug", "show debug logs").Default("false").Bool()
	version = "master"

	storage = cache.New(60*time.Minute, 120*time.Minute)

	re = regexp.MustCompile(`(?i)(Registry Expiry Date|paid-till|Expiration Date|Expiry.*|expires.*|Expires|expire):\s+(.*)`)

	formats = []string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		"20060102",               // .com.br
		"2006-01-02",             // .lt
		"02.01.2006",             // .cz
		"2006-01-02 15:04:05-07", // .ua
	}
)

func main() {
	kingpin.Version("domain_exporter version " + version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	if *debug {
		_ = log.Base().SetLevel("debug")
		log.Debug("debug mode enabled")
	}

	log.Info("starting domain_exporter:", version)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probeHandler)
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
	log.Info("listening on port:", *bind)
	if err := http.ListenAndServe(*bind, nil); err != nil {
		log.Fatalf("error starting server: %s", err)
	}
}

func probeHandler(w http.ResponseWriter, r *http.Request) {

	var params = r.URL.Query()
	var target = params.Get("target")
	var registry = prometheus.NewRegistry()
	var start = time.Now()
	var expiryGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_dns_domain_expiry",
		Help: "Date when the domain expires",
	})
	var probeDurationGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_dns_duration_seconds",
		Help: "Duration of whois request in seconds",
	})
	var probeSuccessGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_dns_success",
		Help: "Displays whether or not the probe was a success",
	})
	registry.Register(expiryGauge)
	registry.MustRegister(probeDurationGauge)
	registry.MustRegister(probeSuccessGauge)
	probeSuccessGauge.Set(0)
	if target == "" {
		http.Error(w, "target parameter is missing", http.StatusBadRequest)
		return
	}
	record, found := storage.Get(target)
	if found {
		log.Info("found cache entory for domain ", target)
		expiryGauge.Set(record.(float64))
		probeSuccessGauge.Set(1)
	} else {
		req, err := whois.NewRequest(target)
		if err != nil {
			log.Error("error processing domain ", target, "\n", err.Error())
		} else {
			resp, err := whois.DefaultClient.Fetch(req)
			if err != nil {
				log.Error("error processing domain ", target, "\n", err.Error())
			} else {
				date, err := extractDate(string(resp.Body))
				if err != nil {
					log.Error("error processing domain ", target, "\n", err.Error())
				} else {
					expiryGauge.Set(float64(date.Unix()))
					storage.Set(target, float64(date.Unix()), cache.DefaultExpiration)
					probeSuccessGauge.Set(1)
				}
			}
		}
	}
	probeDurationGauge.Set(time.Since(start).Seconds())
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

func extractDate(body string) (time.Time, error) {
	var result = re.FindStringSubmatch(body)
	if len(result) < 2 {
		return time.Time{}, fmt.Errorf("could not parse whois response: %s", body)
	}
	var dateStr = strings.TrimSpace(result[2])
	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}
