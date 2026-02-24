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

// ArtistNode represents an artist node in the recommendation graph.
type ArtistNode struct {
	ArtistID string
	Name     string
}

// GenreNode represents a genre node in the recommendation graph.
type GenreNode struct {
	GenreID string
	Name    string
}

// AlbumNode represents an album node in the recommendation graph.
type AlbumNode struct {
	AlbumID string
	Title   string
}

// Rating represents a RATED relationship in the graph.
type Rating struct {
	UserID string
	SongID string
	Value  int // 1-5
}

// Listened represents a LISTENED relationship in the graph.
type Listened struct {
	UserID string
	SongID string
}

// ArtistSubscription represents a SUBSCRIBED relationship in the graph.
type ArtistSubscription struct {
	UserID   string
	ArtistID string
}

// GenreSubscription represents a SUBSCRIBED_GENRE relationship in the graph.
type GenreSubscription struct {
	UserID  string
	GenreID string
}
