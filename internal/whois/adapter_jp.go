package whois

import "github.com/domainr/whois"

// Adapter for whois.jprs.jp.
type jpAdapter struct{}

func (a *jpAdapter) Prepare(req *whois.Request) error {
	// Add /e to suppress Japanese output.
	// https://jprs.jp/about/dom-search/jprs-whois/whois-guide-view.html
	req.Query += " /e"
	return whois.DefaultAdapter.Prepare(req)
}

func (a *jpAdapter) Text(res *whois.Response) ([]byte, error) {
	return whois.DefaultAdapter.Text(res)
}

// nolint: gochecknoinits
func init() {
	whois.BindAdapter(
		&jpAdapter{},
		"whois.jprs.jp",
	)
}
