package whois

import (
	"regexp"

	"github.com/domainr/whois"
)

// ruAdapter handles .ru, .su, and .рф (xn--p1ai) domains served by whois.tcinet.ru.
//
// The tcinet WHOIS response contains a "state:" line with domain status:
//
//	state:         REGISTERED, DELEGATED, VERIFIED
//
// The expiryRE regexp matches "registered" in this line before reaching "paid-till",
// then fails to parse "DELEGATED, VERIFIED" as a date.
// This adapter removes the "state:" line so "paid-till" is matched correctly.
type ruAdapter struct{}

func (a *ruAdapter) Prepare(req *whois.Request) error {
	return whois.DefaultAdapter.Prepare(req)
}

var tcinetStateRE = regexp.MustCompile(`(?im)^state:.*\n?`)

func (a *ruAdapter) Text(res *whois.Response) ([]byte, error) {
	text, err := whois.DefaultAdapter.Text(res)
	if err != nil {
		return nil, err
	}
	return tcinetStateRE.ReplaceAll(text, nil), nil
}

func init() {
	adapter := &ruAdapter{}
	whois.BindAdapter(adapter, "whois.tcinet.ru")
}
