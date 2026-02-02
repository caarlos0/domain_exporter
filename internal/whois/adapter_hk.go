package whois

import (
	"github.com/domainr/whois"
)

// hkAdapter for .hk and .香港 (xn--j6c99c) domains.
type hkAdapter struct{}

func (a *hkAdapter) Prepare(req *whois.Request) error {
	req.Host = "whois.hkirc.hk"
	return whois.DefaultAdapter.Prepare(req)
}

func (a *hkAdapter) Text(res *whois.Response) ([]byte, error) {
	return whois.DefaultAdapter.Text(res)
}

func init() {
	whois.BindAdapter(&hkAdapter{}, "whois.hkirc.hk")
}
