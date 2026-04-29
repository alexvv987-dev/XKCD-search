package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

func NewServer(service core.Updater) *Server {
	return &Server{service: service}
}

type Server struct {
	updatepb.UnimplementedUpdateServer
	service core.Updater
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, core.ErrBadArguments):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, core.ErrTooLarge):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, core.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Status(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatusReply, error) {
	serviceStatus := s.service.Status(ctx)
	switch serviceStatus {
	case core.StatusRunning:
		return &updatepb.StatusReply{Status: updatepb.Status_STATUS_RUNNING}, nil
	case core.StatusIdle:
		return &updatepb.StatusReply{Status: updatepb.Status_STATUS_IDLE}, nil
	default:
		return &updatepb.StatusReply{Status: updatepb.Status_STATUS_UNSPECIFIED}, nil
	}
}

func (s *Server) Update(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.service.Update(ctx); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Stats(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatsReply, error) {
	serviceStats, err := s.service.Stats(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &updatepb.StatsReply{
		WordsTotal:    int64(serviceStats.WordsTotal),
		WordsUnique:   int64(serviceStats.WordsUnique),
		ComicsTotal:   int64(serviceStats.ComicsTotal),
		ComicsFetched: int64(serviceStats.ComicsFetched),
	}, nil
}

func (s *Server) Drop(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.service.Drop(ctx); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
