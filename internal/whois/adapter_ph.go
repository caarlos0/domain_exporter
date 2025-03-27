package whois

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/domainr/whois"
)

type phAdapter struct{}

func (a *phAdapter) Prepare(req *whois.Request) error {
	req.URL = generatePhWhoisRequestUrl(req.Query)
	req.Body = nil
	return nil
}

func (a *phAdapter) Text(res *whois.Response) ([]byte, error) {
	return parsePhWhoisResponseBody(res.Body)
}

func init() {
	whois.BindAdapter(
		&phAdapter{},
		"whois.dot.ph",
	)
}

// Generate URL for .ph WHOIS request (GET with "search" parameter)
func generatePhWhoisRequestUrl(query string) string {
	return "https://whois.dot.ph/whois.php?search=" + url.QueryEscape(query)
}

// Basic HTML cleanup similar to .vn adapter
func parsePhWhoisResponseBody(bodyContent []byte) ([]byte, error) {
	resBodyString := string(bodyContent)

	start := strings.Index(resBodyString, "<pre>")
	end := strings.Index(resBodyString, "</pre>")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("failed to find <pre> tags in response body")
	}
	resBodyString = resBodyString[start+5 : end]

	htmlTags := []string{"<br>", "</br>", "<b>", "</b>"}
	for _, tag := range htmlTags {
		resBodyString = strings.ReplaceAll(resBodyString, tag, "")
	}

	resBodyString = strings.ReplaceAll(resBodyString, "\t", "")
	resBodyString = strings.TrimSpace(resBodyString)

	return []byte(resBodyString), nil
}
