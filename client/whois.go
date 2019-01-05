package client

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/domainr/whois"
)

var formats = []string{
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	"20060102",               // .com.br
	"2006-01-02",             // .lt
	"2006-01-02 15:04:05-07", // .ua
	"2006-01-02 15:04:05",    // .ch
	"2006-01-02T15:04:05Z",   // .name
	"January  2 2006",        // .is
	"02.01.2006",             // .cz
	"02/01/2006",             // .fr
	"02-January-2006",        // .ie
	"2006.01.02 15:04:05",    // .pl
	"02-Jan-2006",            // .co.uk
	"2006/01/02",             // .ca
}

// nolint: lll
var re = regexp.MustCompile(`(?i)(Registry Expiry Date|paid-till|Expiration Date|Expiration Time|Expiry.*|expires.*|Expires|Expires On|expire|Renewal Date|Expire Date):\s+(.*)`)

type whoisClient struct {
}

// NewWhoisClient return a "live" whois client
func NewWhoisClient() Client {
	return whoisClient{}
}

func (whoisClient) ExpireTime(domain string) (time.Time, error) {
	req, err := whois.NewRequest(domain)
	if err != nil {
		return time.Now(), err
	}
	resp, err := whois.DefaultClient.Fetch(req)
	if err != nil {
		return time.Now(), err
	}
	var body = string(resp.Body)
	var result = re.FindStringSubmatch(body)
	if len(result) < 2 {
		return time.Now(), fmt.Errorf("could not parse whois response: %s", body)
	}
	var dateStr = strings.TrimSpace(result[2])
	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}
	return time.Now(), fmt.Errorf("could not parse date: %s", dateStr)
}
