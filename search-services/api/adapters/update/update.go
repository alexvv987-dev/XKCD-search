package update

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

type Client struct {
	log    *slog.Logger
	client updatepb.UpdateClient
}

func toAppError(err error) error {
	switch status.Code(err) {
	case codes.ResourceExhausted:
		return core.ErrTooLarge
	case codes.InvalidArgument:
		return core.ErrBadArguments
	case codes.AlreadyExists:
		return core.ErrAlreadyExists
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
		client: updatepb.NewUpdateClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	if _, err := c.client.Ping(ctx, &emptypb.Empty{}); err != nil {
		return toAppError(err)
	}
	return nil
}

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	clientStatus, err := c.client.Status(ctx, &emptypb.Empty{})
	if err != nil {
		return core.StatusUpdateUnknown, toAppError(err)
	}
	switch clientStatus.GetStatus() {
	case updatepb.Status_STATUS_RUNNING:

		return core.StatusUpdateRunning, nil
	case updatepb.Status_STATUS_IDLE:
		return core.StatusUpdateIdle, nil
	default:
		return core.StatusUpdateUnknown, nil
	}
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	clientStats, err := c.client.Stats(ctx, &emptypb.Empty{})
	if err != nil {
		return core.UpdateStats{}, toAppError(err)
	}
	return core.UpdateStats{
		WordsTotal:    int(clientStats.GetWordsTotal()),
		WordsUnique:   int(clientStats.GetWordsUnique()),
		ComicsTotal:   int(clientStats.GetComicsTotal()),
		ComicsFetched: int(clientStats.GetComicsFetched()),
	}, nil
}

func (c Client) Update(ctx context.Context) error {
	if _, err := c.client.Update(ctx, &emptypb.Empty{}); err != nil {
		return toAppError(err)
	}
	return nil
}

func (c Client) Drop(ctx context.Context) error {
	if _, err := c.client.Drop(ctx, &emptypb.Empty{}); err != nil {
		return toAppError(err)
	}
	return nil
}
