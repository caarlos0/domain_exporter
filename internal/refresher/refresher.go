package refresher

import (
	"context"
	"time"

	"github.com/caarlos0/domain_exporter/internal/client"
	"github.com/rs/zerolog/log"
)

type Refresher struct {
	ticker  *time.Ticker
	client  client.Client
	domains []string
}

func New(interval time.Duration, client client.Client, domains ...string) Refresher {
	ticker := time.NewTicker(interval)
	return Refresher{
		ticker:  ticker,
		client:  client,
		domains: domains,
	}
}

func (r Refresher) Stop() {
	r.ticker.Stop()
}

func (r Refresher) Run(ctx context.Context) {
	log.Info().Msg("run refresher")
	r.Refresh(ctx)

	select {
	case <-r.ticker.C:
		r.Refresh(ctx)
	case <-ctx.Done():
		log.Info().Msg("refresher is finished")
		return
	}
}

func (r Refresher) Refresh(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	for _, domain := range r.domains {
		if _, err := r.client.ExpireTime(ctx, domain); err != nil {
			log.Error().Err(err).Msgf("failed to get expire time for %s", domain)
		}
	}
	log.Debug().Msg("refresh is done")
}
