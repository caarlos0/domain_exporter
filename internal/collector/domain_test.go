package collector

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/caarlos0/domain_exporter/internal/client"
	"github.com/caarlos0/domain_exporter/internal/rdap"
	"github.com/caarlos0/domain_exporter/internal/safeconfig"
	"github.com/caarlos0/domain_exporter/internal/whois"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

func TestCollectorError(t *testing.T) {
	multi := client.NewMultiClient(rdap.NewClient(), whois.NewClient())
	testCollector(t, NewDomainCollector(multi, time.Second, safeconfig.Domain{Name: "fake.foo", Host: ""}), func(t *testing.T, status int, body string) {
		require.Equal(t, 200, status)
		require.Contains(t, body, "domain_probe_success{domain=\"fake.foo\"} 0")
		require.Contains(t, body, "domain_expiry_days{domain=\"fake.foo\"} -1")
	})
}

func TestNotExpired(t *testing.T) {
	multi := client.NewMultiClient(rdap.NewClient(), whois.NewClient())
	testCollector(
		t,
		NewDomainCollector(multi, time.Second, safeconfig.Domain{Name: "goreleaser.com", Host: ""}),
		func(t *testing.T, status int, body string) {
			t.Log(body)
			if strings.Contains(body, "domain_probe_success{domain=\"goreleaser.com\"} 0") {
				t.Skip("request failed")
				return
			}
			require.Equal(t, 200, status)
			require.Contains(t, body, "domain_probe_success{domain=\"goreleaser.com\"} 1")
			require.Regexp(t, `domain_expiry_days{domain=\"goreleaser.com\"} \d+`, body)
		},
	)
}

func testCollector(t *testing.T, collector prometheus.Collector, checker func(t *testing.T, status int, body string)) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	srv := httptest.NewServer(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	checker(t, resp.StatusCode, string(body))
}
