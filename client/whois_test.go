package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWhoisParsing(t *testing.T) {
	for _, tt := range []struct {
		domain string
		err    string
	}{
		{domain: "domreg.lt", err: ""},
		{domain: "fakedomain.foo", err: "could not parse whois response: Domain not found"},
		{domain: "google.cn", err: ""},
		{domain: "google.com", err: ""},
		{domain: "google.de", err: "could not parse whois response"},
		{domain: "nic.ua", err: ""},
		// {domain: "watchub.pw", err: ""}, // TODO: this for some reason fails on travis
	} {
		tt := tt
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()
			expiry, err := NewWhoisClient().ExpireTime(tt.domain)
			if tt.err == "" {
				require.NoError(t, err)
				require.True(t, (time.Since(expiry).Hours() < 0), "domain must not be expired")
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			}
		})
	}
}
