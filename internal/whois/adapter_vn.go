package whois

import "github.com/domainr/whois"

// Adapter for whois.net.vn
type vnAdapter struct{}

func (a *vnAdapter) Prepare(req *whois.Request) error {
	return whois.DefaultAdapter.Prepare(req)
}

func (a *vnAdapter) Text(res *whois.Response) ([]byte, error) {
	return whois.DefaultAdapter.Text(res)
}

func init() {
	whois.BindAdapter(
		&vnAdapter{},
		"whois.net.vn",
	)
}
