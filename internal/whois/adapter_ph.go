package whois

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/domainr/whois"
)

// Adapter for whois.dot.ph
type phAdapter struct{}

func (a *phAdapter) Prepare(req *whois.Request) error {
	req.URL = generatePhWhoisRequestUrl(req.Query)
	// Override request body to avoid any conflict.
	req.Body = nil

	return nil
}

func (a *phAdapter) Text(res *whois.Response) ([]byte, error) {
	return parsePhWhoisResponseBody(res.Body), nil
}

func init() {
	whois.BindAdapter(
		&phAdapter{},
		"whois.dot.ph",
	)
}

// Generate URL for .ph whois request.
// Query sent to whois.dot.ph should go to `/` route with `search` parameter.
func generatePhWhoisRequestUrl(query string) string {
	whoisEndpoint := "https://whois.dot.ph/?"

	whoisQueryParams := url.Values{}
	whoisQueryParams.Set("search", query)

	return whoisEndpoint + whoisQueryParams.Encode()
}

// Remove HTML Tags, grab dates and generate a registar field
// to avoid parsing errors in later step.
func parsePhWhoisResponseBody(bodyContent []byte) []byte {
	htmlTags := []string{"<br/>", "<div>", "</div>"}

	// Convert the body to string for easier handling
	bodyString := string(bodyContent)

	// Extract dates from inline js
	createDate := extractDateFromJS(bodyString, "createDate")
	expiryDate := extractDateFromJS(bodyString, "expiryDate")
	updateDate := extractDateFromJS(bodyString, "updateDate")

	// Find the content within the result-whois div using regex
	resultWhoisRegex := regexp.MustCompile(`<div id="result-whois"[^>]*>([\s\S]*?)</div>`)
	resultMatch := resultWhoisRegex.FindStringSubmatch(bodyString)

	if len(resultMatch) < 2 {
		return []byte("Registrar WHOIS Server: whois.dot.ph\nDomain not found or parsing error")
	}

	// Get the content of the whois response
	resBodyString := resultMatch[1]

	// Remove HTML tag (<br>, <div>).
	for _, htmlTag := range htmlTags {
		resBodyString = strings.ReplaceAll(resBodyString, htmlTag, "")
	}

	// Remove tab characters.
	resBodyString = strings.ReplaceAll(resBodyString, "\t", "")

	// Replace the placeholders with the extracted date values
	resBodyString = strings.Replace(resBodyString, `<span id="create-date"></span>`, createDate, 1)
	resBodyString = strings.Replace(resBodyString, `<span id="expiry-date"></span>`, expiryDate, 1)
	resBodyString = strings.Replace(resBodyString, `<span id="update-date"></span>`, updateDate, 1)

	// Generate Registrar field.
	resBodyString = "Registrar WHOIS Server: Not available" + resBodyString

	return []byte(resBodyString)
}

// Extract date values from inline js in the HTML
func extractDateFromJS(htmlContent, dateVarName string) string {
	// Pattern to match moment('DATE_VALUE') in the JavaScript
	jsDateRegex := regexp.MustCompile(`var\s+` + dateVarName + `\s*=\s*moment\(['"]([^'"]+)['"]\)`)
	match := jsDateRegex.FindStringSubmatch(htmlContent)

	if len(match) > 1 {
		return match[1] // Return the captured date value
	}

	return ""
}
