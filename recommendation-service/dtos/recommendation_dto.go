package dtos

// RecommendedSongDto represents a recommended song in the response
type RecommendedSongDto struct {
	SongID        string   `json:"songId"`
	Title         string   `json:"title"`
	ArtistNames   []string `json:"artistNames"`
	AverageRating float64  `json:"averageRating"`
	RatingCount   int64    `json:"ratingCount"`
	Reason        string   `json:"reason"` // e.g., "subscribed_genres", "similar_users"
}

// RecommendationFilterDto represents query parameters for recommendations
type RecommendationFilterDto struct {
	Limit                int  `json:"limit"`
	IncludeCollaborative bool `json:"includeCollaborative"`
}

// RecommendationResponseDto represents the response for recommendation requests
type RecommendationResponseDto struct {
	Songs   []RecommendedSongDto `json:"songs"`
	Message string               `json:"message"`
}

// SongMetadata represents enriched song data from content-service
type SongMetadata struct {
	SongID      string   `json:"songId"`
	Title       string   `json:"title"`
	ArtistNames []string `json:"artistNames"`
	Duration    int      `json:"duration"`
}
