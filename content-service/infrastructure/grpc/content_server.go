package grpc

import (
	"context"

	pb "github.com/vanjmali/spotlite/common-lib/proto/content_service"
	"github.com/vanjmali/spotlite/content/services"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ContentServer struct {
	pb.UnimplementedContentCheckerServer

	GenreService  *services.GenreService
	ArtistService *services.ArtistService
}

func (s *ContentServer) CheckGenreExistence(ctx context.Context, req *pb.CheckIdRequest) (*pb.ExistenceResponse, error) {
	exists, err := s.GenreService.Exists(ctx, req.GetEntityId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "genre existence check failed: %v", err)
	}

	return &pb.ExistenceResponse{Exists: exists}, nil
}

func (s *ContentServer) CheckArtistExistence(ctx context.Context, req *pb.CheckIdRequest) (*pb.ExistenceResponse, error) {
	exists, err := s.ArtistService.Exists(ctx, req.GetEntityId())

	if err != nil {
		return nil, status.Errorf(codes.Internal, "artist existence check failed: %v", err)
	}

	return &pb.ExistenceResponse{Exists: exists}, nil
}
