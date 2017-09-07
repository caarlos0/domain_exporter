package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/domainr/whois"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	bind    = kingpin.Flag("bind", "addr to bind the server").Default(":9222").String()
	version = "master"

	re = regexp.MustCompile(`(?i)(Registry Expiry Date|paid-till|Expiration Date|Expiry.*|expires.*): (.*)`)

	expiryGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_expiry_days",
			Help: "time in days until the domain expires",
		},
		[]string{"domain"},
	)
	probeDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "probe_duration_seconds",
			Help: "Returns how long the probe took to complete in seconds",
		},
		[]string{"domain"},
	)

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
		"20060102", // lol registro.br
	}
)

func main() {
	kingpin.Version("domain_exporter version " + version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Println("starting domain_exporter", version)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probeHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>Domain Exporter</title></head>
			<body>
				<h1>Domain Exporter</h1>
				<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>
		`))
	})
	log.Println("listening on", *bind)
	if err := http.ListenAndServe(*bind, nil); err != nil {
		log.Fatalf("error starting server: %s", err)
	}
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
	var params = r.URL.Query()
	var target = params.Get("target")
	var registry = prometheus.NewRegistry()
	var start = time.Now()
	registry.MustRegister(expiryGauge)
	registry.MustRegister(probeDurationGauge)
	if target == "" {
		http.Error(w, "target parameter is missing", http.StatusBadRequest)
		return
	}
	req, err := whois.NewRequest(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := whois.DefaultClient.Fetch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var result = re.FindStringSubmatch(resp.String())
	if len(result) < 2 {
		http.Error(
			w,
			fmt.Sprintf("couldnt parse whois for domain: %s", target),
			http.StatusInternalServerError,
		)
		return
	}
	var dateStr = strings.TrimSpace(result[2])
	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			var days = math.Floor(date.Sub(time.Now()).Hours() / 24)
			expiryGauge.WithLabelValues(target).Set(days)
			probeDurationGauge.WithLabelValues(target).Set(time.Since(start).Seconds())
			log.Printf("domain: %s, days: %v, date: %s\n", target, days, date)
			promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
			return
		}
	}
	http.Error(
		w,
		fmt.Sprintf("couldnt parse date from whois of domain: %s", target),
		http.StatusInternalServerError,
	)
}
