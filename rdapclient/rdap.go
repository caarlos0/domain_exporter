package rdapclient

import (
	"fmt"
	"time"

	"github.com/openrdap/rdap"
)

// nolint: gochecknoglobals
var (
	formats = []string{
		time.RFC3339,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339Nano,
		"20060102",                 // .com.br
		"2006-01-02",               // .lt
		"2006-01-02 15:04:05-07",   // .ua
		"2006-01-02 15:04:05",      // .ch
		"2006-01-02T15:04:05Z",     // .name
		"January  2 2006",          // .is
		"02.01.2006",               // .cz
		"02/01/2006",               // .fr
		"02-January-2006",          // .ie
		"2006.01.02 15:04:05",      // .pl
		"02-Jan-2006",              // .co.uk
		"2006/01/02",               // .ca
		"2006-01-02 (YYYY-MM-DD)",  // .tw
		"(dd/mm/yyyy): 02/01/2006", //.pt
		"02-Jan-2006 15:04:05 UTC", // .id, .co.id
		": 2006. 01. 02.",          // .kr
	}
)

type rdapClient struct {
}

func NewRdapClient() RdapClient {
	return rdapClient{}
}
func (rdapClient) ExpireTime(domain string) (time.Time, error) {
	client := &rdap.Client{}
	body, err := client.QueryDomain(domain)
	if err != nil {
		return time.Now(), err
	}
	for _, event := range body.Events {
		if event.Action == "expiration" {
			for _, format := range formats {
				if date, err := time.Parse(format, event.Date); err == nil {
					return date, nil
				}
			}
			return time.Now(), fmt.Errorf("could not parse date: %s", event.Date)
		}
	}
	return time.Now(), fmt.Errorf("no expiration event for domain: %s ", domain)
}
