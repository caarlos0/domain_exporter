package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessQueries(t *testing.T) {
	var srv = httptest.NewServer(http.HandlerFunc(probeHandler))
	defer srv.Close()

	var req = func(t *testing.T, domain string, expectedStatus int) string {
		var assert = assert.New(t)
		resp, err := http.Get(fmt.Sprintf("%s?target=%s", srv.URL, domain))
		assert.NoError(err)
		assert.Equal(expectedStatus, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(err)
		return string(body)
	}

	t.Run("valid domain", func(t *testing.T) {
		var re = regexp.MustCompile("(?m)^domain_expiry_days ([0-9]+)$")
		var assert = assert.New(t)
		var body = req(t, "google.com", http.StatusOK)
		var results = re.FindStringSubmatch(string(body))
		assert.Len(results, 2, "should have returned the metric in the body of the response")
		days, err := strconv.Atoi(results[1])
		assert.NoError(err)
		assert.True(days > 0, "domain must not be expired")
	})
	t.Run("target not provided", func(t *testing.T) {
		req(t, "", http.StatusBadRequest)
	})
	t.Run("target tld does not exist", func(t *testing.T) {
		req(t, "this-domain-should-not-exist.blah", http.StatusBadRequest)
	})
	t.Run("target does not exist", func(t *testing.T) {
		req(t, "this-domain-should-not-exist.com", http.StatusInternalServerError)
	})
}

func TestWhoisParsing(t *testing.T) {
	for _, tt := range []struct {
		domain string
		err    string
	}{
		{domain: "domreg.lt", err: ""},
		{domain: "fakedomain.foo", err: "could not parse date"},
		{domain: "google.com", err: ""},
		{domain: "google.de", err: "could not parse whois response"},
		{domain: "nic.ua", err: ""},
		{domain: "watchub.pw", err: ""},
	} {
		tt := tt
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()
			bts, err := ioutil.ReadFile("testdata/" + tt.domain)
			require.NoError(t, err)
			days, err := extractDays(string(bts))
			if tt.err == "" {
				require.NoError(t, err)
				require.True(t, (days > 0), "domain must not be expired")
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
				require.Equal(t, 0.0, days)
			}
		})
	}
}
