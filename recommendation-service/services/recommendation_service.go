package services

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/recommendation-service/dtos"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Error types for recommendation service
var (
	ErrGraphDatabaseUnavailable = errors.New("graph database is currently unavailable")
	ErrLimitOutOfRange          = errors.New("limit must be between 1 and 50")
	ErrInvalidUserID            = errors.New("invalid user ID format")
)

// RecommendationService provides recommendation-related business logic
type RecommendationService struct {
	services *Services
	tracer   trace.Tracer
}

// NewRecommendationService constructs a RecommendationService
func NewRecommendationService(services *Services) *RecommendationService {
	return &RecommendationService{
		services: services,
		tracer:   otel.Tracer("recommendation-service/recommendation-service"),
	}
}

// GetPersonalizedRecommendations returns personalized recommendations combining
// content-based and collaborative filtering
func (rs *RecommendationService) GetPersonalizedRecommendations(
	ctx context.Context,
	userID string,
	limit int,
) ([]dtos.RecommendedSongDto, error) {
	ctx, span := rs.tracer.Start(ctx, "GetPersonalizedRecommendations")
	defer span.End()

	if limit <= 0 || limit > 50 {
		return nil, ErrLimitOutOfRange
	}

	// Get content-based recommendations (subscribed genres + similar rated songs)
	contentBasedIDs, err := rs.services.relationRepository.GetRecommendedSongsForUser(ctx, userID, limit)
	if err != nil {
		logging.Errorf(ctx, "failed to get content-based recommendations: %v", err)
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get content-based recommendations: %w", err)
	}

	// Get collaborative filtering recommendations (similar users' preferences)
	collaborativeIDs, err := rs.services.relationRepository.GetCollaborativeRecommendations(ctx, userID, limit)
	if err != nil {
		logging.Errorf(ctx, "failed to get collaborative recommendations: %v", err)
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get collaborative recommendations: %w", err)
	}

	// Merge and deduplicate results
	mergedSongs := rs.mergeSongRecommendations(contentBasedIDs, collaborativeIDs)
	if len(mergedSongs) > limit {
		mergedSongs = mergedSongs[:limit]
	}

	// Enrich with rating information
	recommendations, err := rs.enrichSongsWithRatings(ctx, mergedSongs)
	if err != nil {
		logging.Errorf(ctx, "failed to enrich songs with ratings: %v", err)
		span.RecordError(err)
		return nil, fmt.Errorf("failed to enrich songs: %w", err)
	}

	return recommendations, nil
}

// GetTrendingSongs returns highly-rated songs (collaborative/popularity-based)
func (rs *RecommendationService) GetTrendingSongs(
	ctx context.Context,
	limit int,
) ([]dtos.RecommendedSongDto, error) {
	ctx, span := rs.tracer.Start(ctx, "GetTrendingSongs")
	defer span.End()

	if limit <= 0 || limit > 50 {
		return nil, ErrLimitOutOfRange
	}

	// Get highly-rated songs (min 4.0 rating, at least 3 ratings)
	songIDs, err := rs.services.relationRepository.GetHighlyRatedSongs(ctx, 4.0, limit)
	if err != nil {
		logging.Errorf(ctx, "failed to get trending songs: %v", err)
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get trending songs: %w", err)
	}

	// Enrich with rating information
	recommendations, err := rs.enrichSongsWithRatings(ctx, songIDs)
	if err != nil {
		logging.Errorf(ctx, "failed to enrich trending songs: %v", err)
		span.RecordError(err)
		return nil, fmt.Errorf("failed to enrich songs: %w", err)
	}

	return recommendations, nil
}

// mergeSongRecommendations merges two lists of song IDs, deduplicating and combining them
func (rs *RecommendationService) mergeSongRecommendations(contentBased, collaborative []string) []string {
	seen := make(map[string]bool)
	var merged []string

	// Add content-based recommendations (higher priority)
	for _, id := range contentBased {
		if !seen[id] {
			merged = append(merged, id)
			seen[id] = true
		}
	}

	// Add collaborative recommendations that aren't already included
	for _, id := range collaborative {
		if !seen[id] {
			merged = append(merged, id)
			seen[id] = true
		}
	}

	return merged
}

// enrichSongsWithRatings enriches song IDs with rating information and artist names
func (rs *RecommendationService) enrichSongsWithRatings(
	ctx context.Context,
	songIDs []string,
) ([]dtos.RecommendedSongDto, error) {
	ctx, span := rs.tracer.Start(ctx, "enrichSongsWithRatings")
	defer span.End()

	recommendations := make([]dtos.RecommendedSongDto, 0, len(songIDs))

	// If songNodeRepository is not available, return basic DTOs
	if rs.services == nil || rs.services.songNodeRepository == nil {
		for _, songID := range songIDs {
			recommendations = append(recommendations, dtos.RecommendedSongDto{
				SongID:        songID,
				Title:         "Unknown",
				ArtistNames:   []string{},
				AverageRating: 0.0,
				RatingCount:   0,
				Reason:        "recommended",
			})
		}
		return recommendations, nil
	}

	var wg sync.WaitGroup
	ordered := make([]*dtos.RecommendedSongDto, len(songIDs))

	// Process songs in parallel for better performance
	semaphore := make(chan struct{}, 5) // Limit to 5 concurrent operations

	for idx, songID := range songIDs {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			song, err := rs.services.songNodeRepository.Get(ctx, id)
			if err != nil {
				logging.Errorf(ctx, "failed to get song %s: %v", id, err)
				ordered[i] = &dtos.RecommendedSongDto{
					SongID:      id,
					Title:       "Unknown",
					ArtistNames: []string{},
					Reason:      "recommended",
				}
				return
			}

			if song == nil {
				return
			}

			// Fetch rating information from graph database
			var avgRating float64
			var ratingCount int64
			if rs.services.relationRepository != nil {
				var err error
				avgRating, ratingCount, err = rs.services.relationRepository.GetSongRatingStats(ctx, id)
				if err != nil {
					logging.Warnf(ctx, "failed to get rating stats for song %s: %v", id, err)
					// Continue with zero values on error
					avgRating = 0.0
					ratingCount = 0
				}
			}

			// Fetch artist names from graph database
			var artistNames []string
			if rs.services.relationRepository != nil {
				var err error
				artistNames, err = rs.services.relationRepository.GetSongArtists(ctx, id)
				if err != nil {
					logging.Warnf(ctx, "failed to get artists for song %s: %v", id, err)
					// Continue with empty list on error
					artistNames = []string{}
				}
			}

			// If no artists found, use empty list
			if artistNames == nil {
				artistNames = []string{}
			}

			dto := &dtos.RecommendedSongDto{
				SongID:        id,
				Title:         song.Title,
				ArtistNames:   artistNames,
				AverageRating: avgRating,
				RatingCount:   ratingCount,
				Reason:        "recommended",
			}

			ordered[i] = dto
		}(idx, songID)
	}

	wg.Wait()

	for _, dto := range ordered {
		if dto != nil {
			recommendations = append(recommendations, *dto)
		}
	}

	return recommendations, nil
}
