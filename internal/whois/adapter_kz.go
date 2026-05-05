package whois

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/domainr/whois"
	"github.com/rs/zerolog/log"
)

var kzStatusRegex = regexp.MustCompile(`(?im)^Domain status\s*:\s*ok\b`)

// kzAdapter implements custom adapter for .kz domains.
type kzAdapter struct{}

func (a *kzAdapter) Prepare(req *whois.Request) error {
	return whois.DefaultAdapter.Prepare(req)
}

func (a *kzAdapter) Text(res *whois.Response) ([]byte, error) {
	text, err := whois.DefaultAdapter.Text(res)
	if err != nil {
		return nil, err
	}

	return enrichKZWhoisResponse(text, time.Now), nil
}

func enrichKZWhoisResponse(text []byte, now func() time.Time) []byte {
	response := string(text)
	if strings.Contains(strings.ToLower(response), "paid-till:") {
		return text
	}

	if !kzStatusRegex.MatchString(response) {
		log.Debug().Msg("KZ domain is not active or status not found")
		return text
	}

	log.Debug().Msg("KZ domain is active based on status")
	expiration := now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)
	return []byte(response + fmt.Sprintf("\npaid-till: %s", expiration))
}

func init() {
	whois.BindAdapter(
		&kzAdapter{},
		"whois.nic.kz",
	)
}
