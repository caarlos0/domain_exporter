package whois

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFallbackToRDAPSucceedsWhenWhoisFails(t *testing.T) {
	t.Parallel()

	previous := rdapFallbackExpireTime
	t.Cleanup(func() {
		rdapFallbackExpireTime = previous
	})

	expected := time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC)
	rdapFallbackExpireTime = func(_ context.Context, domain string) (time.Time, error) {
		require.Equal(t, "tanukifamily.ru", domain)
		return expected, nil
	}

	expiration, err := fallbackToRDAP(context.Background(), "tanukifamily.ru", "", errors.New("whois failed"))
	require.NoError(t, err)
	require.Equal(t, expected, expiration)
}

func TestFallbackToRDAPPreservesErrorForExplicitHost(t *testing.T) {
	t.Parallel()

	previous := rdapFallbackExpireTime
	t.Cleanup(func() {
		rdapFallbackExpireTime = previous
	})

	rdapCalled := false
	rdapFallbackExpireTime = func(_ context.Context, _ string) (time.Time, error) {
		rdapCalled = true
		return time.Time{}, nil
	}

	before := time.Now()
	expiration, err := fallbackToRDAP(context.Background(), "google.com", "whois.dot.ph", errors.New("whois failed"))
	require.ErrorContains(t, err, "whois failed")
	require.WithinDuration(t, before, expiration, time.Second)
	require.False(t, rdapCalled)
}

func TestFallbackToRDAPCombinesErrorsWhenRdapFails(t *testing.T) {
	t.Parallel()

	previous := rdapFallbackExpireTime
	t.Cleanup(func() {
		rdapFallbackExpireTime = previous
	})

	rdapFallbackExpireTime = func(_ context.Context, _ string) (time.Time, error) {
		return time.Time{}, errors.New("rdap failed")
	}

	_, err := fallbackToRDAP(context.Background(), "missing.ru", "", errors.New("whois failed"))
	require.ErrorContains(t, err, "whois failed")
	require.ErrorContains(t, err, "rdap fallback failed: rdap failed")
}
