package whois

import (
	"net/url"
	"strings"

	"github.com/domainr/whois"
)

// Adapter for whois.net.vn
type vnAdapter struct{}

func (a *vnAdapter) Prepare(req *whois.Request) error {
	req.URL = generateVnWhoisRequestUrl(req.Query)
	// Override request body to avoid any conflict.
	req.Body = nil

	return nil
}

func (a *vnAdapter) Text(res *whois.Response) ([]byte, error) {
	return parseVnWhoisResponseBody(res.Body), nil
}

func init() {
	whois.BindAdapter(
		&vnAdapter{},
		"whois.net.vn",
	)
}

// Generate URL for .vn whois request.
// Query sent to whois.net.vn should go to `/whois.php` route with `act` & `domain` parameters.
func generateVnWhoisRequestUrl(query string) string {
	whoisEndpoint := "https://whois.net.vn/whois.php?"

	whoisQueryParams := url.Values{}
	whoisQueryParams.Set("act", "getwhois")
	whoisQueryParams.Set("domain", query)

	return whoisEndpoint + whoisQueryParams.Encode()
}

// Remove HTML Tags, tab characters, exceeding spaces, correct some keys and generate a registar field
// to avoid parsing errors in later step.
func parseVnWhoisResponseBody(bodyContent []byte) []byte {
	htmlTags := []string{"<br/>", "<div>", "</div>"}

	resBodyString := string(bodyContent)

	// Remove HTML tag (<br>, <div>).
	for _, htmlTag := range htmlTags {
		resBodyString = strings.ReplaceAll(resBodyString, htmlTag, "")
	}

	// Remove tab characters.
	resBodyString = strings.ReplaceAll(resBodyString, "\t", "")

	// Remove exceeding spaces & correct some keys in response.
	resBodyString = strings.ReplaceAll(resBodyString, "Expired Date : ", "Expire Date:")
	resBodyString = strings.ReplaceAll(resBodyString, "Issue Date : ", "Issue Date:")

	// Generate Registrar field.
	resBodyString = "Registrar WHOIS Server: Not available" + resBodyString

	return []byte(resBodyString)
}
