package core

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
)

type Service struct {
	log         *slog.Logger
	db          DB
	words       Words
	concurrency int
}

type searchResult struct {
	comic Comics
	score int
}

func NewService(
	log *slog.Logger, db DB, words Words, concurrency int,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		words:       words,
		concurrency: concurrency,
	}, nil
}

func (s *Service) Search(ctx context.Context, query string, limit int) ([]Comics, error) {
	words, err := s.words.Norm(ctx, query)
	if err != nil {
		s.log.Error("Failed to norm words", "query", query, "err", err)
		return nil, err
	}
	comics, err := s.db.Get(ctx)
	if err != nil {
		s.log.Error("Failed to get comics", "query", query, "err", err)
		return nil, err
	}
	results := []searchResult{}
	queryWords := make(map[string]struct{}, len(words))
	for _, word := range words {
		queryWords[word] = struct{}{}
	}
	wg := sync.WaitGroup{}
	var mu sync.Mutex
	jobs := make(chan Comics, len(comics))
	for _, comic := range comics {
		jobs <- comic
	}
	close(jobs)
	for i := 0; i < s.concurrency; i++ {
		wg.Go(func() {
			var score int
			for comic := range jobs {
				for _, word := range comic.Words {
					if _, ok := queryWords[word]; ok {
						score++
					}
				}
				if score > 0 {
					mu.Lock()
					results = append(results, searchResult{comic, score})
					mu.Unlock()
				}
				score = 0
			}
		})
	}
	wg.Wait()
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	resultComics := []Comics{}
	for num, result := range results {
		if num >= limit {
			break
		}
		resultComics = append(resultComics, result.comic)
	}
	return resultComics, nil
}
