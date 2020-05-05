package client

import (
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/common/log"
)

// NewCachedClient returns a new cached client.
func NewCachedClient(client Client, cache *cache.Cache) Client {
	return cachedClient{
		client: client,
		cache:  cache,
	}
}

type cachedClient struct {
	client Client
	cache  *cache.Cache
}

func (c cachedClient) ExpireTime(domain string) (time.Time, error) {
	cached, found := c.cache.Get(domain)
	if found {
		log.Debugf("using result from cache for %s", domain)
		return cached.(time.Time), nil
	}
	log.Debugf("using result from whois for %s", domain)
	live, err := c.client.ExpireTime(domain)
	if err == nil {
		log.Debugf("not caching %s because it errored", domain)
		c.cache.Set(domain, live, cache.DefaultExpiration)
	}
	return live, err
}
