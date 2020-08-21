package collector

import (

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/caarlos0/domain_exporter/client"
	"github.com/caarlos0/domain_exporter/rdapclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

func TestCollectorError(t *testing.T) {
	var client = client.NewWhoisClient()
	var rdapClient = rdapclient.NewRdapClient()
	testCollector(t, NewDomainCollector(client, rdapClient, "fake.foo"), func(t *testing.T, status int, body string) {
		require.Equal(t, 200, status)
		require.Contains(t, body, "domain_probe_success 0")
		require.Contains(t, body, "domain_expiry_days -1")
	})
}

func TestNotExpired(t *testing.T) {
	var client = client.NewWhoisClient()
	var rdapClient = rdapclient.NewRdapClient()
	testCollector(t, NewDomainCollector(client, rdapClient, "goreleaser.com"), func(t *testing.T, status int, body string) {
		require.Equal(t, 200, status)
		require.Contains(t, body, "domain_probe_success 1")
		require.Regexp(t, regexp.MustCompile("domain_expiry_days \\d+"), body)
	})
}

func testCollector(t *testing.T, collector prometheus.Collector, checker func(t *testing.T, status int, body string)) {
	var registry = prometheus.NewRegistry()
	registry.MustRegister(collector)

	var srv = httptest.NewServer(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	checker(t, resp.StatusCode, string(body))
}
