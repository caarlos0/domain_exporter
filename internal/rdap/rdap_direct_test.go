package rdap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDirectRDAPEndpoint(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		domain   string
		expected string
		ok       bool
	}{
		{domain: "tanukifamily.kz", expected: "https://rdap.nic.kz/domain/", ok: true},
		{domain: "tanukifamily.uz", expected: "", ok: false},
		{domain: "tanukifamily.ru", expected: "", ok: false},
		{domain: "goreleaser.com", expected: "", ok: false},
	} {
		t.Run(tt.domain, func(t *testing.T) {
			t.Parallel()

			actual, ok := directRDAPEndpoint(tt.domain)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestParseExpirationDate(t *testing.T) {
	t.Parallel()

	expiration, err := parseExpirationDate("2027-04-04T16:45:03Z")
	require.NoError(t, err)
	require.Equal(t, time.Date(2027, 4, 4, 16, 45, 3, 0, time.UTC), expiration)
}
