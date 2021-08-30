package collector

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/caarlos0/domain_exporter/internal/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type domainCollector struct {
	mutex   sync.Mutex
	client  client.Client
	domains []string
	timeout time.Duration

	expiryDays    *prometheus.Desc
	probeSuccess  *prometheus.Desc
	probeDuration *prometheus.Desc
}

// NewDomainCollector returns a domain collector.
func NewDomainCollector(client client.Client, domains ...string) prometheus.Collector {
	const namespace = "domain"
	const subsystem = ""
	return &domainCollector{
		client:  client,
		domains: domains,
		timeout: time.Second * 10,
		expiryDays: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "expiry_days"),
			"time in days until the domain expires",
			[]string{"domain"},
			nil,
		),
		probeSuccess: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "probe_success"),
			"wether the probe was successful or not",
			[]string{"domain"},
			nil,
		),
		probeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "probe_duration_seconds"),
			"returns how long the probe took to complete in seconds",
			[]string{"domain"},
			nil,
		),
	}
}

// Describe all metrics
func (c *domainCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.expiryDays
	ch <- c.probeDuration
	ch <- c.probeSuccess
}

// Collect all metrics
func (c *domainCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	for _, domain := range c.domains {
		start := time.Now()
		date, err := c.client.ExpireTime(ctx, domain)
		if err != nil {
			log.Error().Err(err).Msgf("failed to probe %s", domain)
		}

		success := err == nil
		ch <- prometheus.MustNewConstMetric(
			c.probeSuccess,
			prometheus.GaugeValue,
			boolToFloat(success),
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			c.expiryDays,
			prometheus.GaugeValue,
			math.Floor(time.Until(date).Hours()/24),
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			c.probeDuration,
			prometheus.GaugeValue,
			time.Since(start).Seconds(),
			domain,
		)
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
