package initiator

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"yadro.com/course/search/core"
)

const (
	subscribeTopic = "xkcd.db.*"
	updatedTopic   = "xkcd.db.updated"
	droppedTopic   = "xkcd.db.dropped"
)

type Initiator struct {
	log     *slog.Logger
	indexer core.Indexer
	ttl     time.Duration
	ch      chan *nats.Msg
	sub     *nats.Subscription
	nc      *nats.Conn
}

func NewInitiator(log *slog.Logger, indexer core.Indexer, ttl time.Duration, nc *nats.Conn) (*Initiator, error) {
	if ttl <= 0 {
		log.Error("ttl must be greater than 0")
		return nil, core.ErrInternalServerError
	}
	ch := make(chan *nats.Msg, 1)
	sub, err := nc.ChanSubscribe(subscribeTopic, ch)
	if err != nil {
		return nil, core.ErrInternalServerError
	}
	return &Initiator{
		log:     log,
		indexer: indexer,
		ttl:     ttl,
		ch:      ch,
		sub:     sub,
		nc:      nc,
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
		case msg := <-i.ch:
			switch msg.Subject {
			case updatedTopic:
				if err := i.indexer.Index(ctx); err != nil {
					i.log.Error("Failed to index", "error", err)
				}
			case droppedTopic:
				if err := i.indexer.ResetIndex(ctx); err != nil {
					i.log.Error("Failed to reset index", "error", err)
				}
			}
		case <-ctx.Done():
			if err := i.sub.Unsubscribe(); err != nil {
				i.log.Error("Failed to unsubscribe", "error", err)
			}
			return
		}
	}
}
