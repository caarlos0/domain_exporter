package client

import "time"

type Client interface {
	ExpireTime(domain string) (time.Time, error)
}
