package grpc

import (
	"context"

	pb "github.com/vanjmali/spotlite/common-lib/proto/content_service"
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

func (g *GrpcContentEntityGetter) GetSong(ctx context.Context, songID string) (string, error) {
	req := &pb.EntityIDRequest{EntityId: songID}

	resp, err := g.client.GetSong(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.GetName(), nil
}
