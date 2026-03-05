package whois

import (
	"regexp"

	"github.com/domainr/whois"
)

// eeAdapter handles .ee domains served by whois.tld.ee.
//
// The TLD.EE WHOIS response contains a "registered:" line with the registration
// date before the "expire:" line:
//
//	registered: 2016-02-05 12:00:54 +02:00
//	expire:     2027-02-06
//
// The expiryRE regexp matches "registered" before reaching "expire:",
// then parses the registration date instead of the expiry date.
// This adapter removes the "registered:" line so "expire:" is matched correctly.
type eeAdapter struct{}

func (a *eeAdapter) Prepare(req *whois.Request) error {
	return whois.DefaultAdapter.Prepare(req)
}

var eeRegisteredRE = regexp.MustCompile(`(?im)^registered:.*\n?`)

func (a *eeAdapter) Text(res *whois.Response) ([]byte, error) {
	text, err := whois.DefaultAdapter.Text(res)
	if err != nil {
		return nil, err
	}
	return eeRegisteredRE.ReplaceAll(text, nil), nil
}

func init() {
	adapter := &eeAdapter{}
	whois.BindAdapter(adapter, "whois.tld.ee")
}

