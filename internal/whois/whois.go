package whois

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/caarlos0/domain_exporter/internal/client"
	"github.com/domainr/whois"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/idna"
)

// nolint: gochecknoglobals
var (
	formats = []string{
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
		"2006-01-02T15:04:05-0700",    // some .com
		"20060102",                    // .com.br
		"2006-01-02",                  // .lt
		"2006-01-02 15:04:05-07",      // .ua
		"2006-01-02 15:04:05",         // .ch
		"2006-01-02T15:04:05Z",        // .name
		"2006-01-02T15:04:05.0Z",      // .host
		"January  2 2006",             // .is
		"02.01.2006",                  // .cz
		"02/01/2006",                  // .fr
		"02-January-2006",             // .ie
		"2006.01.02 15:04:05",         // .pl
		"02-Jan-2006",                 // .co.uk
		"02-Jan-2006 15:04:05",        // .sg
		"2006-01-02T15:04:05Z",        // .co
		"2006/01/02",                  // .ca, .jp
		"2006-01-02 (YYYY-MM-DD)",     // .tw
		"(dd/mm/yyyy): 02/01/2006",    // .pt
		"02-Jan-2006 15:04:05 UTC",    // .id, .co.id
		": 2006. 01. 02.",             // .kr
		"03/05/2006 15:04:05",         // .im
		"2006-01-02 15:04:05 (UTC+8)", // .tw
		"02/01/2006 15:04:05",         // .im
		"02.01.2006 15:04:05",         // .rs
		"02 Jan 2006",                 // .co.th
		"2.1.2006 15:04:05",           // .fi
	}

	// nolint: lll
	expiryRE = regexp.MustCompile(`(?i)(` + strings.Join([]string{
		"Registrar Registration Expiration Date",
		"expire-date",
		"Valid Until",
		"Expire Date",
		"Registry Expiry Date",
		"paid-till",
		"Expiration Date",
		"Expiration Time",
		"Expiry date",
		"Expiry",
		"Expires On",
		"expires\\.{12}",
		"expires",
		"Expires",
		"expire",
		"Renewal Date",
		"Record expires on",
		"Exp date",
	}, "|") + `)\]?:?\s?(.*)`)
	registrarRE = regexp.MustCompile(`(?i)Registrar WHOIS Server: (.*)`)
)

type whoisClient struct{}

// NewClient return a "live" whois client.
func NewClient() client.Client {
	return whoisClient{}
}

func (c whoisClient) ExpireTime(ctx context.Context, domain string, host string) (time.Time, error) {
	log.Debug().Msgf("trying whois client for %q", domain)
	body, err := c.request(ctx, domain, host)
	if err != nil {
		return time.Now(), err
	}
	result := expiryRE.FindStringSubmatch(body)
	if len(result) < 2 {
		return time.Now(), fmt.Errorf("could not parse whois response: %q", body)
	}
	dateStr := strings.TrimSpace(result[2])
	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			log.Debug().Msgf("domain %q will expire at %q", domain, date.String())
			return date, nil
		}
	}
	return time.Now(), fmt.Errorf("could not parse date: %q", dateStr)
}

func (c whoisClient) request(ctx context.Context, domain, host string) (string, error) {
	normalizedDomain := strings.ToLower(domain)

	normalizedDomain, err := idna.ToASCII(normalizedDomain)
	if err != nil {
		return "", fmt.Errorf("failed to normalize domain name: %w", err)
	}

	req := &whois.Request{
		Query: normalizedDomain,
		Host:  host,
	}
	if err := req.Prepare(); err != nil {
		return "", fmt.Errorf("failed to prepare: %w", err)
	}
	resp, err := whois.DefaultClient.FetchContext(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch whois request: %w", err)
	}
	respText, err := resp.Text()
	if err != nil {
		return "", fmt.Errorf("failed to parse response body into text: %w", err)
	}

	body := string(respText)

	if host == "" {
		// do not recurse
		return body, nil
	}

	result := registrarRE.FindStringSubmatch(body)
	if len(result) < 2 {
		log.Debug().Msgf("couldn't find registrar url in whois response: %s", domain)
		return body, nil
	}

	foundHost := strings.TrimSpace(result[1])
	if foundHost == host || foundHost == "" {
		return body, nil
	}

	log.Debug().Msgf("found whois host %s for domain %s", foundHost, domain)
	if newBody, err := c.request(ctx, domain, foundHost); err == nil {
		return newBody, err
	}

	log.Debug().Msgf("ignoring error from %s for %s", foundHost, domain)
	return body, nil
}
