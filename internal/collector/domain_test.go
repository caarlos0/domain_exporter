package collector

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/caarlos0/domain_exporter/internal/client"
	"github.com/caarlos0/domain_exporter/internal/rdap"
	"github.com/caarlos0/domain_exporter/internal/safeconfig"
	"github.com/caarlos0/domain_exporter/internal/whois"
	"github.com/matryer/is"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestCollectorError(t *testing.T) {
	is := is.New(t)
	multi := client.NewMultiClient(rdap.NewClient(), whois.NewClient())
	testCollector(t, NewDomainCollector(multi, time.Second, safeconfig.Domain{Name: "fake.foo", Host: ""}), func(t *testing.T, status int, body string) {
		is := is.New(t)
		is.Equal(200, status)                                                          // request should succeed
		is.True(strings.Contains(body, "domain_probe_success{domain=\"fake.foo\"} 0")) // probe should succeed
		is.True(strings.Contains(body, "domain_expiry_days{domain=\"fake.foo\"} -1"))  // should contain domain expiry
	})
}

func TestNotExpired(t *testing.T) {
	is := is.New(t)
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
			is := is.New(t)
			is.Equal(200, status)                                                                                   // request should succeed
			is.True(strings.Contains(body, "domain_probe_success{domain=\"goreleaser.com\"} 1"))                    // probe should succeed
			is.True(regexp.MustCompile(`domain_expiry_days{domain=\"goreleaser.com\"} \d+`).FindString(body) != "") // should contain domain expiry
		},
	)
}

func testCollector(t *testing.T, collector prometheus.Collector, checker func(t *testing.T, status int, body string)) {
	is := is.New(t)
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	srv := httptest.NewServer(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	is.NoErr(err) // expected no error
	body, err := io.ReadAll(resp.Body)
	is.NoErr(err) // expected no error
	checker(t, resp.StatusCode, string(body))
}
