package dtos

// UserAnalyticsResponseDto represents the analytics summary for a user.
// This mirrors the spec-required metrics for analytics (requirement 1.16).
type UserAnalyticsResponseDto struct {
	UserID                 string         `json:"user_id"`
	TotalSongsPlayed       int            `json:"total_songs_played"`
	AverageRating          float64        `json:"average_rating"`
	SongsByGenre           map[string]int `json:"songs_by_genre"`
	TopArtists             []TopArtistDto `json:"top_artists"`
	SubscribedArtistsCount int            `json:"subscribed_artists_count"`
}

// TopArtistDto represents an artist and play count for top artists analytics.
type TopArtistDto struct {
	ArtistID  string `json:"artist_id"`
	PlayCount int    `json:"play_count"`
}
