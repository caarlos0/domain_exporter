package whois

import (
	"fmt"
	"regexp"
	"time"
	
	"github.com/domainr/whois"
	"github.com/rs/zerolog/log"
)

// kzAdapter implements custom adapter for .kz domains
type kzAdapter struct{}

func (a *kzAdapter) Prepare(req *whois.Request) error {
	return whois.DefaultAdapter.Prepare(req)
}

func (a *kzAdapter) Text(res *whois.Response) ([]byte, error) {
	text, err := whois.DefaultAdapter.Text(res)
	if err != nil {
		return nil, err
	}

	// Look for domain status - if it contains "ok", the domain is active
	statusRegex := regexp.MustCompile(`Domain status\s*:\s*ok`)
	if statusRegex.Match(text) {
		log.Debug().Msg("KZ domain is active based on status")
		
		// For active domains, use current date + 1 year as expiry
		// KZ WHOIS doesn't provide explicit expiry dates
		expiration := time.Now().AddDate(1, 0, 0)
		
		response := string(text)
		response += fmt.Sprintf("\npaid-till: %s", expiration.Format("2006-01-02T15:04:05Z"))
		return []byte(response), nil
	}
	
	// If domain is not active or status is not found, return original text
	// This will likely cause the domain to be reported as expired, which is correct
	log.Debug().Msg("KZ domain is not active or status not found")
	return text, nil
}

func init() {
	whois.BindAdapter(
		&kzAdapter{},
		"whois.nic.kz",
	)
}