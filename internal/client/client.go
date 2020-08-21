package client

import "time"

// Client is a DNS client impl.
type Client interface {
	ExpireTime(domain string) (time.Time, error)
}
