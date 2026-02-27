package grpc

import (
	"context"
	"errors"

	"github.com/vanjmali/spotlite/common-lib/logging"
	pb "github.com/vanjmali/spotlite/common-lib/proto/rating_service"
	"github.com/vanjmali/spotlite/rating-service/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RatingServer struct {
	pb.UnimplementedGetSongRatingServer
	rs *services.RatingService
}

func NewRatingServer(rs *services.RatingService) *RatingServer {
	return &RatingServer{rs: rs}
}

func (s *RatingServer) GetSongRatingSummary(
	ctx context.Context,
	req *pb.SongIDRequest,
) (*pb.SongRatingSummaryResponse, error) {
	summary, err := s.rs.GetAverageRatingBySongID(ctx, req.GetSongId())
	if err != nil {
		logging.Errorf(ctx, "failed while fetching song rating summary: %v", err)
		if errors.Is(err, services.ErrObjectIdCastFailed) {
			return nil, status.Error(codes.InvalidArgument, "invalid song ID format")
		}
		return nil, status.Error(codes.Internal, "an unexpected error has occurred while fetching song rating summary")
	}

	return &pb.SongRatingSummaryResponse{
		Average: summary.Avg,
		Count:   summary.Count,
	}, nil
}
