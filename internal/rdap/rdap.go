package rdap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/caarlos0/domain_exporter/internal/client"
	"github.com/openrdap/rdap"
	"github.com/rs/zerolog/log"
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
		"2006-01-02T15:04:05.0Z",   // .host
		"January  2 2006",          // .is
		"02.01.2006",               // .cz
		"02/01/2006",               // .fr
		"02-January-2006",          // .ie
		"2006.01.02 15:04:05",      // .pl
		"02-Jan-2006",              // .co.uk
		"02-Jan-2006 15:04:05",     // .sg
		"2006-01-02T15:04:05Z",     // .co
		"2006/01/02",               // .ca
		"2006-01-02 (YYYY-MM-DD)",  // .tw
		"(dd/mm/yyyy): 02/01/2006", // .pt
		"02-Jan-2006 15:04:05 UTC", // .id, .co.id
		": 2006. 01. 02.",          // .kr
	}
	directEndpoints = map[string]string{
		"kz": "https://rdap.nic.kz/domain/",
	}
)

type rdapClient struct{}

type directRDAPResponse struct {
	Events []directRDAPEvent `json:"events"`
}

type directRDAPEvent struct {
	Action string `json:"eventAction"`
	Date   string `json:"eventDate"`
}

// NewClient returns a new RDAP client.
func NewClient() client.Client {
	return rdapClient{}
}

func (rdapClient) ExpireTime(ctx context.Context, domain string, host string) (time.Time, error) {
	log.Debug().Msgf("trying rdap client for %s", domain)
	if hasDirectEndpoint(domain) {
		return lookupDirectExpireTime(ctx, domain)
	}

	req := &rdap.Request{
		Type:  rdap.DomainRequest,
		Query: domain,
	}
	req = req.WithContext(ctx)

	client := &rdap.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return time.Now(), fmt.Errorf("failed to do rdap request: %w", err)
	}

	body, ok := resp.Object.(*rdap.Domain)
	if !ok {
		return time.Now(), fmt.Errorf("failed to cast rdap domain object: %w", err)
	}

	return extractExpirationFromRdapEvents(body.Events, domain)
}

func hasDirectEndpoint(domain string) bool {
	_, ok := directRDAPEndpoint(domain)
	return ok
}

func directRDAPEndpoint(domain string) (string, bool) {
	idx := strings.LastIndex(strings.ToLower(domain), ".")
	if idx == -1 || idx == len(domain)-1 {
		return "", false
	}

	endpoint, ok := directEndpoints[strings.ToLower(domain[idx+1:])]
	return endpoint, ok
}

func lookupDirectExpireTime(ctx context.Context, domain string) (time.Time, error) {
	endpoint, ok := directRDAPEndpoint(domain)
	if !ok {
		return time.Time{}, fmt.Errorf("no direct rdap endpoint for domain: %s", domain)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+url.PathEscape(domain), nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to create direct rdap request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to do direct rdap request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("unexpected direct rdap status: %s", resp.Status)
	}

	var body directRDAPResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return time.Time{}, fmt.Errorf("failed to decode direct rdap response: %w", err)
	}

	return extractExpirationFromDirectEvents(body.Events, domain)
}

func extractExpirationFromDirectEvents(events []directRDAPEvent, domain string) (time.Time, error) {
	for _, event := range events {
		if event.Action == "expiration" {
			return parseExpirationDate(event.Date)
		}
	}

	return time.Now(), fmt.Errorf("no expiration event for domain: %s", domain)
}

func extractExpirationFromRdapEvents(events []rdap.Event, domain string) (time.Time, error) {
	for _, event := range events {
		if event.Action == "expiration" {
			return parseExpirationDate(event.Date)
		}
	}

	return time.Now(), fmt.Errorf("no expiration event for domain: %s", domain)
}

func parseExpirationDate(value string) (time.Time, error) {
	for _, format := range formats {
		if date, err := time.Parse(format, value); err == nil {
			return date, nil
		}
	}

	return time.Now(), fmt.Errorf("could not parse date: %s", value)
}
