package rdap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
                "2006-01-02 15:04:05 (GMT-7:00)", // .kz
                "2006-01-02 15:04:05 Z07:00", // normalized .kz
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
)

type rdapClient struct{}

type kzRDAPResponse struct {
	Events []struct {
		Action string `json:"eventAction"`
		Date   string `json:"eventDate"`
	} `json:"events"`
}

// NewClient returns a new RDAP client.
func NewClient() client.Client {
	return rdapClient{}
}

func parseDate(value string) (time.Time, error) {
	normalized := strings.TrimSpace(value)

	// KZ RDAP returns e.g. "2026-08-01 11:49:04 (GMT+0:00)"
	// Normalize to a form Go can parse reliably.
	normalized = strings.ReplaceAll(normalized, "(GMT+0:00)", "Z")
	normalized = strings.ReplaceAll(normalized, "(GMT-0:00)", "Z")

	for _, format := range formats {
		if date, err := time.Parse(format, normalized); err == nil {
			return date, nil
		}
	}

	// Extra direct fallback for KZ after normalization.
	if date, err := time.Parse("2006-01-02 15:04:05 Z07:00", strings.ReplaceAll(strings.ReplaceAll(value, "(GMT", ""), ")", "")); err == nil {
		return date, nil
	}

	return time.Now(), fmt.Errorf("could not parse date: %s", value)
}

func (rdapClient) ExpireTime(ctx context.Context, domain string, host string) (time.Time, error) {
	log.Debug().Msgf("trying rdap client for %s", domain)
	if strings.HasSuffix(strings.ToLower(domain), ".kz") {
		log.Debug().Msgf("trying direct KZ RDAP for %s", domain)

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"https://rdap.nic.kz/domain/"+domain,
			nil,
		)
		if err != nil {
			return time.Now(), fmt.Errorf("failed to create KZ RDAP request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return time.Now(), fmt.Errorf("failed to do KZ RDAP request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return time.Now(), fmt.Errorf("KZ RDAP returned status %d", resp.StatusCode)
		}

		var body kzRDAPResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return time.Now(), fmt.Errorf("failed to decode KZ RDAP response: %w", err)
		}

		for _, event := range body.Events {
			if event.Action == "expiration" {
				return parseDate(event.Date)
			}
		}

		return time.Now(), fmt.Errorf("no expiration event for domain: %s", domain)
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

for _, event := range body.Events {
	if event.Action == "expiration" {
		return parseDate(event.Date)
	}
}
	return time.Now(), fmt.Errorf("no expiration event for domain: %s ", domain)
}
