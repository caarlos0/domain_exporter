package client

import (
	"context"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
)

type cachedClient struct {
	client Client
	cache  *cache.Cache
}

// NewCachedClient returns a new cached client.
func NewCachedClient(client Client, cache *cache.Cache) Client {
	return cachedClient{
		client: client,
		cache:  cache,
	}
}

func (c cachedClient) ExpireTime(ctx context.Context, domain string, host string) (time.Time, error) {
	cached, found := c.cache.Get(domain)
	if found {
		log.Debug().Msgf("using result from cache for %s", domain)
		return cached.(time.Time), nil
	}
	log.Debug().Msgf("getting live result for %s", domain)
	live, err := c.client.ExpireTime(ctx, domain, host)
	if err == nil {
		log.Debug().Msgf("caching result for %s", domain)
		c.cache.Set(domain, live, cache.DefaultExpiration)
		return live, nil
	}

	log.Debug().Err(err).Msgf("not caching %s because it errored", domain)
	return live, err
}
