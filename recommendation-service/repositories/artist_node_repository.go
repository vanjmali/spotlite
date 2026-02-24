package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

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
			`MERGE (a:Artist {artist_id: $artist_id}) SET a.name = $name`,
			map[string]interface{}{
				"artist_id": artist.ArtistID,
				"name":      artist.Name,
			},
		)
	})
	return err
}

// Get retrieves an artist node by artist ID.
func (r *ArtistNodeRepository) Get(ctx context.Context, artistID string) (*entities.ArtistNode, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (a:Artist {artist_id: $artist_id}) RETURN a.artist_id, a.name`,
			map[string]interface{}{"artist_id": artistID},
		)
		if err != nil {
			return nil, err
		}

		if res.Next(ctx) {
			record := res.Record()
			return &entities.ArtistNode{
				ArtistID: record.Values[0].(string),
				Name:     record.Values[1].(string),
			}, nil
		}

		return nil, ErrNotFound
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.ArtistNode), nil
}}

// Exists checks if an artist node exists.
func (r *ArtistNodeRepository) Exists(ctx context.Context, artistID string) (bool, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (a:Artist {artist_id: $artist_id}) RETURN count(a) > 0 AS exists`,
			map[string]interface{}{"artist_id": artistID},
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
