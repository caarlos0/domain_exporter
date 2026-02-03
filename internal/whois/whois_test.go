package whois

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWhoisParsing(t *testing.T) {
	for _, tt := range []struct {
		domain  string
		host    string
		err     string
		expired bool
	}{
		{domain: "google.ai", host: "", err: ""},
		{domain: "google.lt", host: "", err: ""},
		{domain: "fakedomain.foo", host: "", err: "no such host"},
		{domain: "google.cn", host: "", err: ""},
		{domain: "google.com", host: "", err: ""},
		{domain: "google.lu", host: "", err: "could not parse whois response"},
		{domain: "dns.lu", host: "", err: "could not parse whois response"},
		{domain: "google.de", host: "", err: "could not parse whois response"},
		{domain: "nic.ua", host: "", err: ""},
		{domain: "mod.gov.ua", host: "", err: ""},
		{domain: "google.com.tw", host: "", err: ""},
		{domain: "bbc.co.uk", host: "", err: ""},
		{domain: "google.sg", host: "", err: ""},
		{domain: "google.sk", host: "", err: ""},
		{domain: "google.ro", host: "", err: ""},
		{domain: "google.pt", host: "", err: ""},
		// {domain: "microsoft.it", host: "whois.nic.it", err: "", expired: true}, TODO: fix
		{domain: "google.pw", host: "", err: ""},
		{domain: "google.co.id", host: "", err: ""},
		{domain: "google.kr", host: "", err: ""},
		{domain: "google.jp", host: "", err: ""},
		{domain: "microsoft.im", host: "", err: ""},
		{domain: "google.rs", host: "", err: ""},
		{domain: "мвд.рф", host: "", err: ""},
		{domain: "МВД.РФ", host: "", err: ""},
		{domain: "GOOGLE.RS", host: "", err: ""},
		{domain: "google.co.th", host: "", err: ""},
		{domain: "google.fi", host: "", err: ""},
		{domain: "google.com.hk", host: "", err: ""},
		{domain: "hknic.hk", host: "", err: ""},
		{domain: "test.idv.hk", host: "", err: ""},
		{domain: "test.org.hk", host: "", err: ""},
		{domain: "hkirc.香港", host: "", err: ""},
		{domain: "google.vn", host: "whois.net.vn", err: ""},
		{domain: "google.com.tr", host: "", err: ""},
		{domain: "google.com.ru", host: "whois.nic.ru", err: ""},
		{domain: "nic.kz", host: "", err: ""},
		{domain: "google.io", host: "", err: ""},
		{domain: "google.ph", host: "whois.dot.ph", err: ""},
		{domain: "google.com", host: "whois.dot.ph", err: "Domain not found or parsing error"},
		{domain: "google.uz", host: "", err: ""},
		{domain: "google.cl", host: "", err: ""},
		{domain: "google.ru", host: "", err: ""},
	} {
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(cancel)

			expiry, err := NewClient().ExpireTime(ctx, tt.domain, tt.host)
			if err != nil {
				errs := err.Error()
				if strings.Contains(errs, "i/o timeout") {
					t.Skip("timeout")
				}
				if strings.Contains(errs, "Too may requests") {
					t.Skip("rate limit")
				}
			}
			if tt.err == "" {
				require.NoError(t, err)
				if tt.expired {
					require.Greater(t, time.Since(expiry).Hours(), 0.0)
				} else {
					require.Less(t, time.Since(expiry).Hours(), 0.0)
				}
			} else {
				require.ErrorContains(t, err, tt.err)
				t.Log(err)
			}
		})
	}
}
