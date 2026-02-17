package grpc

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/logging"
	pb "github.com/vanjmali/spotlite/common-lib/proto/content_service"
	"github.com/vanjmali/spotlite/content/services"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ContentServer struct {
	pb.UnimplementedGetContentEntityServer

	gs *services.GenreService
	as *services.ArtistService
}

func NewContentServer(gs *services.GenreService, as *services.ArtistService) *ContentServer {
	return &ContentServer{
		gs: gs,
		as: as,
	}
}

func (s *ContentServer) GetArtist(ctx context.Context, req *pb.EntityIDRequest) (*pb.GetContentEntityResponse, error) {
	a, err := s.as.FindArtistByID(ctx, req.GetEntityId())
	if err != nil {
		logging.Errorf(ctx, "failed while fetching artist: %v", err)

		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			return nil, status.Error(codes.InvalidArgument, "Invalid artist ID format")
		case errors.Is(err, services.ErrArtistNotFound):
			return nil, status.Error(codes.NotFound, "Artist not found")
		default:
			return nil, status.Error(codes.Internal, "An unexpected error has occurred while fetching artist")
		}
	}

	return &pb.GetContentEntityResponse{Name: a.Name}, nil
}

func (s *ContentServer) GetGenre(ctx context.Context, req *pb.EntityIDRequest) (*pb.GetContentEntityResponse, error) {
	g, err := s.gs.FindGenreByID(ctx, req.GetEntityId())
	if err != nil {
		logging.Errorf(ctx, "failed while fetching genre: %v", err)

		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			return nil, status.Error(codes.InvalidArgument, "Invalid genre ID format")
		case errors.Is(err, services.ErrGenreNotFound):
			return nil, status.Error(codes.NotFound, "Genre not found")
		default:
			return nil, status.Error(codes.Internal, "An unexpected error has occurred while fetching genre")
		}
	}

	return &pb.GetContentEntityResponse{Name: g.Name}, nil
}
