package whois

import (
	"github.com/domainr/whois"
)

type luAdapter struct{}

func (a *luAdapter) Prepare(req *whois.Request) error {
	req.Host = "whois.dns.lu"
	return whois.DefaultAdapter.Prepare(req)
}

func (a *luAdapter) Text(res *whois.Response) ([]byte, error) {
	return whois.DefaultAdapter.Text(res)
}

func init() {
	whois.BindAdapter(&luAdapter{}, "whois.dns.lu")
}
