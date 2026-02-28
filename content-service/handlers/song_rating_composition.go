package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	pb "github.com/vanjmali/spotlite/common-lib/proto/rating_service"
	"github.com/vanjmali/spotlite/content/entities"
)

type songRatingCacheEntry struct {
	Average float64 `json:"average"`
	Count   int64   `json:"count"`
}

type SongRatingCache interface {
	GetSummary(ctx context.Context, songID string) (float64, int64, bool, error)
	SetSummary(ctx context.Context, songID string, average float64, count int64) error
}

type RedisSongRatingCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisSongRatingCache(client *redis.Client, ttl time.Duration) *RedisSongRatingCache {
	return &RedisSongRatingCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *RedisSongRatingCache) GetSummary(ctx context.Context, songID string) (float64, int64, bool, error) {
	if c == nil || c.client == nil {
		return 0, 0, false, nil
	}

	raw, err := c.client.Get(ctx, songRatingCacheKey(songID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}
	if raw == "" {
		return 0, 0, false, nil
	}

	var entry songRatingCacheEntry
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		return 0, 0, false, err
	}

	return entry.Average, entry.Count, true, nil
}

func (c *RedisSongRatingCache) SetSummary(ctx context.Context, songID string, average float64, count int64) error {
	if c == nil || c.client == nil {
		return nil
	}

	payload, err := json.Marshal(songRatingCacheEntry{
		Average: average,
		Count:   count,
	})
	if err != nil {
		return err
	}

	return c.client.Set(ctx, songRatingCacheKey(songID), payload, c.ttl).Err()
}

func getSongRatings(client pb.GetSongRatingClient, cache SongRatingCache, ctx context.Context, songs []entities.Song) {
	if len(songs) == 0 || client == nil {
		return
	}

	const maxConcurrent = 8
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i := range songs {

		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			avg, count, ok := fetchSongRatings(client, cache, ctx, songs[i].ID.Hex())
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

func getSongRating(client pb.GetSongRatingClient, cache SongRatingCache, ctx context.Context, song *entities.Song) {
	if song == nil || client == nil {
		return
	}

	avg, count, ok := fetchSongRatings(client, cache, ctx, song.ID.Hex())
	if !ok {
		return
	}

	song.Rating = &entities.SongRating{
		Average: avg,
		Count:   count,
	}
}

func fetchSongRatings(
	client pb.GetSongRatingClient,
	cache SongRatingCache,
	ctx context.Context,
	songID string,
) (float64, int64, bool) {
	if cache != nil {
		avg, count, ok, err := cache.GetSummary(ctx, songID)
		if err == nil && ok {
			return avg, count, true
		}
	}

	reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	summary, err := client.GetSongRatingSummary(reqCtx, &pb.SongIDRequest{
		SongId: songID,
	})
	if err != nil || summary == nil {
		return 0, 0, false
	}

	avg := summary.GetAverage()
	count := summary.GetCount()
	if cache != nil {
		_ = cache.SetSummary(ctx, songID, avg, count)
	}

	return avg, count, true
}

func songRatingCacheKey(songID string) string {
	return "song_rating_summary:" + songID
}
