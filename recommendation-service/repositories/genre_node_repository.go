package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/vanjmali/spotlite/recommendation-service/entities"
)

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
			`MERGE (g:Genre {genre_id: $genre_id}) SET g.name = $name`,
			map[string]interface{}{
				"genre_id": genre.GenreID,
				"name":     genre.Name,
			},
		)
	})
	return err
}

// Get retrieves a genre node by genre ID.
func (r *GenreNodeRepository) Get(ctx context.Context, genreID string) (*entities.GenreNode, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (g:Genre {genre_id: $genre_id}) RETURN g.genre_id, g.name`,
			map[string]interface{}{"genre_id": genreID},
		)
		if err != nil {
			return nil, err
		}

		if res.Next(ctx) {
			record := res.Record()
			return &entities.GenreNode{
				GenreID: record.Values[0].(string),
				Name:    record.Values[1].(string),
			}, nil
		}

		return nil, ErrNotFound
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.GenreNode), nil
}}

// Exists checks if a genre node exists.
func (r *GenreNodeRepository) Exists(ctx context.Context, genreID string) (bool, error) {
	session := r.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(
			ctx,
			`MATCH (g:Genre {genre_id: $genre_id}) RETURN count(g) > 0 AS exists`,
			map[string]interface{}{"genre_id": genreID},
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
