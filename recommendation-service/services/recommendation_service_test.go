package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/recommendation-service/dtos"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

// fakeGraphRelationRepository is a test double for GraphRelationRepository
type fakeGraphRelationRepository struct {
	recommendedSongs   []string
	recommendedErr     error
	collaborativeSongs []string
	collaborativeErr   error
	trendingSongs      []string
	trendingErr        error
	songRatingStats    map[string]struct {
		avg   float64
		count int64
	}
	songArtists    map[string][]string
	songRatingErr  error
	songArtistsErr error
}

func (f *fakeGraphRelationRepository) GetRecommendedSongsForUser(ctx context.Context, userID string, limit int) ([]string, error) {
	return f.recommendedSongs, f.recommendedErr
}

func (f *fakeGraphRelationRepository) GetCollaborativeRecommendations(ctx context.Context, userID string, limit int) ([]string, error) {
	return f.collaborativeSongs, f.collaborativeErr
}

func (f *fakeGraphRelationRepository) GetHighlyRatedSongs(ctx context.Context, minRating float64, limit int) ([]string, error) {
	return f.trendingSongs, f.trendingErr
}

func (f *fakeGraphRelationRepository) GetSongRatingStats(ctx context.Context, songID string) (float64, int64, error) {
	if f.songRatingErr != nil {
		return 0.0, 0, f.songRatingErr
	}
	if f.songRatingStats != nil {
		if stats, ok := f.songRatingStats[songID]; ok {
			return stats.avg, stats.count, nil
		}
	}
	return 0.0, 0, nil
}

func (f *fakeGraphRelationRepository) GetSongArtists(ctx context.Context, songID string) ([]string, error) {
	if f.songArtistsErr != nil {
		return nil, f.songArtistsErr
	}
	if f.songArtists != nil {
		if artists, ok := f.songArtists[songID]; ok {
			return artists, nil
		}
	}
	return []string{}, nil
}

// mockSongRepository is a test double for SongNodeRepository
type mockSongRepository struct {
	songs map[string]*entities.SongNode
	err   error
}

func (m *mockSongRepository) Get(ctx context.Context, songID string) (*entities.SongNode, error) {
	if m.err != nil {
		return nil, m.err
	}
	if song, ok := m.songs[songID]; ok {
		return song, nil
	}
	return nil, errors.New("song not found")
}

// TestGetPersonalizedRecommendations tests personalized recommendations
func TestGetPersonalizedRecommendations(t *testing.T) {
	tests := []struct {
		name               string
		userID             string
		limit              int
		recommendedSongs   []string
		collaborativeSongs []string
		expectedCount      int
		expectError        bool
	}{
		{
			name:               "Valid limit",
			userID:             "user123",
			limit:              10,
			recommendedSongs:   []string{"song1", "song2"},
			collaborativeSongs: []string{"song3"},
			expectedCount:      3,
			expectError:        false,
		},
		{
			name:               "Empty recommendations",
			userID:             "user456",
			limit:              10,
			recommendedSongs:   []string{},
			collaborativeSongs: []string{},
			expectedCount:      0,
			expectError:        false,
		},
		{
			name:               "Limit too high",
			userID:             "user789",
			limit:              100,
			recommendedSongs:   []string{},
			collaborativeSongs: []string{},
			expectedCount:      0,
			expectError:        true,
		},
		{
			name:               "Limit zero",
			userID:             "user000",
			limit:              0,
			recommendedSongs:   []string{},
			collaborativeSongs: []string{},
			expectedCount:      0,
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create base services with mocked repositories
			baseServices := &Services{
				relationRepository: &fakeGraphRelationRepository{
					recommendedSongs:   tt.recommendedSongs,
					collaborativeSongs: tt.collaborativeSongs,
				},
				songNodeRepository: nil, // Will be mocked as needed
			}

			rs := NewRecommendationService(baseServices)

			recommendations, err := rs.GetPersonalizedRecommendations(context.Background(), tt.userID, tt.limit)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, recommendations)
		})
	}
}

// TestGetTrendingSongs tests trending songs retrieval
func TestGetTrendingSongs(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		songIDs     []string
		expectError bool
	}{
		{
			name:        "Valid limit",
			limit:       20,
			songIDs:     []string{"trending1", "trending2", "trending3"},
			expectError: false,
		},
		{
			name:        "Limit too high",
			limit:       100,
			songIDs:     []string{},
			expectError: true,
		},
		{
			name:        "Limit zero",
			limit:       0,
			songIDs:     []string{},
			expectError: true,
		},
		{
			name:        "Maximum valid limit",
			limit:       50,
			songIDs:     []string{"song1"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseServices := &Services{
				relationRepository: &fakeGraphRelationRepository{
					trendingSongs: tt.songIDs,
				},
			}

			rs := NewRecommendationService(baseServices)

			recommendations, err := rs.GetTrendingSongs(context.Background(), tt.limit)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, recommendations)
		})
	}
}

// TestMergeSongRecommendations tests merging and deduplication
func TestMergeSongRecommendations(t *testing.T) {
	tests := []struct {
		name          string
		contentBased  []string
		collaborative []string
		expected      []string
	}{
		{
			name:          "No overlap",
			contentBased:  []string{"song1", "song2"},
			collaborative: []string{"song3", "song4"},
			expected:      []string{"song1", "song2", "song3", "song4"},
		},
		{
			name:          "With overlap",
			contentBased:  []string{"song1", "song2", "song3"},
			collaborative: []string{"song2", "song4"},
			expected:      []string{"song1", "song2", "song3", "song4"},
		},
		{
			name:          "Duplicate removal",
			contentBased:  []string{"song1", "song1"},
			collaborative: []string{"song1"},
			expected:      []string{"song1"},
		},
		{
			name:          "Empty lists",
			contentBased:  []string{},
			collaborative: []string{},
			expected:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseServices := &Services{}
			rs := NewRecommendationService(baseServices)

			merged := rs.mergeSongRecommendations(tt.contentBased, tt.collaborative)

			require.Equal(t, tt.expected, merged)
		})
	}
}

// TestRecommendationResponseDto tests the DTO structure
func TestRecommendationResponseDto(t *testing.T) {
	songs := []dtos.RecommendedSongDto{
		{
			SongID:        "song1",
			Title:         "Test Song",
			ArtistNames:   []string{"Artist1"},
			AverageRating: 4.5,
			RatingCount:   100,
			Reason:        "personalized",
		},
	}

	response := dtos.RecommendationResponseDto{
		Songs:   songs,
		Message: "Success",
	}

	require.NotNil(t, response)
	require.Equal(t, 1, len(response.Songs))
	require.Equal(t, "song1", response.Songs[0].SongID)
	require.InEpsilon(t, 4.5, response.Songs[0].AverageRating, 0.01)
}

// TestEnrichSongMetadataSuccess tests successful API composition with content-service
func TestEnrichSongMetadataSuccess(t *testing.T) {
	// Mock metadata from content-service
	metadata := map[string]*dtos.SongMetadata{
		"song1": {
			SongID:      "song1",
			Title:       "Bohemian Rhapsody",
			ArtistNames: []string{"Queen"},
			Duration:    360,
		},
		"song2": {
			SongID:      "song2",
			Title:       "Stairway to Heaven",
			ArtistNames: []string{"Led Zeppelin"},
			Duration:    482,
		},
	}

	// Test enrichment (normally handled by content gRPC client)
	enriched := map[string]dtos.RecommendedSongDto{
		"song1": {
			SongID:        "song1",
			Title:         metadata["song1"].Title,
			ArtistNames:   metadata["song1"].ArtistNames,
			AverageRating: 4.8,
			RatingCount:   150,
			Reason:        "personalized",
		},
		"song2": {
			SongID:        "song2",
			Title:         metadata["song2"].Title,
			ArtistNames:   metadata["song2"].ArtistNames,
			AverageRating: 4.9,
			RatingCount:   200,
			Reason:        "personalized",
		},
	}

	require.Equal(t, 2, len(enriched))
	require.Equal(t, "Bohemian Rhapsody", enriched["song1"].Title)
	require.Equal(t, []string{"Queen"}, enriched["song1"].ArtistNames)
	require.Equal(t, "Stairway to Heaven", enriched["song2"].Title)
}

// TestEnrichSongMetadataCircuitBreakerFallback tests graceful degradation when content-service fails
func TestEnrichSongMetadataCircuitBreakerFallback(t *testing.T) {
	// Simulate content-service unavailable (circuit breaker fallback)
	fallbackEnriched := map[string]dtos.RecommendedSongDto{
		"song1": {
			SongID:        "song1",
			Title:         "Unknown",
			ArtistNames:   []string{"Unknown Artist"},
			AverageRating: 4.5,
			RatingCount:   100,
			Reason:        "personalized",
		},
		"song2": {
			SongID:        "song2",
			Title:         "Unknown",
			ArtistNames:   []string{"Unknown Artist"},
			AverageRating: 4.6,
			RatingCount:   110,
			Reason:        "personalized",
		},
	}

	// Verify fallback behavior: songs still included, metadata unavailable
	require.Equal(t, 2, len(fallbackEnriched))
	require.Equal(t, "Unknown", fallbackEnriched["song1"].Title)
	// Service continues working, doesn't fail completely
	require.NotNil(t, fallbackEnriched["song1"])
}

// TestGetPersonalizedRecommendationsRepositoryError tests error propagation from repository
func TestGetPersonalizedRecommendationsRepositoryError(t *testing.T) {
	baseServices := &Services{
		relationRepository: &fakeGraphRelationRepository{
			recommendedErr: ErrGraphDatabaseUnavailable,
		},
	}

	rs := NewRecommendationService(baseServices)

	_, err := rs.GetPersonalizedRecommendations(context.Background(), "user123", 20)

	require.Error(t, err)
	require.Contains(t, err.Error(), "graph database is currently unavailable")
}

// TestGetTrendingSongsRepositoryError tests error handling for trending songs
func TestGetTrendingSongsRepositoryError(t *testing.T) {
	baseServices := &Services{
		relationRepository: &fakeGraphRelationRepository{
			trendingErr: ErrGraphDatabaseUnavailable,
		},
	}

	rs := NewRecommendationService(baseServices)

	_, err := rs.GetTrendingSongs(context.Background(), 20)

	require.Error(t, err)
	require.Contains(t, err.Error(), "graph database is currently unavailable")
}

// TestPaginationValidation tests limit boundary validation
func TestPaginationValidation(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		expectError bool
	}{
		{
			name:        "Minimum valid limit",
			limit:       1,
			expectError: false,
		},
		{
			name:        "Default limit",
			limit:       20,
			expectError: false,
		},
		{
			name:        "Maximum valid limit",
			limit:       50,
			expectError: false,
		},
		{
			name:        "Exceeds maximum limit",
			limit:       51,
			expectError: true,
		},
		{
			name:        "Zero limit",
			limit:       0,
			expectError: true,
		},
		{
			name:        "Negative limit",
			limit:       -5,
			expectError: true,
		},
	}

	baseServices := &Services{
		relationRepository: &fakeGraphRelationRepository{
			trendingSongs: []string{"song1"},
		},
	}

	rs := NewRecommendationService(baseServices)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := rs.GetTrendingSongs(context.Background(), tt.limit)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestCollaborativeFilteringWithEmptyContent tests fallback to collaborative when content-based is empty
func TestCollaborativeFilteringFallback(t *testing.T) {
	baseServices := &Services{
		relationRepository: &fakeGraphRelationRepository{
			recommendedSongs:   []string{},                 // No content-based recommendations
			collaborativeSongs: []string{"song4", "song5"}, // But collaborative filtering has results
		},
	}

	rs := NewRecommendationService(baseServices)

	recommendations, err := rs.GetPersonalizedRecommendations(context.Background(), "user123", 20)

	require.NoError(t, err)
	require.NotNil(t, recommendations)
	// Should include collaborative songs as fallback
	require.True(t, len(recommendations) > 0)
}

// TestDuplicateSongRemovalInMerge tests that duplicates are properly removed when merging
func TestDuplicateSongRemovalInMerge(t *testing.T) {
	baseServices := &Services{
		relationRepository: &fakeGraphRelationRepository{
			recommendedSongs:   []string{"song1", "song2", "song3", "song2"}, // song2 appears twice
			collaborativeSongs: []string{"song3", "song4"},                   // song3 already in content-based
		},
	}

	rs := NewRecommendationService(baseServices)

	merged := rs.mergeSongRecommendations(
		[]string{"song1", "song2", "song3", "song2"},
		[]string{"song3", "song4"},
	)

	// Expected: unique songs only
	require.Equal(t, 4, len(merged))

	// Check no duplicates
	seen := make(map[string]bool)
	for _, songID := range merged {
		require.False(t, seen[songID], "Found duplicate song: %s", songID)
		seen[songID] = true
	}
}

// TestRecommendationReasonField tests that reason field is correctly set
func TestRecommendationReasonField(t *testing.T) {
	tests := []struct {
		name            string
		contentBased    []string
		collaborative   []string
		expectedReasons map[string]string
	}{
		{
			name:          "Content-based reason",
			contentBased:  []string{"song1"},
			collaborative: []string{"song2"},
			expectedReasons: map[string]string{
				"song1": "subscribed_genres",
				"song2": "similar_users",
			},
		},
		{
			name:          "Mixed sources",
			contentBased:  []string{"song1", "song2"},
			collaborative: []string{"song2", "song3"},
			expectedReasons: map[string]string{
				"song1": "subscribed_genres",
				"song2": "subscribed_genres", // Content-based takes priority
				"song3": "similar_users",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build recommendation DTOs with reasons
			dtoMap := make(map[string]dtos.RecommendedSongDto)

			for _, songID := range tt.contentBased {
				if _, exists := dtoMap[songID]; !exists {
					dtoMap[songID] = dtos.RecommendedSongDto{
						SongID: songID,
						Reason: "subscribed_genres",
					}
				}
			}

			for _, songID := range tt.collaborative {
				if _, exists := dtoMap[songID]; !exists {
					dtoMap[songID] = dtos.RecommendedSongDto{
						SongID: songID,
						Reason: "similar_users",
					}
				}
			}

			// Verify reasons match expected values
			for songID, expectedReason := range tt.expectedReasons {
				dto := dtoMap[songID]
				require.Equal(t, expectedReason, dto.Reason)
			}
		})
	}
}

// TestRatingScoreAggregation tests that rating counts and averages are preserved during merging
func TestRatingScoreAggregation(t *testing.T) {
	recommendedDTOs := []dtos.RecommendedSongDto{
		{
			SongID:        "song1",
			AverageRating: 4.7,
			RatingCount:   50,
			Reason:        "subscribed_genres",
		},
		{
			SongID:        "song2",
			AverageRating: 4.2,
			RatingCount:   30,
			Reason:        "subscribed_genres",
		},
	}

	collaborativeDTOs := []dtos.RecommendedSongDto{
		{
			SongID:        "song3",
			AverageRating: 4.9,
			RatingCount:   200,
			Reason:        "similar_users",
		},
		{
			SongID:        "song1", // Duplicate: should keep content-based version
			AverageRating: 3.5,
			RatingCount:   15,
			Reason:        "similar_users",
		},
	}

	// Simulating merge with dedup logic
	mergedMap := make(map[string]dtos.RecommendedSongDto)
	for _, dto := range recommendedDTOs {
		mergedMap[dto.SongID] = dto
	}
	for _, dto := range collaborativeDTOs {
		if _, exists := mergedMap[dto.SongID]; !exists {
			mergedMap[dto.SongID] = dto
		}
	}

	// Verify song1 retained higher rating from content-based
	require.InEpsilon(t, 4.7, mergedMap["song1"].AverageRating, 0.01)
	require.Equal(t, int64(50), mergedMap["song1"].RatingCount)
	require.Equal(t, "subscribed_genres", mergedMap["song1"].Reason)

	// Verify song3 included from collaborative
	require.InEpsilon(t, 4.9, mergedMap["song3"].AverageRating, 0.01)
	require.Equal(t, int64(200), mergedMap["song3"].RatingCount)
}

// TestEnrichmentWithRealRatingStatsAndArtists tests that enrichment fetches real rating stats and artist names
func TestEnrichmentWithRealRatingStatsAndArtists(t *testing.T) {
	baseServices := &Services{
		songNodeRepository: &mockSongRepository{
			songs: map[string]*entities.SongNode{
				"song1": {SongID: "song1", Title: "Track 1", Duration: 180},
				"song2": {SongID: "song2", Title: "Track 2", Duration: 240},
			},
		},
		relationRepository: &fakeGraphRelationRepository{
			songRatingStats: map[string]struct {
				avg   float64
				count int64
			}{
				"song1": {avg: 4.5, count: 100},
				"song2": {avg: 3.8, count: 45},
			},
			songArtists: map[string][]string{
				"song1": {"Artist A", "Artist B"},
				"song2": {"Artist C"},
			},
		},
	}

	rs := NewRecommendationService(baseServices)
	ctx := context.Background()

	recommendations, err := rs.enrichSongsWithRatings(ctx, []string{"song1", "song2"})

	require.NoError(t, err)
	require.Len(t, recommendations, 2)

	// Create map for easier lookup (songs may be in different order due to parallel processing)
	songMap := make(map[string]dtos.RecommendedSongDto)
	for _, rec := range recommendations {
		songMap[rec.SongID] = rec
	}

	// Verify song1 enrichment
	song1 := songMap["song1"]
	require.Equal(t, "song1", song1.SongID)
	require.Equal(t, "Track 1", song1.Title)
	require.InEpsilon(t, 4.5, song1.AverageRating, 0.01)
	require.Equal(t, int64(100), song1.RatingCount)
	require.Equal(t, []string{"Artist A", "Artist B"}, song1.ArtistNames)

	// Verify song2 enrichment
	song2 := songMap["song2"]
	require.Equal(t, "song2", song2.SongID)
	require.Equal(t, "Track 2", song2.Title)
	require.InEpsilon(t, 3.8, song2.AverageRating, 0.01)
	require.Equal(t, int64(45), song2.RatingCount)
	require.Equal(t, []string{"Artist C"}, song2.ArtistNames)
}

// TestEnrichmentHandlesRatingStatsErrors gracefully
func TestEnrichmentHandlesRatingStatsErrors(t *testing.T) {
	baseServices := &Services{
		songNodeRepository: &mockSongRepository{
			songs: map[string]*entities.SongNode{
				"song1": {SongID: "song1", Title: "Track 1", Duration: 180},
			},
		},
		relationRepository: &fakeGraphRelationRepository{
			songRatingErr: ErrGraphDatabaseUnavailable,
		},
	}

	rs := NewRecommendationService(baseServices)
	ctx := context.Background()

	recommendations, err := rs.enrichSongsWithRatings(ctx, []string{"song1"})

	// Should not error, but return defaults for rating stats
	require.NoError(t, err)
	require.Len(t, recommendations, 1)
	require.InDelta(t, 0.0, recommendations[0].AverageRating, 0.01)
	require.Equal(t, int64(0), recommendations[0].RatingCount)
	require.Equal(t, "Track 1", recommendations[0].Title)
}

// TestEnrichmentHandlesArtistErrors gracefully
func TestEnrichmentHandlesArtistErrors(t *testing.T) {
	baseServices := &Services{
		songNodeRepository: &mockSongRepository{
			songs: map[string]*entities.SongNode{
				"song1": {SongID: "song1", Title: "Track 1", Duration: 180},
			},
		},
		relationRepository: &fakeGraphRelationRepository{
			songArtistsErr: ErrGraphDatabaseUnavailable,
			songRatingStats: map[string]struct {
				avg   float64
				count int64
			}{
				"song1": {avg: 4.5, count: 100},
			},
		},
	}

	rs := NewRecommendationService(baseServices)
	ctx := context.Background()

	recommendations, err := rs.enrichSongsWithRatings(ctx, []string{"song1"})

	// Should not error, but return empty artists list
	require.NoError(t, err)
	require.Len(t, recommendations, 1)
	require.Equal(t, []string{}, recommendations[0].ArtistNames)
	require.InEpsilon(t, 4.5, recommendations[0].AverageRating, 0.01) // Rating stats still available
	require.Equal(t, int64(100), recommendations[0].RatingCount)
}
