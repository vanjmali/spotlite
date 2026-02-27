package handlers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	pb "github.com/vanjmali/spotlite/common-lib/proto/rating_service"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/content/entities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	ratingSummaryClientOnce sync.Once
	ratingSummaryClient     pb.GetSongRatingClient
	ratingSummaryClientErr  error
	ratingGrpcAddress       = utils.GetEnv("RATING_GRPC_ADDRESS", "rating-service:50052")
)

func getSongRatings(ctx context.Context, songs []entities.Song) {
	if len(songs) == 0 {
		return
	}

	client, err := getRatingSummaryClient()
	if err != nil {
		return
	}

	const maxConcurrent = 8
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i := range songs {
		i := i
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			avg, count, ok := fetchSongSummary(ctx, client, songs[i].ID.Hex())
			if !ok {
				return
			}
			songs[i].Rating = &entities.SongRating{
				Average: avg,
				Count:   count,
			}
		}()
	}

	wg.Wait()
}

func getSongRating(ctx context.Context, song *entities.Song) {
	if song == nil {
		return
	}

	client, err := getRatingSummaryClient()
	if err != nil {
		return
	}

	avg, count, ok := fetchSongSummary(ctx, client, song.ID.Hex())
	if !ok {
		return
	}
	song.Rating = &entities.SongRating{
		Average: avg,
		Count:   count,
	}
}

func getRatingSummaryClient() (pb.GetSongRatingClient, error) {
	ratingSummaryClientOnce.Do(func() {
		rootPath := utils.MustGetEnv("ROOT_CERT_PATH")
		pool := x509.NewCertPool()
		rootPEM, err := os.ReadFile(rootPath)
		if err != nil {
			ratingSummaryClientErr = fmt.Errorf("failed to read root cert: %w", err)
			return
		}
		if ok := pool.AppendCertsFromPEM(rootPEM); !ok {
			ratingSummaryClientErr = fmt.Errorf("failed to parse root cert at %s", rootPath)
			return
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
			ratingSummaryClientErr = fmt.Errorf("failed to connect to rating grpc: %w", err)
			return
		}

		ratingSummaryClient = pb.NewGetSongRatingClient(conn)
	})

	return ratingSummaryClient, ratingSummaryClientErr
}

func fetchSongSummary(
	ctx context.Context,
	client pb.GetSongRatingClient,
	songID string,
) (float64, int64, bool) {
	reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	summary, err := client.GetSongRatingSummary(reqCtx, &pb.SongIDRequest{
		SongId: songID,
	})
	if err != nil || summary == nil {
		return 0, 0, false
	}

	return summary.GetAverage(), summary.GetCount(), true
}
