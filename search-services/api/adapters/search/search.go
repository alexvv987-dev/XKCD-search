package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpb.SearchClient
}

func toAppError(err error) error {
	switch status.Code(err) {
	case codes.ResourceExhausted:
		return core.ErrTooLarge
	case codes.InvalidArgument:
		return core.ErrBadArguments
	case codes.NotFound:
		return core.ErrNotFound
	default:
		return core.ErrInternalServerError
	}
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: searchpb.NewSearchClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	if _, err := c.client.Ping(ctx, &emptypb.Empty{}); err != nil {
		return toAppError(err)
	}
	return nil
}

func (c Client) Search(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	searched, err := c.client.Search(ctx, &searchpb.SearchRequest{Phrase: phrase, Limit: int64(limit)})
	if err != nil {
		return nil, toAppError(err)
	}
	comics := make([]core.Comics, 0, len(searched.GetComics()))
	for _, comic := range searched.GetComics() {
		comics = append(comics, core.Comics{ID: int(comic.GetId()), URL: comic.GetUrl()})
	}
	return comics, nil
}
