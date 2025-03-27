package whois

import (
	"fmt"
	"html"
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

// Parses the WHOIS response from whois.dot.ph
func parsePhWhoisResponseBody(bodyContent []byte) ([]byte, error) {
	resBodyString := string(bodyContent)

	// Debug: Print raw response for troubleshooting
	fmt.Println("Raw WHOIS Response:", resBodyString)

	// Extract WHOIS details inside <pre> tags
	start := strings.Index(resBodyString, "<pre>")
	end := strings.Index(resBodyString, "</pre>")

	if start == -1 || end == -1 {
		return nil, fmt.Errorf("failed to find <pre> tags in response body: %s", resBodyString)
	}

	resBodyString = resBodyString[start+5 : end]

	// Cleanup unwanted HTML tags and special characters
	htmlTags := []string{"<br>", "</br>", "<b>", "</b>", "&nbsp;", "&lt;", "&gt;", "&amp;"}
	for _, tag := range htmlTags {
		resBodyString = strings.ReplaceAll(resBodyString, tag, "")
	}

	// Decode HTML entities (e.g., &quot;, &#8212;)
	resBodyString = html.UnescapeString(resBodyString)

	// Remove excessive whitespace
	resBodyString = strings.TrimSpace(resBodyString)

	return []byte(resBodyString), nil
}
