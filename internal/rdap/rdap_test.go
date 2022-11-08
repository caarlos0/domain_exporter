package rdap

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestRdapParsing(t *testing.T) {
	for _, tt := range []struct {
		domain string
		err    string
	}{
		{domain: "google.ai", err: "No RDAP servers found for 'google.ai'"},
		{domain: "domreg.lt", err: "No RDAP servers found for 'domreg.lt'"},
		{domain: "fakedomain.foo", err: "RDAP server returned 404, object does not exist."},
		{domain: "google.cn", err: "No RDAP servers found for 'google.cn'"},
		{domain: "google.com", err: ""},
		{domain: "google.de", err: "No RDAP servers found for 'google.de'"},
		{domain: "nic.ua", err: "No RDAP servers found for 'nic.ua'"},
		{domain: "taiwannews.com.tw", err: "No RDAP servers found for 'taiwannews.com.tw'"},
		// {domain: "bbc.co.uk", err: "No RDAP servers found for 'bbc.co.uk'"},
		{domain: "google.sg", err: "No RDAP servers found for 'google.sg'"},
		{domain: "google.sk", err: "No RDAP servers found for 'google.sk'"},
		{domain: "google.ro", err: "No RDAP servers found for 'google.ro'"},
		{domain: "google.pw", err: ""},
		{domain: "google.co.id", err: ""},
		{domain: "google.kr", err: "No RDAP servers found for 'google.kr'"},
		{domain: "google.host", err: ""},
	} {
		tt := tt
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			expiry, err := NewClient().ExpireTime(context.Background(), tt.domain)
			if tt.err == "" {
				is.NoErr(err)                           // should not err
				is.True(time.Since(expiry).Hours() < 0) // domain must not be expired
			} else {
				is.True(err != nil)                            // should have errored
				is.True(strings.Contains(err.Error(), tt.err)) // should have error message
			}
		})
	}
}
