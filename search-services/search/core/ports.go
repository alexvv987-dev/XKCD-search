package core

import (
	"context"
)

type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]Comics, error)
	SearchIndex(ctx context.Context, query string, limit int) ([]Comics, error)
}

type DB interface {
	Get(context.Context) ([]Comics, error)
	GetByIDs(ctx context.Context, ids []int) ([]Comics, error)
	Search(ctx context.Context, words []string) ([]Comics, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type Indexer interface {
	Index(ctx context.Context) error
}
