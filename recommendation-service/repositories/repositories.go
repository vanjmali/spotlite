package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

// UserNodeRepository provides data access for user nodes in the graph.
type UserNodeRepository struct {
	Driver neo4j.DriverWithContext
}

// NewUserNodeRepository constructs a UserNodeRepository.
func NewUserNodeRepository(driver neo4j.DriverWithContext) *UserNodeRepository {
	return &UserNodeRepository{Driver: driver}
}

// Create creates a user node in the graph.
func (r *UserNodeRepository) Create(ctx context.Context, user entities.UserNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(
			ctx,
			`CREATE (u:User {user_id: $user_id, username: $username})`,
			map[string]interface{}{
				"user_id":  user.UserID,
				"username": user.Username,
			},
		)
	})
	return err
}

// SongNodeRepository provides data access for song nodes in the graph.
type SongNodeRepository struct {
	Driver neo4j.DriverWithContext
}

// NewSongNodeRepository constructs a SongNodeRepository.
func NewSongNodeRepository(driver neo4j.DriverWithContext) *SongNodeRepository {
	return &SongNodeRepository{Driver: driver}
}

// Create creates a song node in the graph.
func (r *SongNodeRepository) Create(ctx context.Context, song entities.SongNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(
			ctx,
			`CREATE (s:Song {song_id: $song_id, title: $title, duration: $duration})`,
			map[string]interface{}{
				"song_id":  song.SongID,
				"title":    song.Title,
				"duration": song.Duration,
			},
		)
	})
	return err
}

// ArtistNodeRepository provides data access for artist nodes in the graph.
type ArtistNodeRepository struct {
	Driver neo4j.DriverWithContext
}

// NewArtistNodeRepository constructs an ArtistNodeRepository.
func NewArtistNodeRepository(driver neo4j.DriverWithContext) *ArtistNodeRepository {
	return &ArtistNodeRepository{Driver: driver}
}

// Create creates an artist node in the graph.
func (r *ArtistNodeRepository) Create(ctx context.Context, artist entities.ArtistNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(
			ctx,
			`CREATE (a:Artist {artist_id: $artist_id, name: $name})`,
			map[string]interface{}{
				"artist_id": artist.ArtistID,
				"name":      artist.Name,
			},
		)
	})
	return err
}

// GenreNodeRepository provides data access for genre nodes in the graph.
type GenreNodeRepository struct {
	Driver neo4j.DriverWithContext
}

// NewGenreNodeRepository constructs a GenreNodeRepository.
func NewGenreNodeRepository(driver neo4j.DriverWithContext) *GenreNodeRepository {
	return &GenreNodeRepository{Driver: driver}
}

// Create creates a genre node in the graph.
func (r *GenreNodeRepository) Create(ctx context.Context, genre entities.GenreNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(
			ctx,
			`CREATE (g:Genre {genre_id: $genre_id, name: $name})`,
			map[string]interface{}{
				"genre_id": genre.GenreID,
				"name":     genre.Name,
			},
		)
	})
	return err
}
