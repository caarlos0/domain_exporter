package whois

import (
	"regexp"

	"github.com/domainr/whois"
)

// ltAdapter handles .lt domains served by whois.domreg.lt.
//
// The DOMREG WHOIS response contains:
// - a "Registered:" line with the registration date
// - a "Status: registered" line
//
// The expiryRE regexp matches "registered" in the Status line before reaching
// "Expires:", then captures the rest including "Expires:\t\t2026-12-31".
// This adapter removes both lines so "Expires:" is matched correctly.
type ltAdapter struct{}

func (a *ltAdapter) Prepare(req *whois.Request) error {
	return whois.DefaultAdapter.Prepare(req)
}

var ltStripRE = regexp.MustCompile(`(?im)^(Registered:|Status:).*\n?`)

func (a *ltAdapter) Text(res *whois.Response) ([]byte, error) {
	text, err := whois.DefaultAdapter.Text(res)
	if err != nil {
		return nil, err
	}
	return ltStripRE.ReplaceAll(text, nil), nil
}

func init() {
	adapter := &ltAdapter{}
	whois.BindAdapter(adapter, "whois.domreg.lt")
}

