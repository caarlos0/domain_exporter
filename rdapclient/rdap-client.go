package rdapclient

import "time"

type RdapClient interface {
	ExpireTime(domain string) (time.Time, error)
}
