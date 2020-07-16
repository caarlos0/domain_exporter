package collector

import (
	"github.com/caarlos0/domain_exporter/client"
	"github.com/caarlos0/domain_exporter/rdapclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"math"
	"sync"
	"time"
)

type domainCollector struct {
	mutex      sync.Mutex
	client     client.Client
	rdapClient rdapclient.RdapClient
	domain     string

	expiryDays    *prometheus.Desc
	probeSuccess  *prometheus.Desc
	probeDuration *prometheus.Desc
}

// NewDomainCollector returns a domain collector.
func NewDomainCollector(client client.Client, rdapClient rdapclient.RdapClient, domain string) prometheus.Collector {
	const namespace = "domain"
	const subsystem = ""
	return &domainCollector{
		client:     client,
		rdapClient: rdapClient,
		domain:     domain,
		expiryDays: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "expiry_days"),
			"time in days until the domain expires",
			nil,
			nil,
		),
		probeSuccess: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "probe_success"),
			"wether the probe was successful or not",
			nil,
			nil,
		),
		probeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "probe_duration_seconds"),
			"returns how long the probe took to complete in seconds",
			nil,
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
	var start = time.Now()
	date, err := c.rdapClient.ExpireTime(c.domain)
	if err != nil {
		date, err = c.client.ExpireTime(c.domain)
		if err != nil {
			log.Errorf("failed to probe %s: %v", c.domain, err)
		}
	}

	var success = err == nil
	ch <- prometheus.MustNewConstMetric(
		c.probeSuccess,
		prometheus.GaugeValue,
		boolToFloat(success),
	)
	ch <- prometheus.MustNewConstMetric(
		c.expiryDays,
		prometheus.GaugeValue,
		math.Floor(time.Until(date).Hours()/24),
	)
	ch <- prometheus.MustNewConstMetric(
		c.probeDuration,
		prometheus.GaugeValue,
		time.Since(start).Seconds(),
	)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
