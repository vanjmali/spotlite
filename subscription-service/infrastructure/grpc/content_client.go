package grpc

import (
	"context"
	"fmt"

	pb "github.com/vanjmali/spotlite/common-lib/proto/content_service"
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"google.golang.org/grpc"
)

type GrpcContentEntityGetter struct {
	client pb.GetContentEntityClient
}

func NewGrpcContentEntityGetter(gc *grpc.ClientConn) *GrpcContentEntityGetter {
	return &GrpcContentEntityGetter{
		client: pb.NewGetContentEntityClient(gc),
	}
}

func (g *GrpcContentEntityGetter) GetEntity(ctx context.Context, entityID string, subType subscription.SubscriptionType) (string, error) {
	req := &pb.EntityIDRequest{EntityId: entityID}

	switch subType {
	case subscription.GenreSubscription:
		resp, err := g.client.GetGenre(ctx, req)
		if err != nil {
			return "", err
		}
		return resp.GetName(), nil

	case subscription.ArtistSubscription:
		resp, err := g.client.GetArtist(ctx, req)
		if err != nil {
			return "", err
		}
		return resp.GetName(), nil

	default:
		return "", fmt.Errorf("unsupported subscription type: %s", subType)
	}
}
