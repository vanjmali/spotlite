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
}

// GenreNode represents a genre node in the recommendation graph.
type GenreNode struct {
	GenreID string
	Name    string
}

// Rating represents a RATED relationship in the graph.
type Rating struct {
	UserID string
	SongID string
	Value  int
}

// GenreSubscription represents a SUBSCRIBED_GENRE relationship in the graph.
type GenreSubscription struct {
	UserID  string
	GenreID string
}
