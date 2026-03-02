package entities

import (
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CQRS Read Models
//
// This file defines the read models for the analytics service following the CQRS pattern.
// Read models are denormalized projections built from the event log (event sourcing).
//
// CQRS Benefits:
// - Optimized query performance: read models are designed specifically for query patterns
// - Denormalized data: eliminates joins and aggregations at query time
// - Scalability: read and write sides can be scaled independently
// - Flexibility: multiple read models can be built from the same event log
//
// Read models are eventually consistent with the event log. When events are written to
// the event store, event handlers update the corresponding read models asynchronously.
//
// See README.md Event Sourcing + CQRS section and specification requirements:
// - 1.15: Activity history
// - 1.16: Analytics
// - 2.15: Event sourcing + CQRS

// UserAnalyticsReadModel represents the denormalized analytics data for a user.
// This read model is optimized for querying user analytics (requirement 1.16).
// It is updated by projecting events from the event log.
//
// Stored in MongoDB collection: user_analytics
// Updated by: AnalyticsProjectionService when processing events
type UserAnalyticsReadModel struct {
	// ID is the MongoDB document ID
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// UserID is the unique identifier of the user
	UserID string `bson:"user_id" json:"user_id" validate:"required"`

	// TotalSongsPlayed is the total number of songs the user has listened to
	// Updated on: EventTypeSongPlayed
	TotalSongsPlayed int `bson:"total_songs_played" json:"total_songs_played"`

	// AverageRating is the average of all ratings the user has given to songs
	// Calculated from RatingSum / RatingsCount
	// Updated by projection from rating events
	AverageRating float64 `bson:"average_rating" json:"average_rating"`

	// RatingSum is the sum of all ratings given by the user
	// Used to calculate AverageRating incrementally
	// Updated on: EventTypeRatingCreated, EventTypeRatingUpdated, EventTypeRatingDeleted
	RatingSum int `bson:"rating_sum" json:"rating_sum"`

	// RatingsCount is the total number of ratings given by the user
	// Used to calculate AverageRating incrementally
	// Updated on: EventTypeRatingCreated, EventTypeRatingDeleted
	RatingsCount int `bson:"ratings_count" json:"ratings_count"`

	// SongsByGenre tracks the number of songs played per genre
	// Key: genre ID, Value: play count
	// Updated on: EventTypeSongPlayed
	SongsByGenre map[string]int `bson:"songs_by_genre" json:"songs_by_genre"`

	// TopArtists contains the top artists the user has listened to
	// Sorted by play count (descending), limited to top 5
	// Updated on: EventTypeSongPlayed
	TopArtists []ArtistPlayCount `bson:"top_artists" json:"top_artists"`

	// SubscribedArtistsCount is the number of artists the user is subscribed to
	// Updated on: EventTypeSubscriptionCreated, EventTypeSubscriptionDeleted
	SubscribedArtistsCount int `bson:"subscribed_artists_count" json:"subscribed_artists_count"`
}

// ArtistPlayCount represents the play count for a specific artist
// Used in TopArtists slice to track most listened artists
type ArtistPlayCount struct {
	// ArtistID is the unique identifier of the artist
	ArtistID string `bson:"artist_id" json:"artist_id"`

	// PlayCount is how many times the user has listened to songs by this artist
	PlayCount int `bson:"play_count" json:"play_count"`
}

// NewUserAnalyticsReadModel creates a new analytics read model for a user
func NewUserAnalyticsReadModel(userID string) *UserAnalyticsReadModel {
	return &UserAnalyticsReadModel{
		UserID:                 userID,
		TotalSongsPlayed:       0,
		AverageRating:          0.0,
		RatingSum:              0,
		RatingsCount:           0,
		SongsByGenre:           make(map[string]int),
		TopArtists:             []ArtistPlayCount{},
		SubscribedArtistsCount: 0,
	}
}

// AddSongPlayed updates analytics when a song is played
// Updates: TotalSongsPlayed, SongsByGenre, TopArtists
func (u *UserAnalyticsReadModel) AddSongPlayed(genreID, artistID string) {
	u.TotalSongsPlayed++

	// Update genre play count
	if genreID != "" {
		u.SongsByGenre[genreID]++
	}

	// Update artist play count
	if artistID != "" {
		updated := false
		for i := range u.TopArtists {
			if u.TopArtists[i].ArtistID == artistID {
				u.TopArtists[i].PlayCount++
				updated = true
				break
			}
		}
		if !updated {
			u.TopArtists = append(u.TopArtists, ArtistPlayCount{
				ArtistID:  artistID,
				PlayCount: 1,
			})
		}

		// Sort and keep top 5 artists
		u.sortAndLimitTopArtists()
	}

}

// AddSubscription updates subscription counts
func (u *UserAnalyticsReadModel) AddSubscription(subscriptionType subscription.SubscriptionType) {
	if subscriptionType == subscription.ArtistSubscription {
		u.SubscribedArtistsCount++
	}
}

// DeleteSubscription updates subscription counts
func (u *UserAnalyticsReadModel) DeleteSubscription(subscriptionType subscription.SubscriptionType) {
	if subscriptionType == subscription.ArtistSubscription && u.SubscribedArtistsCount > 0 {
		u.SubscribedArtistsCount--
	}
}

// AddRating updates rating statistics when a new rating is created
// Updates: RatingSum, RatingsCount, AverageRating
func (u *UserAnalyticsReadModel) AddRating(rating int) {
	u.RatingSum += rating
	u.RatingsCount++
	u.calculateAverageRating()
}

// UpdateRating updates rating statistics when an existing rating is changed
// Updates: RatingSum, AverageRating (RatingsCount stays the same)
func (u *UserAnalyticsReadModel) UpdateRating(oldRating int, newRating int) {
	u.RatingSum = u.RatingSum - oldRating + newRating
	u.calculateAverageRating()
}

// DeleteRating updates rating statistics when a rating is removed
// Updates: RatingSum, RatingsCount, AverageRating
func (u *UserAnalyticsReadModel) DeleteRating(rating int) {
	if u.RatingsCount > 0 {
		u.RatingSum -= rating
		u.RatingsCount--
		u.calculateAverageRating()
	}
}

// calculateAverageRating recalculates the average rating from sum and count
func (u *UserAnalyticsReadModel) calculateAverageRating() {
	if u.RatingsCount == 0 {
		u.AverageRating = 0.0
	} else {
		u.AverageRating = float64(u.RatingSum) / float64(u.RatingsCount)
	}
}

// sortAndLimitTopArtists sorts artists by play count and keeps only top 5
func (u *UserAnalyticsReadModel) sortAndLimitTopArtists() {
	// Simple bubble sort (sufficient for small arrays)
	for i := 0; i < len(u.TopArtists); i++ {
		for j := i + 1; j < len(u.TopArtists); j++ {
			if u.TopArtists[j].PlayCount > u.TopArtists[i].PlayCount {
				u.TopArtists[i], u.TopArtists[j] = u.TopArtists[j], u.TopArtists[i]
			}
		}
	}

	// Keep only top 5
	if len(u.TopArtists) > 5 {
		u.TopArtists = u.TopArtists[:5]
	}
}
