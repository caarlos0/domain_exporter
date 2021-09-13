package whois

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWhoisParsing(t *testing.T) {
	for _, tt := range []struct {
		domain string
		err    string
	}{
		{domain: "google.ai", err: "could not parse whois response"},
		{domain: "domreg.lt", err: ""},
		{domain: "fakedomain.foo", err: "could not parse whois response: Domain not found"},
		{domain: "google.cn", err: ""},
		{domain: "google.com", err: ""},
		{domain: "google.de", err: "could not parse whois response"},
		{domain: "nic.ua", err: ""},
		{domain: "google.com.tw", err: ""},
		{domain: "bbc.co.uk", err: ""},
		{domain: "google.sk", err: ""},
		{domain: "google.ro", err: ""},
		{domain: "google.pt", err: "i/o timeout"},
		{domain: "google.it", err: ""},
		{domain: "watchub.pw", err: ""},
		{domain: "google.co.id", err: ""},
		{domain: "google.kr", err: ""},
		{domain: "google.jp", err: ""},
		{domain: "microsoft.im", err: ""},
		{domain: "google.rs", err: ""},
		{domain: "мвд.рф", err: ""},
		{domain: "МВД.РФ", err: ""},
		{domain: "GOOGLE.RS", err: ""},
	} {
		tt := tt
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			expiry, err := NewClient().ExpireTime(ctx, tt.domain)
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
