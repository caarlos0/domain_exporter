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
		err    string
	}{
		{domain: "google.ai", err: "could not parse whois response"},
		{domain: "google.lt", err: ""},
		{domain: "fakedomain.foo", err: "Domain not found"},
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
		{domain: "google.co.th", err: ""},
	} {
		tt := tt
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			t.Cleanup(cancel)

			expiry, err := NewClient().ExpireTime(ctx, tt.domain)
			if tt.err == "" {
				is.NoErr(err)                           // expected no errors
				is.True(time.Since(expiry).Hours() < 0) // domain must not be expired
			} else {
				is.True(err != nil) // expected an error
				t.Log(err)
				is.True(strings.Contains(err.Error(), tt.err)) // expected error to contain message
			}
		})
	}
}
