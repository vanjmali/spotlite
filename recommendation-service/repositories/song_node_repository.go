package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

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
			`MERGE (s:Song {song_id: $song_id}) SET s.title = $title, s.duration = $duration`,
			map[string]interface{}{
				"song_id":  song.SongID,
				"title":    song.Title,
				"duration": song.Duration,
			},
		)
	})
	return err
}

// Get retrieves a song node by song ID.
func (r *SongNodeRepository) Get(ctx context.Context, songID string) (*entities.SongNode, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (s:Song {song_id: $song_id}) RETURN s.song_id, s.title, s.duration`,
			map[string]interface{}{"song_id": songID},
		)
		if err != nil {
			return nil, err
		}

		if res.Next(ctx) {
			record := res.Record()
			return &entities.SongNode{
				SongID:   record.Values[0].(string),
				Title:    record.Values[1].(string),
				Duration: int(record.Values[2].(int64)),
			}, nil
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return result.(*entities.SongNode), nil
}

// Exists checks if a song node exists.
func (r *SongNodeRepository) Exists(ctx context.Context, songID string) (bool, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (s:Song {song_id: $song_id}) RETURN count(s) > 0 AS exists`,
			map[string]interface{}{"song_id": songID},
		)
		if err != nil {
			return false, err
		}

		if res.Next(ctx) {
			record := res.Record()
			return record.Values[0].(bool), nil
		}

		return false, nil
	})

	if err != nil {
		return false, err
	}

	return result.(bool), nil
}
