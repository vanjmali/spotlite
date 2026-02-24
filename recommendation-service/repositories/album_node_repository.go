package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

// AlbumNodeRepository provides data access for album nodes in the graph.
type AlbumNodeRepository struct {
	Driver neo4j.DriverWithContext
}

// NewAlbumNodeRepository constructs an AlbumNodeRepository.
func NewAlbumNodeRepository(driver neo4j.DriverWithContext) *AlbumNodeRepository {
	return &AlbumNodeRepository{Driver: driver}
}

// Create creates an album node in the graph.
func (r *AlbumNodeRepository) Create(ctx context.Context, album entities.AlbumNode) error {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(
			ctx,
			`MERGE (a:Album {album_id: $album_id}) SET a.title = $title`,
			map[string]interface{}{
				"album_id": album.AlbumID,
				"title":    album.Title,
			},
		)
	})
	return err
}

// Get retrieves an album node by album ID.
func (r *AlbumNodeRepository) Get(ctx context.Context, albumID string) (*entities.AlbumNode, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (a:Album {album_id: $album_id}) RETURN a.album_id, a.title`,
			map[string]interface{}{"album_id": albumID},
		)
		if err != nil {
			return nil, err
		}

		if res.Next(ctx) {
			record := res.Record()
			return &entities.AlbumNode{
				AlbumID: record.Values[0].(string),
				Title:   record.Values[1].(string),
			}, nil
		}

		return nil, ErrNotFound
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.AlbumNode), nil
}

// Exists checks if an album node exists.
func (r *AlbumNodeRepository) Exists(ctx context.Context, albumID string) (bool, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (a:Album {album_id: $album_id}) RETURN count(a) > 0 AS exists`,
			map[string]interface{}{"album_id": albumID},
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
