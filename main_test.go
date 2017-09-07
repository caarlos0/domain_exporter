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
)

var srv *httptest.Server

func TestMain(m *testing.M) {
	srv = httptest.NewServer(http.HandlerFunc(probeHandler))
	defer srv.Close()
	m.Run()
}

func TestSuccessQueries(t *testing.T) {
	var re = regexp.MustCompile("(?m)^domain_expiry_days ([0-9]+)$")
	for _, domain := range []string{
		"carlosbecker.com",
		"carinebecker.com",
		"watchub.pw",
		"google.com",
	} {
		t.Run(domain, func(t *testing.T) {
			var assert = assert.New(t)
			var body = req(t, domain, http.StatusOK)
			var results = re.FindStringSubmatch(string(body))
			assert.Len(results, 2, "should have returned the metric in the body of the response")
			days, err := strconv.Atoi(results[1])
			assert.NoError(err)
			assert.True(days > 0, "domain must not be expired")
		})
	}
}

func TestTargetNotProvided(t *testing.T) {
	req(t, "", http.StatusBadRequest)
}

func TestTargetDoesntExist(t *testing.T) {
	req(t, "this-domain-should-not-exist.blah", http.StatusBadRequest)
}

func req(t *testing.T, domain string, expectedStatus int) string {
	var assert = assert.New(t)
	resp, err := http.Get(fmt.Sprintf("%s?target=%s", srv.URL, domain))
	assert.NoError(err)
	assert.Equal(expectedStatus, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)
	return string(body)
}
