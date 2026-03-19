package whois

import (
	"regexp"

	"github.com/domainr/whois"
)

// ruAdapter cleans tcinet WHOIS responses so expiry parsing reaches paid-till.
// tcinet responses contain a line like:
// state: REGISTERED, DELEGATED, VERIFIED
// which can be mistaken for an expiry value because of the "registered" keyword.
type ruAdapter struct{}

func (a *ruAdapter) Prepare(req *whois.Request) error {
	return whois.DefaultAdapter.Prepare(req)
}

var tcinetStateRE = regexp.MustCompile(`(?im)^state:\s*.*(?:\n[ \t].*)*`)

func (a *ruAdapter) Text(res *whois.Response) ([]byte, error) {
	text, err := whois.DefaultAdapter.Text(res)
	if err != nil {
		return nil, err
	}
	cleaned := tcinetStateRE.ReplaceAll(text, nil)
	return cleaned, nil
}

func init() {
	whois.BindAdapter(&ruAdapter{}, "whois.tcinet.ru")
}
