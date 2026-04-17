package core

import (
	"context"
)

type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]Comics, error)
}

type DB interface {
	Get(context.Context) ([]Comics, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}
