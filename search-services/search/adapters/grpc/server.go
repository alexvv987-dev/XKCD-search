package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

const maxLimit int64 = 100

func NewServer(service core.Searcher) *Server {
	return &Server{service: service}
}

type Server struct {
	searchpb.UnimplementedSearchServer
	service core.Searcher
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, core.ErrBadArguments):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, core.ErrTooLarge):
		return status.Error(codes.ResourceExhausted, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Search(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	if len(in.GetPhrase()) == 0 {
		return nil, toGRPCError(core.ErrBadArguments)
	}
	limit := in.GetLimit()
	if limit <= 0 || limit > maxLimit {
		return nil, toGRPCError(core.ErrBadArguments)
	}
	comics, err := s.service.Search(ctx, in.GetPhrase(), int(in.GetLimit()))
	if err != nil {
		return nil, toGRPCError(err)
	}
	pbComics := make([]*searchpb.Comic, 0, len(comics))
	for _, comic := range comics {
		pbComics = append(pbComics, &searchpb.Comic{
			Id:  int64(comic.ID),
			Url: comic.URL,
		})
	}
	return &searchpb.SearchReply{
		Comics: pbComics,
		Total:  int64(len(pbComics)),
	}, nil
}

func (s *Server) SearchIndex(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	if len(in.GetPhrase()) == 0 {
		return nil, toGRPCError(core.ErrBadArguments)
	}
	limit := in.GetLimit()
	if limit <= 0 || limit > maxLimit {
		return nil, toGRPCError(core.ErrBadArguments)
	}
	comics, err := s.service.SearchIndex(ctx, in.GetPhrase(), int(in.GetLimit()))
	if err != nil {
		return nil, toGRPCError(err)
	}
	pbComics := make([]*searchpb.Comic, 0, len(comics))
	for _, comic := range comics {
		pbComics = append(pbComics, &searchpb.Comic{
			Id:  int64(comic.ID),
			Url: comic.URL,
		})
	}
	return &searchpb.SearchReply{
		Comics: pbComics,
		Total:  int64(len(pbComics)),
	}, nil
}