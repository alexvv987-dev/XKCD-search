package broker

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go"
	"yadro.com/course/update/core"
)

const (
	publishSubject     = "xkcd.db.updated"
	publishPayload     = "updated"
	publishDropSubject = "xkcd.db.dropped"
	publishDropPayload = "dropped"
)

type Publisher struct {
	log *slog.Logger
	nc  *nats.Conn
}

func NewPublisher(log *slog.Logger, address string) (*Publisher, error) {
	nc, err := nats.Connect(address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, core.ErrInternalServerError
	}
	return &Publisher{
		log: log,
		nc:  nc,
	}, nil
}

func (p *Publisher) publish(subject string, data []byte) error {
	if err := p.nc.Publish(subject, data); err != nil {
		p.log.Error("publish problem", "error", err)
		return core.ErrInternalServerError
	}
	if err := p.nc.Flush(); err != nil {
		p.log.Error("flush problem", "error", err)
		return core.ErrInternalServerError
	}
	p.log.Info("publish ok")
	return nil
}

func (p *Publisher) Publish(ctx context.Context) error {
	return p.publish(publishSubject, []byte(publishPayload))
}

func (p *Publisher) PublishDrop(ctx context.Context) error {
	return p.publish(publishDropSubject, []byte(publishDropPayload))
}

func (p *Publisher) Close() {
	p.nc.Close()
	p.log.Info("publisher closed")
}
