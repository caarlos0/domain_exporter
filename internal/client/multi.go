package client

import "time"

type multiClient []Client

func (clients multiClient) ExpireTime(domain string) (time.Time, error) {
	var t time.Time
	var err error
	for _, client := range clients {
		t, err = client.ExpireTime(domain)
		if err == nil {
			break
		}
	}
	return t, err
}

// NewMultiClient returns a client that wraps multiple clients.
// It returns the first success, or, if all clients fail, the latest failure.
func NewMultiClient(clients ...Client) Client {
	return multiClient(clients)
}
