package rdap

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRdapParsing(t *testing.T) {
	for _, tt := range []struct {
		domain string
		err    string
	}{
		// {domain: "google.ai", err: "No RDAP servers found for 'google.ai'"},
		{domain: "domreg.lt", err: "No RDAP servers found for 'domreg.lt'"},
		{domain: "fakedomain.foo", err: "RDAP server returned 404, object does not exist."},
		{domain: "google.cn", err: "No RDAP servers found for 'google.cn'"},
		{domain: "google.com", err: ""},
		{domain: "google.lu", err: "No RDAP servers found for 'google.lu'"},
		{domain: "google.de", err: "No RDAP servers found for 'google.de'"},
		{domain: "nic.ua", err: ""},
		{domain: "taiwannews.com.tw", err: ""},
		// {domain: "bbc.co.uk", err: "No RDAP servers found for 'bbc.co.uk'"},
		{domain: "google.sg", err: "No RDAP servers found for 'google.sg'"},
		{domain: "google.sk", err: "No RDAP servers found for 'google.sk'"},
		{domain: "google.ro", err: "No RDAP servers found for 'google.ro'"},
		{domain: "google.pw", err: ""},
		// {domain: "google.co.id", err: ""}, // random failures
		{domain: "google.kr", err: "No RDAP servers found for 'google.kr'"},
		{domain: "google.host", err: ""},
	} {
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()
			expiry, err := NewClient().ExpireTime(context.Background(), tt.domain, "")
			if tt.err == "" {
				require.NoError(t, err)
				require.Less(t, time.Since(expiry).Hours(), 0.0)
			} else {
				require.ErrorContains(t, err, tt.err)
			}
		})
	}
}
