package whois

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEnrichKZWhoisResponseAddsSyntheticPaidTillForActiveDomains(t *testing.T) {
	t.Parallel()

	body := []byte(`Domain Name............: nic.kz
Domain status : ok`)

	now := time.Date(2026, 4, 14, 8, 9, 23, 0, time.UTC)
	result := enrichKZWhoisResponse(body, func() time.Time { return now })

	require.Contains(t, string(result), "paid-till: 2027-04-14T08:09:23Z")
}

func TestEnrichKZWhoisResponseKeepsOriginalBodyWhenDomainIsNotActive(t *testing.T) {
	t.Parallel()

	body := []byte(`Domain Name............: missing.kz
Domain status : pendingDelete`)

	result := enrichKZWhoisResponse(body, time.Now)

	require.Equal(t, string(body), string(result))
}
