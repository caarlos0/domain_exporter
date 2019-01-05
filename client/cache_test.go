package client

import (
	"testing"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
)

type cacheTestClient struct {
	result *time.Time
}

func (f cacheTestClient) ExpireTime(domain string) (time.Time, error) {
	return *f.result, nil
}

func TestCachedClient(t *testing.T) {
	var cache = cache.New(1*time.Minute, 1*time.Minute)
	var expected = time.Now()
	var domain = "foo.bar"

	var cli = NewCachedClient(cacheTestClient{result: &expected}, cache)

	// test getting from out fake client
	t.Run("get fresh", func(t *testing.T) {
		res, err := cli.ExpireTime(domain)
		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	// here we change the inner fake client result, but the result
	// should be the cached one
	t.Run("get from cache", func(t *testing.T) {
		var oldExpected = expected
		expected = time.Now()
		res, err := cli.ExpireTime(domain)
		require.NoError(t, err)
		require.Equal(t, oldExpected, res)
	})

	// here we flush the cache and verify that the result is the one
	// from the fake client
	t.Run("flush cache", func(t *testing.T) {
		cache.Flush()
		res, err := cli.ExpireTime(domain)
		require.NoError(t, err)
		require.Equal(t, expected, res)
	})
}
