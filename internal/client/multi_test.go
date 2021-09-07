package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type clifail int

func (clifail) ExpireTime(_ context.Context, domain string) (time.Time, error) {
	return time.Time{}, errors.New("foo")
}

type clisuccess time.Time

func (c clisuccess) ExpireTime(_ context.Context, domain string) (time.Time, error) {
	return time.Time(c), nil
}

func TestMulti(t *testing.T) {
	ctx := context.Background()
	t.Run("first client succeed", func(t *testing.T) {
		expected := time.Now()
		expire, err := NewMultiClient(clisuccess(expected), clifail(0)).ExpireTime(ctx, "a")
		require.NoError(t, err)
		require.Equal(t, expected, expire)
	})
	t.Run("last client succeed", func(t *testing.T) {
		expected := time.Now()
		expire, err := NewMultiClient(clifail(0), clifail(0), clisuccess(expected)).ExpireTime(ctx, "a")
		require.NoError(t, err)
		require.Equal(t, expected, expire)
	})
	t.Run("no client succeed", func(t *testing.T) {
		expire, err := NewMultiClient(clifail(0), clifail(0), clifail(0)).ExpireTime(ctx, "a")
		require.EqualError(t, err, "foo")
		require.Zero(t, expire)
	})
}
