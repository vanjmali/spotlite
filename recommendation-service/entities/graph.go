package entities

// UserNode represents a user node in the recommendation graph.
type UserNode struct {
	UserID   string
	Username string
}

// SongNode represents a song node in the recommendation graph.
type SongNode struct {
	SongID   string
	Title    string
	Duration int
	GenreIDs []string
	Artists  []string
}

// GenreNode represents a genre node in the recommendation graph.
type GenreNode struct {
	GenreID string
	Name    string
}

type GenreSubscription struct {
	GenreID string
	UserID  string
}

type SongRating struct {
	SongID string
	UserID string
	Value  int
}

type SongRecommendation struct {
	SongID   string   `json:"song_d"`
	Title    string   `json:"title"`
	Duration int      `json:"duration"`
	Rating   float64  `json:"rating"`
	Artists  []string `json:"artists"`
}
