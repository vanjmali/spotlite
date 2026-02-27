package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	pb "github.com/vanjmali/spotlite/common-lib/proto/rating_service"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewRatingSummaryClient() (pb.GetSongRatingClient, *grpc.ClientConn, error) {
	ratingGrpcAddress := utils.MustGetEnv("RATING_GRPC_ADDRESS")
	rootPath := utils.MustGetEnv("ROOT_CERT_PATH")

	pool := x509.NewCertPool()
	rootPEM, err := os.ReadFile(rootPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read root cert: %w", err)
	}
	if ok := pool.AppendCertsFromPEM(rootPEM); !ok {
		return nil, nil, fmt.Errorf("failed to parse root cert at %s", rootPath)
	}

	tlsConfig := &tls.Config{
		RootCAs:    pool,
		ServerName: "rating-service",
		MinVersion: tls.VersionTLS13,
	}

	conn, err := grpc.NewClient(
		ratingGrpcAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to rating grpc: %w", err)
	}

	return pb.NewGetSongRatingClient(conn), conn, nil
}
