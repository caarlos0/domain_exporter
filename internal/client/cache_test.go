package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
	cache "github.com/patrickmn/go-cache"
)

type testClient struct {
	result *time.Time
}

func (f testClient) ExpireTime(_ context.Context, _ string, _ string) (time.Time, error) {
	return *f.result, nil
}

type errTestClient struct {
}

func (f errTestClient) ExpireTime(_ context.Context, _ string, _ string) (time.Time, error) {
	return time.Now(), fmt.Errorf("failed to get domain info blah")
}

func TestCachedClient(t *testing.T) {
	ctx := context.Background()
	cache := cache.New(1*time.Minute, 1*time.Minute)
	expected := time.Now()
	domain := "foo.bar"
	host := ""

	cli := NewCachedClient(testClient{result: &expected}, cache)

	// test getting from out fake client
	t.Run("get fresh", func(t *testing.T) {
		res, err := cli.ExpireTime(ctx, domain, host)
		is := is.New(t)
		is.NoErr(err)           // expected an error
		is.Equal(expected, res) // expected the same result
	})

	// here we change the inner fake client result, but the result
	// should be the cached one
	t.Run("get from cache", func(t *testing.T) {
		oldExpected := expected
		expected = time.Now()
		res, err := cli.ExpireTime(ctx, domain, host)
		is := is.New(t)
		is.NoErr(err)              // expected an error
		is.Equal(oldExpected, res) // expected the same result
	})

	// here we flush the cache and verify that the result is the one
	// from the fake client
	t.Run("flush cache", func(t *testing.T) {
		cache.Flush()
		res, err := cli.ExpireTime(ctx, domain, host)
		is := is.New(t)
		is.NoErr(err)           // expected an error
		is.Equal(expected, res) // expected the same result
	})

	t.Run("do not cache errors", func(t *testing.T) {
		cache.Flush()
		is := is.New(t)

		cli := NewCachedClient(errTestClient{}, cache)
		_, err := cli.ExpireTime(ctx, domain, host)
		is.True(err != nil) // expected an error

		_, err = cli.ExpireTime(ctx, domain, host)
		is.True(err != nil) // expected an error

		cached, got := cache.Get(domain)
		is.True(cached == nil) // expected a nil result
		is.Equal(got, false)   // expect it not to get from cache
	})
}
