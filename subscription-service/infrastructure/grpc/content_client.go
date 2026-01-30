package grpc

import (
	"context"
	"fmt"

	pb "github.com/vanjmali/spotlite/common-lib/proto/content_service"
	"github.com/vanjmali/spotlite/subscriptions/entities"
	"google.golang.org/grpc"
)

type GrpcContentChecker struct {
	client pb.ContentCheckerClient
}

func NewGrpcContentChecker(gc *grpc.ClientConn) *GrpcContentChecker {
	return &GrpcContentChecker{
		client: pb.NewContentCheckerClient(gc),
	}
}

func (g *GrpcContentChecker) CheckExistence(ctx context.Context, entityID string, subType entities.SubscriptionType) (bool, error) {
	req := &pb.CheckIdRequest{EntityId: entityID}

	switch subType {
	case entities.GenreSubscription:
		resp, err := g.client.CheckGenreExistence(ctx, req)
		if err != nil {
			return false, err
		}
		return resp.GetExists(), nil

	case entities.ArtistSubscription:
		resp, err := g.client.CheckArtistExistence(ctx, req)
		if err != nil {
			return false, err
		}
		return resp.GetExists(), nil

	default:
		return false, fmt.Errorf("unsupported subscription type: %s", subType)
	}
}
