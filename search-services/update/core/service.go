package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	running     atomic.Bool
	notFound    atomic.Int64
	publisher   Publisher
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, concurrency int, publisher Publisher,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
		publisher:   publisher,
	}, nil
}

func (s *Service) Update(ctx context.Context) (err error) {
	if s.running.Swap(true) {
		return ErrAlreadyExists
	}
	defer s.running.Store(false)
	comicsNum, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("failed to fetch comic number", "err", err)
		return err
	}
	currDbIds, err := s.db.IDs(ctx)
	if err != nil {
		s.log.Error("error getting database ids", "error", err)
		return err
	}
	dbIds := map[int]struct{}{}
	for _, num := range currDbIds {
		dbIds[num] = struct{}{}
	}
	ids := make(chan int)
	go func() {
		defer close(ids)
		for id := 1; id <= comicsNum; id++ {
			if _, ok := dbIds[id]; !ok {
				ids <- id
			}
		}
	}()
	var added atomic.Int64
	wg := sync.WaitGroup{}
	s.notFound.Store(0)
	for i := 0; i < s.concurrency; i++ {
		wg.Go(func() {
			for id := range ids {
				comics, err := s.xkcd.Get(ctx, id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						s.notFound.Add(1)
						continue
					}
					s.log.Error("get comics", "id", id, "err", err)
					continue
				}
				normalizedDescription, err := s.words.Norm(ctx, comics.Title+" "+comics.SafeTitle+" "+comics.Description+" "+comics.Transcript)
				if err != nil {
					s.log.Error("norm", "id", id, "err", err)
					continue
				}
				err = s.db.Add(ctx, Comics{ID: id, URL: comics.URL, Words: normalizedDescription})
				if err != nil {
					s.log.Error("add comics", "id", id, "err", err)
					continue
				}
				added.Add(1)
			}
		})
	}
	wg.Wait()
	if added.Load() > 0 {
		if err = s.publisher.Publish(ctx); err != nil {
			s.log.Error("publish comics", "err", err)
			return err
		}
	}
	return nil
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	stats, err := s.db.Stats(ctx)
	if err != nil {
		s.log.Error("error getting stats", "err", err)
		return ServiceStats{}, err
	}
	comicsTotal, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("error getting comics total", "err", err)
		return ServiceStats{}, err
	}
	return ServiceStats{
		DBStats:     stats,
		ComicsTotal: comicsTotal - int(s.notFound.Load()),
	}, nil
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	if s.running.Load() {
		return StatusRunning
	}
	return StatusIdle
}

func (s *Service) waitIdle(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for s.running.Load() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
	return nil
}

func (s *Service) Drop(ctx context.Context) error {
	if err := s.waitIdle(ctx); err != nil {
		s.log.Error("error waiting for update to finish", "err", err)
		return err
	}
	if err := s.db.Drop(ctx); err != nil {
		s.log.Error("error dropping database", "err", err)
		return err
	}
	if err := s.publisher.PublishDrop(ctx); err != nil {
		s.log.Error("publish comics", "err", err)
		return err
	}
	s.log.Info("dropped")
	return nil
}
