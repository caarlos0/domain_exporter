package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
)

type testClient struct {
	result *time.Time
}

func (f testClient) ExpireTime(_ context.Context, _ string) (time.Time, error) {
	return *f.result, nil
}

type errTestClient struct {
}

func (f errTestClient) ExpireTime(_ context.Context, _ string) (time.Time, error) {
	return time.Now(), fmt.Errorf("failed to get domain info blah")
}

func TestCachedClient(t *testing.T) {
	ctx := context.Background()
	cache := cache.New(1*time.Minute, 1*time.Minute)
	expected := time.Now()
	domain := "foo.bar"

	cli := NewCachedClient(testClient{result: &expected}, cache)

	// test getting from out fake client
	t.Run("get fresh", func(t *testing.T) {
		res, err := cli.ExpireTime(ctx, domain)
		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	// here we change the inner fake client result, but the result
	// should be the cached one
	t.Run("get from cache", func(t *testing.T) {
		oldExpected := expected
		expected = time.Now()
		res, err := cli.ExpireTime(ctx, domain)
		require.NoError(t, err)
		require.Equal(t, oldExpected, res)
	})

	// here we flush the cache and verify that the result is the one
	// from the fake client
	t.Run("flush cache", func(t *testing.T) {
		cache.Flush()
		res, err := cli.ExpireTime(ctx, domain)
		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("do not cache errors", func(t *testing.T) {
		cache.Flush()

		cli := NewCachedClient(errTestClient{}, cache)
		_, err := cli.ExpireTime(ctx, domain)
		require.Error(t, err)
		_, err = cli.ExpireTime(ctx, domain)
		require.Error(t, err)
		cached, got := cache.Get(domain)
		require.Nil(t, cached)
		require.False(t, got)
	})
}
