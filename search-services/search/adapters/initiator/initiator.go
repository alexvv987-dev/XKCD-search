package initiator

import (
	"context"
	"log/slog"
	"time"

	"yadro.com/course/search/core"
)

type Initiator struct {
	log     *slog.Logger
	indexer core.Indexer
	ttl     time.Duration
}

func NewInitiator(log *slog.Logger, indexer core.Indexer, ttl time.Duration) (*Initiator, error) {
	if ttl <= 0 {
		log.Error("ttl must be greater than 0")
		return nil, core.ErrInternalServerError
	}
	return &Initiator{
		log:     log,
		indexer: indexer,
		ttl:     ttl,
	}, nil
}

func (i *Initiator) Start(ctx context.Context) {
	if err := i.indexer.Index(ctx); err != nil {
		i.log.Error("Failed to index", "error", err)
	}
	ticker := time.NewTicker(i.ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := i.indexer.Index(ctx); err != nil {
				i.log.Error("Failed to index", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
