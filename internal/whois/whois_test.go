package whois

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestWhoisParsing(t *testing.T) {
	for _, tt := range []struct {
		domain string
		host   string
		err    string
	}{
		{domain: "google.ai", host: "", err: "could not parse whois response"},
		{domain: "google.lt", host: "", err: ""},
		{domain: "fakedomain.foo", host: "", err: "Domain not found"},
		{domain: "google.cn", host: "", err: ""},
		{domain: "google.com", host: "", err: ""},
		{domain: "google.de", host: "", err: "could not parse whois response"},
		{domain: "nic.ua", host: "", err: ""},
		{domain: "google.com.tw", host: "", err: ""},
		{domain: "bbc.co.uk", host: "", err: ""},
		{domain: "google.sg", host: "", err: ""},
		{domain: "google.sk", host: "", err: ""},
		{domain: "google.ro", host: "", err: ""},
		{domain: "google.pt", host: "", err: "i/o timeout"},
		{domain: "microsoft.it", host: "", err: ""},
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
		{domain: "google.vn", host: "whois.net.vn", err: ""},
	} {
		tt := tt
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
				is.NoErr(err)                           // expected no errors
				is.True(time.Since(expiry).Hours() < 0) // domain must not be expired
			} else {
				is.True(err != nil)                            // expected an error
				is.True(strings.Contains(err.Error(), tt.err)) // expected error to contain message
				t.Log(err)
			}
		})
	}
}
