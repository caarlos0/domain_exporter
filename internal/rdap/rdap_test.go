package rdap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRdapParsing(t *testing.T) {
	for _, tt := range []struct {
		domain string
		err    string
	}{
		{domain: "domreg.lt", err: "No RDAP servers found for 'domreg.lt'"},
		{domain: "fakedomain.foo", err: "RDAP server returned 404, object does not exist."},
		{domain: "google.cn", err: "No RDAP servers found for 'google.cn'"},
		{domain: "google.com", err: ""},
		{domain: "google.de", err: "No RDAP servers found for 'google.de'"},
		{domain: "nic.ua", err: "No RDAP servers found for 'nic.ua'"},
		{domain: "taiwannews.com.tw", err: "No RDAP servers found for 'taiwannews.com.tw'"},
		{domain: "bbc.co.uk", err: "No RDAP servers found for 'bbc.co.uk'"},
		{domain: "google.sk", err: "No RDAP servers found for 'google.sk'"},
		{domain: "google.ro", err: "No RDAP servers found for 'google.ro'"},
		{domain: "watchub.pw", err: ""},
		{domain: "google.co.id", err: ""},
		{domain: "google.kr", err: "No RDAP servers found for 'google.kr'"},
	} {
		tt := tt
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()
			expiry, err := NewClient().ExpireTime(tt.domain)
			if tt.err == "" {
				require.NoError(t, err)
				require.True(t, time.Since(expiry).Hours() < 0, "domain must not be expired")
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			}
		})
	}
}
